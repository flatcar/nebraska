package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
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

func TestGetInstancesPerOEMMetrics(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, err := as.AddTeam(&Team{Name: "oem_metrics_team"})
	require.NoError(t, err)
	tApp, err := as.AddApp(&Application{Name: "oem_metrics_app", TeamID: tTeam.ID})
	require.NoError(t, err)
	tPkg, err := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	require.NoError(t, err)
	tChannel, err := as.AddChannel(&Channel{Name: "oem_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	require.NoError(t, err)
	tGroup, err := as.AddGroup(&Group{Name: "oem_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	require.NoError(t, err)

	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.1", OEM: "azure"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.2", OEM: "azure"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.3", OEM: "aws"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	_, err = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.4"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)

	metrics, err := a.GetInstancesPerOEMMetrics()
	require.NoError(t, err)

	var got []InstancesPerOEMMetric
	for _, m := range metrics {
		if m.ApplicationName == "oem_metrics_app" {
			got = append(got, m)
		}
	}

	require.Equal(t, []InstancesPerOEMMetric{
		{ApplicationName: "oem_metrics_app", OEM: "aws", InstancesCount: 1},
		{ApplicationName: "oem_metrics_app", OEM: "azure", InstancesCount: 2},
		{ApplicationName: "oem_metrics_app", OEM: "unknown", InstancesCount: 1},
	}, got)
}
