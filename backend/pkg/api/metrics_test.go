package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAppInstancesPerChannelMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	metrics, err := a.GetAppInstancesPerChannelMetrics()
	require.NoError(t, err)
	expectedMetrics := []AppInstancesPerChannelMetric{
		{ApplicationName: "Sample application", Version: "1.0.1", ChannelName: "Failing", InstancesCount: 1},
		{ApplicationName: "Sample application", Version: "1.0.1", ChannelName: "Master", InstancesCount: 1},
		{ApplicationName: "Sample application", Version: "1.0.1", ChannelName: "Stable", InstancesCount: 1},
		{ApplicationName: "Sample application", Version: "1.0.2", ChannelName: "Master", InstancesCount: 1},
		{ApplicationName: "Sample application", Version: "1.0.2", ChannelName: "Stable", InstancesCount: 2},
		{ApplicationName: "Sample application", Version: "1.0.3", ChannelName: "Master", InstancesCount: 1},
		{ApplicationName: "Sample application", Version: "1.0.3", ChannelName: "Stable", InstancesCount: 4},
		{ApplicationName: "Sample application", Version: "1.0.4", ChannelName: "Master", InstancesCount: 1},
	}
	require.Equal(t, expectedMetrics, metrics)
}

func TestGetFailedUpdatesMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	metrics, err := a.GetFailedUpdatesMetrics()
	require.NoError(t, err)
	expectedMetrics := []FailedUpdatesMetric{
		{ApplicationName: "Sample application", FailureCount: 1},
	}
	require.Equal(t, expectedMetrics, metrics)
}

func TestGetInstancesPerOEMMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	metrics, err := a.GetInstancesPerOEMMetrics()
	require.NoError(t, err)

	// Sample data assigns OEM values to all 12 instances of "Sample application":
	//   aws:    instance1,2 (west), instance4,5 (east), instance8,9 (qa)  -> 6
	//   azure:  instance7 (east), instance11 (qa)                         -> 2
	//   qemu:   instance12 (failing)                                      -> 1
	//   vmware: instance3 (west), instance6 (east), instance10 (qa)       -> 3
	expectedMetrics := []InstancesPerOEMMetric{
		{ApplicationName: "Sample application", OEM: "aws", InstancesCount: 6},
		{ApplicationName: "Sample application", OEM: "azure", InstancesCount: 2},
		{ApplicationName: "Sample application", OEM: "qemu", InstancesCount: 1},
		{ApplicationName: "Sample application", OEM: "vmware", InstancesCount: 3},
	}
	require.Equal(t, expectedMetrics, metrics)
}
