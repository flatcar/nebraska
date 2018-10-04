package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAppInstancesPerChannelMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// defaultTeamID constant is defined in users_test.go
	metrics, err := a.GetAppInstancesPerChannelMetrics(defaultTeamID)
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
	metrics, err := a.GetFailedUpdatesMetrics(defaultTeamID)
	require.NoError(t, err)
	expectedMetrics := []FailedUpdatesMetric{
		{
			ApplicationName: "Sample application",
			FailureCount:    1,
		},
	}
	require.Equal(t, expectedMetrics, metrics)
}
