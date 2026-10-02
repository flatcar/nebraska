package dbreads_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
)

const defaultTestDbURL = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}
	if _, ok := os.LookupEnv("NEBRASKA_DB_URL"); !ok {
		_ = os.Setenv("NEBRASKA_DB_URL", defaultTestDbURL)
	}
	os.Exit(m.Run())
}

func TestGetInstanceStatsByTimestamp(t *testing.T) {
	a, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()

	db := dbconn.DB(a.Conn())

	// 1. Set the session timezone to non-UTC
	_, err = db.Exec("SET timezone = 'America/New_York'")
	require.NoError(t, err)

	// Clean up any existing stats just in case
	_, err = db.Exec("DELETE FROM instance_stats")
	require.NoError(t, err)

	// 2. Seed a row using a fixed timestamp.
	testTime, err := time.Parse(time.RFC3339, "2026-08-16T12:00:00Z")
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO instance_stats (timestamp, channel_name, arch, version, instances)
		VALUES ($1, $2, $3, $4, $5)`,
		testTime, "stable", "amd64", "1.0.0", 5,
	)
	require.NoError(t, err)

	// 3. Assert GetInstanceStatsByTimestamp still finds it
	stats, err := a.Reads().GetInstanceStatsByTimestamp(testTime)
	require.NoError(t, err)

	require.Len(t, stats, 1)
	assert.Equal(t, "1.0.0", stats[0].Version)
	assert.Equal(t, 5, stats[0].Instances)
}
