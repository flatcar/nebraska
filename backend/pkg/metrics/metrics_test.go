package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
)

func seriesCount(t *testing.T, c prometheus.Collector) int {
	t.Helper()
	ch := make(chan prometheus.Metric, 1024)
	c.Collect(ch)
	close(ch)
	count := 0
	for range ch {
		count++
	}
	return count
}

func TestSetAppInstancesPerChannelDropsStaleSeries(t *testing.T) {
	t.Cleanup(appInstancePerChannelGaugeMetric.Reset)
	appInstancePerChannelGaugeMetric.Reset()

	setAppInstancesPerChannel([]api.AppInstancesPerChannelMetric{
		{ApplicationName: "app", Version: "1.0.0", ChannelName: "stable", InstancesCount: 7},
		{ApplicationName: "app", Version: "2.0.0", ChannelName: "stable", InstancesCount: 3},
	})
	require.Equal(t, 2, seriesCount(t, appInstancePerChannelGaugeMetric))

	setAppInstancesPerChannel([]api.AppInstancesPerChannelMetric{
		{ApplicationName: "app", Version: "2.0.0", ChannelName: "stable", InstancesCount: 10},
	})

	assert.Equal(t, 1, seriesCount(t, appInstancePerChannelGaugeMetric),
		"a version that is no longer reported must not keep its last value")
}

func TestSetFailedUpdatesDropsStaleSeries(t *testing.T) {
	t.Cleanup(failedUpdatesGaugeMetric.Reset)
	failedUpdatesGaugeMetric.Reset()

	setFailedUpdates([]api.FailedUpdatesMetric{
		{ApplicationName: "app-a", FailureCount: 4},
		{ApplicationName: "app-b", FailureCount: 1},
	})
	require.Equal(t, 2, seriesCount(t, failedUpdatesGaugeMetric))

	setFailedUpdates([]api.FailedUpdatesMetric{
		{ApplicationName: "app-a", FailureCount: 4},
	})

	assert.Equal(t, 1, seriesCount(t, failedUpdatesGaugeMetric),
		"an application that stopped reporting failures must not keep its series")
}

func TestSetMetricsEmptyResultClearsEverything(t *testing.T) {
	t.Cleanup(appInstancePerChannelGaugeMetric.Reset)
	t.Cleanup(failedUpdatesGaugeMetric.Reset)
	appInstancePerChannelGaugeMetric.Reset()
	failedUpdatesGaugeMetric.Reset()

	setAppInstancesPerChannel([]api.AppInstancesPerChannelMetric{
		{ApplicationName: "app", Version: "1.0.0", ChannelName: "stable", InstancesCount: 7},
	})
	setFailedUpdates([]api.FailedUpdatesMetric{{ApplicationName: "app", FailureCount: 2}})
	require.Equal(t, 1, seriesCount(t, appInstancePerChannelGaugeMetric))
	require.Equal(t, 1, seriesCount(t, failedUpdatesGaugeMetric))

	setAppInstancesPerChannel(nil)
	setFailedUpdates(nil)

	assert.Zero(t, seriesCount(t, appInstancePerChannelGaugeMetric))
	assert.Zero(t, seriesCount(t, failedUpdatesGaugeMetric))
}
