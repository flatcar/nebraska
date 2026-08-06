package dbreads_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
)

func TestGetInstanceStatsLatest(t *testing.T) {
	a, err := api.NewForTest(api.OptionInitDB, api.OptionDisableUpdatesOnFailedRollout)
	require.NoError(t, err)
	defer a.Close()

	db := dbreads.DB(a.Queries)

	// 1. Seed two different timestamps
	ts1 := time.Now().UTC().Add(-1 * time.Hour)
	ts2 := time.Now().UTC()

	_, err = db.Exec(`INSERT INTO instance_stats (timestamp, channel_name, arch, version, instances) VALUES ($1, $2, $3, $4, $5)`, ts1, "channel", "AMD64", "1.0.0", 1)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO instance_stats (timestamp, channel_name, arch, version, instances) VALUES ($1, $2, $3, $4, $5)`, ts2, "channel", "AMD64", "1.0.1", 2)
	require.NoError(t, err)

	// Get latest stats
	instanceStats, err := a.GetInstanceStatsLatest()
	assert.NoError(t, err)
	require.NotEmpty(t, instanceStats, "GetInstanceStatsLatest returned 0 items. This is likely due to the known timezone offset dropping bug in GetInstanceStatsByTimestamp's timestamp cast.")

	assert.Equal(t, 1, len(instanceStats))
	for _, stat := range instanceStats {
		maxTs, err := a.GetLatestInstanceStatsTimestamp()
		assert.NoError(t, err)
		
		// Postgres might truncate microseconds, check with diff <= 1 sec
		assert.WithinDuration(t, maxTs, stat.Timestamp, time.Second)
		assert.WithinDuration(t, ts2, maxTs, time.Second)
		assert.Equal(t, "1.0.1", stat.Version)
		assert.Equal(t, 2, stat.Instances)
	}
}
