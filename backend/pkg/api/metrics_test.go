package api

import (
	"testing"

	"github.com/google/uuid"
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

func TestGetAppInstancesByOEMMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "oem_metrics_team"})
	tApp, _ := as.AddApp(&Application{Name: "oem_app", TeamID: tTeam.ID})
	tGroup, _ := as.AddGroup(&Group{Name: "oem_group", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	// Sample data instances all predate migration 0021 and carry an empty OEM,
	// so the metric must be empty until instances with a known OEM are registered.
	metrics, err := a.GetAppInstancesByOEMMetrics()
	require.NoError(t, err)
	require.Empty(t, metrics)

	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.10", OEM: "aws"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.11", OEM: "aws"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.12", OEM: "azure"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.13"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)

	metrics, err = a.GetAppInstancesByOEMMetrics()
	require.NoError(t, err)
	expectedMetrics := []AppInstancesByOEMMetric{
		{
			ApplicationName: "oem_app",
			OEM:             "aws",
			InstancesCount:  2,
		},
		{
			ApplicationName: "oem_app",
			OEM:             "azure",
			InstancesCount:  1,
		},
	}
	require.Equal(t, expectedMetrics, metrics)
}
