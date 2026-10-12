package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestGetAppInstancesPerChannelMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// defaultTeamID constant is defined in users_test.go
	metrics, err := a.GetAppInstancesPerChannelMetrics()
	require.NoError(t, err)
	expectedMetrics := []types.AppInstancesPerChannelMetric{
		{
			ApplicationName: "Sample application",
			Version:         "1.0.1",
			ChannelName:     "Failing",
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.1",
			ChannelName:     "Master",
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.1",
			ChannelName:     "Stable",
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.2",
			ChannelName:     "Master",
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.2",
			ChannelName:     "Stable",
			InstancesCount:  2,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.3",
			ChannelName:     "Master",
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.3",
			ChannelName:     "Stable",
			InstancesCount:  4,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.4",
			ChannelName:     "Master",
			InstancesCount:  1,
		},
	}

	require.Equal(t, expectedMetrics, metrics)
}

// TestGetAppInstancesPerChannelMetricsExcludesStaleInstances checks that an
// instance is dropped from the metric once it stops checking in, matching
// the activity window (validityInterval) used by GetApp/GetApps/GetInstances
// and friends. Regression test for
// https://github.com/flatcar/nebraska/issues/1562.
func TestGetAppInstancesPerChannelMetricsExcludesStaleInstances(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Make every sample instance look like it hasn't checked in for 400 days.
	longAgo := time.Now().UTC().AddDate(0, 0, -400)
	_, err := a.db().Exec(`UPDATE instance_application SET last_check_for_updates = $1`, longAgo)
	require.NoError(t, err)

	metrics, err := a.GetAppInstancesPerChannelMetrics()
	require.NoError(t, err)
	require.Empty(t, metrics, "instances that stopped checking in long ago must not be counted")
}

// TestGetAppInstancesPerChannelMetricsIncludesGroupsWithoutChannel checks
// that instances belonging to a group with no channel assigned
// (groups.channel_id can be NULL, e.g. after the channel was deleted) are
// still counted instead of being silently dropped by the join. Regression
// test for https://github.com/flatcar/nebraska/issues/1562.
func TestGetAppInstancesPerChannelMetricsIncludesGroupsWithoutChannel(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// "Prod EC2 us-west-2" (bcaa68bc-...) has 3 active instances pointed at
	// the "Stable" channel: instance1 (1.0.3), instance2 (1.0.3), instance3 (1.0.2).
	_, err := a.db().Exec(`UPDATE groups SET channel_id = NULL WHERE id = 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f'`)
	require.NoError(t, err)

	metrics, err := a.GetAppInstancesPerChannelMetrics()
	require.NoError(t, err)

	var totalInstances int
	var noChannel103, noChannel102 int
	for _, m := range metrics {
		totalInstances += m.InstancesCount
		if m.ChannelName == "" {
			switch m.Version {
			case "1.0.3":
				noChannel103 = m.InstancesCount
			case "1.0.2":
				noChannel102 = m.InstancesCount
			}
		}
	}

	require.Equal(t, 12, totalInstances, "all 12 active sample instances must still be counted, channel or not")
	require.Equal(t, 2, noChannel103, "instance1 and instance2 should fall into the no-channel bucket")
	require.Equal(t, 1, noChannel102, "instance3 should fall into the no-channel bucket")
}

func TestGetFailedUpdatesMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// defaultTeamID constant is defined in users_test.go
	metrics, err := a.GetFailedUpdatesMetrics()
	require.NoError(t, err)
	expectedMetrics := []types.FailedUpdatesMetric{
		{
			ApplicationName: "Sample application",
			FailureCount:    1,
		},
	}
	require.Equal(t, expectedMetrics, metrics)
}
