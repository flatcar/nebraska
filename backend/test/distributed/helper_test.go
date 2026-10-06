package distributed_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// A wedged node should name itself in the failure rather than block until Go's
// test binary panics.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// waitNodeReady polls until the node serves. docker compose returns when a
// container has started, which is well before Nebraska has migrated.
func waitNodeReady(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		err := probeHealth(url)
		if err == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return err
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func probeHealth(url string) error {
	resp, err := httpClient.Get(url + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "OK" {
		return fmt.Errorf("status %d, body %q", resp.StatusCode, body)
	}

	return nil
}

func singleURL() string  { return os.Getenv("NEBRASKA_TEST_SINGLE_URL") }
func controlURL() string { return os.Getenv("NEBRASKA_TEST_CONTROL_URL") }
func edgeURL() string    { return os.Getenv("NEBRASKA_TEST_EDGE_URL") }

func controlDBURL() string { return os.Getenv("NEBRASKA_TEST_CONTROL_DB_URL") }
func edgeDBURL() string    { return os.Getenv("NEBRASKA_TEST_EDGE_DB_URL") }

// setUpdatesOverride writes group_local directly. Nothing in the API sets this
// column without a timed-out update under safe mode, which no E2E test can wait
// for.
func setUpdatesOverride(t *testing.T, dbURL, groupID string, value any) {
	t.Helper()

	db, err := sqlx.Open("pgx", dbURL)
	require.NoError(t, err)

	defer db.Close()

	result, err := db.Exec("update group_local set policy_updates_enabled_override = $1 where group_id = $2", value, groupID)
	require.NoError(t, err)

	rows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rows, "no group_local row for %s", groupID)
}

// setupReplication subscribes the edge database to the control database's admin
// tables. Both nodes migrate to the same seeded rows, so nothing is copied.
func setupReplication(controlDSN, edgeDSN string) error {
	control, err := sqlx.Open("pgx", controlDSN)
	if err != nil {
		return err
	}
	defer control.Close()

	edge, err := sqlx.Open("pgx", edgeDSN)
	if err != nil {
		return err
	}
	defer edge.Close()

	var subscribed bool
	if err := edge.Get(&subscribed, "select exists (select 1 from pg_subscription where subname = $1)", replicationName); err != nil {
		return err
	}
	if subscribed {
		return nil
	}

	// The admin tables are the ones grants.sql lets only the admin role write.
	var tables []string
	err = control.Select(&tables, `
		select format('public.%I', c.relname)
		from pg_class c join pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relkind = 'r'
		  and has_table_privilege('nebraska_admin_' || current_database(), c.oid, 'insert')
		  and not has_table_privilege('nebraska_runtime_' || current_database(), c.oid, 'insert')
		order by 1`)
	if err != nil {
		return err
	}
	if len(tables) == 0 {
		return fmt.Errorf("no admin tables found to publish")
	}

	var controlDB string
	if err := control.Get(&controlDB, "select current_database()"); err != nil {
		return err
	}

	if _, err := control.Exec(fmt.Sprintf("create publication %s for table %s", replicationName, strings.Join(tables, ", "))); err != nil {
		return err
	}

	// A subscription cannot create its own slot on the cluster it runs in.
	if _, err := control.Exec("select pg_create_logical_replication_slot($1, 'pgoutput')", replicationName); err != nil {
		return err
	}

	_, err = edge.Exec(fmt.Sprintf(
		"create subscription %[1]s connection 'host=localhost dbname=%[2]s user=postgres password=nebraska' publication %[1]s with (create_slot = false, slot_name = %[1]s, copy_data = false)",
		replicationName, controlDB))

	return err
}

func openDB(t *testing.T, url string) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("pgx", url)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return db
}

func count(t *testing.T, db *sqlx.DB, query string, args ...any) int {
	t.Helper()

	var n int
	require.NoError(t, db.Get(&n, query, args...))

	return n
}

// waitCount polls, because a change on the control node reaches the edge
// asynchronously.
func waitCount(t *testing.T, db *sqlx.DB, want int, query string, args ...any) {
	t.Helper()

	deadline := time.Now().Add(replicationTimeout)

	for {
		got := count(t, db, query, args...)
		if got == want || time.Now().After(deadline) {
			require.Equal(t, want, got, query)
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// createID posts payload and returns the id of the created resource.
func createID(t *testing.T, url, payload string) string {
	t.Helper()

	code, body := request(t, "POST", url, strings.NewReader(payload))
	require.Equal(t, http.StatusOK, code, body)

	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &created))
	require.NotEmpty(t, created.ID, body)

	return created.ID
}

// request returns the status and the raw body, so a test can assert on an error
// payload and not only on the status code.
func request(t *testing.T, method, url string, payload io.Reader) (int, string) {
	t.Helper()

	req, err := http.NewRequest(method, url, payload)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, strings.TrimSpace(string(body))
}
