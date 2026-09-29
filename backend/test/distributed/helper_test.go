package distributed_test

import (
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
