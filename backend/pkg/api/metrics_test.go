package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAppInstancesPerChannelMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// defaultTeamID constant is defined in users_test.go
	metrics, err := a.GetAppInstancesPerChannelMetrics()
	require.NoError(t, err)
	expectedMetrics := []AppInstancesPerChannelMetric{
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

func TestGetFailedUpdatesMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// defaultTeamID constant is defined in users_test.go
	metrics, err := a.GetFailedUpdatesMetrics()
	require.NoError(t, err)
	expectedMetrics := []FailedUpdatesMetric{
		{
			ApplicationName: "Sample application",
			FailureCount:    1,
		},
	}
	require.Equal(t, expectedMetrics, metrics)
}

func TestGetGroupUpdatesStatsMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Get metrics for all groups with updates enabled
	metrics, err := a.GetGroupUpdatesStatsMetrics()
	require.NoError(t, err)

	// Should have metrics for at least one group (sample data includes groups with updates enabled)
	require.NotEmpty(t, metrics, "should have at least one group with updates enabled")

	// Validate structure of returned metrics
	for _, metric := range metrics {
		require.NotEmpty(t, metric.ApplicationID, "application_id should not be empty")
		require.NotEmpty(t, metric.GroupID, "group_id should not be empty")
		require.NotEmpty(t, metric.ApplicationName, "application_name should not be empty")
		require.NotEmpty(t, metric.GroupName, "group_name should not be empty")
		require.NotEmpty(t, metric.ChannelName, "channel_name should not be empty")

		// Validate that counts are non-negative
		require.GreaterOrEqual(t, metric.TotalInstances, 0, "total_instances should be non-negative")
		require.GreaterOrEqual(t, metric.UpdatesToCurrentVersionGranted, 0, "updates_granted should be non-negative")
		require.GreaterOrEqual(t, metric.UpdatesToCurrentVersionAttempted, 0, "updates_attempted should be non-negative")
		require.GreaterOrEqual(t, metric.UpdatesToCurrentVersionSucceeded, 0, "updates_succeeded should be non-negative")
		require.GreaterOrEqual(t, metric.UpdatesToCurrentVersionFailed, 0, "updates_failed should be non-negative")
		require.GreaterOrEqual(t, metric.UpdatesInProgress, 0, "updates_in_progress should be non-negative")
		require.GreaterOrEqual(t, metric.UpdatesTimedOut, 0, "updates_timed_out should be non-negative")
		require.GreaterOrEqual(t, metric.UpdatesGrantedInLastPeriod, 0, "updates_granted_in_period should be non-negative")

		// Logical validations
		// Attempted should be <= Granted
		require.LessOrEqual(t, metric.UpdatesToCurrentVersionAttempted, metric.UpdatesToCurrentVersionGranted,
			"attempted updates should be <= granted updates")

		// Succeeded + Failed should be <= Attempted
		successPlusFailed := metric.UpdatesToCurrentVersionSucceeded + metric.UpdatesToCurrentVersionFailed
		require.LessOrEqual(t, successPlusFailed, metric.UpdatesToCurrentVersionAttempted,
			"succeeded + failed should be <= attempted")
	}
}

