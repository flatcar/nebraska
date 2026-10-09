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
func singleDBURL() string  { return os.Getenv("NEBRASKA_TEST_SINGLE_DB_URL") }

// createGroup adds a group with updates enabled to the seeded application.
func createGroup(t *testing.T, base string) string {
	t.Helper()

	name := fmt.Sprintf("g%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"name":%q,"policy_updates_enabled":true,"policy_timezone":"UTC","policy_period_interval":"15 minutes","policy_max_updates_per_period":2,"policy_update_timeout":"60 minutes"}`, name)
	groups := fmt.Sprintf("%s/api/apps/%s/groups", base, seededAppID)

	code, body := request(t, "POST", groups, strings.NewReader(payload))
	require.Equal(t, http.StatusOK, code, body)

	var group struct{ ID string }
	require.NoError(t, json.Unmarshal([]byte(body), &group))

	t.Cleanup(func() { request(t, "DELETE", groups+"/"+group.ID, nil) })

	return group.ID
}

// editDialogSave builds the body GroupEditDialog sends on save: every field it
// read, with only the description changed.
func editDialogSave(t *testing.T, group string) string {
	t.Helper()

	var read map[string]any
	require.NoError(t, json.Unmarshal([]byte(group), &read))

	save := map[string]any{"description": "edited"}
	for _, field := range []string{"id", "application_id", "name", "track", "policy_updates_enabled", "policy_safe_mode",
		"policy_office_hours", "policy_timezone", "policy_period_interval", "policy_max_updates_per_period", "policy_update_timeout"} {
		save[field] = read[field]
	}

	body, err := json.Marshal(save)
	require.NoError(t, err)

	return string(body)
}

func adminUpdatesEnabled(t *testing.T, dbURL, groupID string) bool {
	t.Helper()

	db, err := sqlx.Open("pgx", dbURL)
	require.NoError(t, err)

	defer db.Close()

	var enabled bool
	require.NoError(t, db.Get(&enabled, "select policy_updates_enabled from groups where id = $1", groupID))

	return enabled
}

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
