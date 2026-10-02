package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// sampleAppArch is the arch of every channel created by sample_data.sql for
// "Sample application" (see migration 0007_add_package_arch.sql).
const sampleAppArch = int(types.ArchAMD64)

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
			Arch:            sampleAppArch,
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.1",
			ChannelName:     "Master",
			Arch:            sampleAppArch,
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.1",
			ChannelName:     "Stable",
			Arch:            sampleAppArch,
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.2",
			ChannelName:     "Master",
			Arch:            sampleAppArch,
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.2",
			ChannelName:     "Stable",
			Arch:            sampleAppArch,
			InstancesCount:  2,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.3",
			ChannelName:     "Master",
			Arch:            sampleAppArch,
			InstancesCount:  1,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.3",
			ChannelName:     "Stable",
			Arch:            sampleAppArch,
			InstancesCount:  4,
		},
		{
			ApplicationName: "Sample application",
			Version:         "1.0.4",
			ChannelName:     "Master",
			Arch:            sampleAppArch,
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
			require.Equal(t, -1, m.Arch, "instances without a channel must not report a real arch")
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

// TestGetAppInstancesPerChannelMetricsSeparatesArch checks that two channels
// that share a name but differ by arch (e.g. "stable" for amd64 and arm64,
// see migration 0008-arm-channels-groups.sql) are reported as separate
// rows instead of being folded together by GROUP BY channel_name. Regression
// test for https://github.com/flatcar/nebraska/issues/1562.
func TestGetAppInstancesPerChannelMetricsSeparatesArch(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	const (
		appID         = "b6458005-8f40-4627-b33b-be70a718c48e" // "Sample application"
		packageID     = "12697fa4-5f83-11e5-9d70-feff819cdc9f" // existing "Stable" package for appID
		archChannelID = "cb2deea8-5f83-11e5-9d70-feff819cdc9e"
		archGroupID   = "bcaa68bc-5f82-11e5-9d70-feff819cda00"
	)

	// A second "Stable" channel for the same application, but for arm64.
	_, err := a.db().Exec(`
		INSERT INTO channel (id, name, color, application_id, package_id, arch)
		VALUES ($1, 'Stable', '#0099FF', $2, $3, $4)`,
		archChannelID, appID, packageID, int(types.ArchAArch64))
	require.NoError(t, err)

	_, err = a.db().Exec(`
		INSERT INTO groups (id, name, description, policy_period_interval, policy_max_updates_per_period, policy_update_timeout, application_id, channel_id, track)
		VALUES ($1, 'Prod EC2 us-west-2 (arm64)', 'Production servers, west coast, arm64', '15 minutes', 2, '60 minutes', $2, $3, $4)`,
		archGroupID, appID, archChannelID, archGroupID)
	require.NoError(t, err)

	_, err = a.db().Exec(`INSERT INTO instance (id, ip) VALUES ('instance-arm1', '10.0.0.100')`)
	require.NoError(t, err)

	_, err = a.db().Exec(`
		INSERT INTO instance_application (version, instance_id, application_id, group_id)
		VALUES ('1.0.3', 'instance-arm1', $1, $2)`, appID, archGroupID)
	require.NoError(t, err)

	metrics, err := a.GetAppInstancesPerChannelMetrics()
	require.NoError(t, err)

	var amd64Stable103, arm64Stable103 int
	for _, m := range metrics {
		if m.ApplicationName == "Sample application" && m.ChannelName == "Stable" && m.Version == "1.0.3" {
			switch m.Arch {
			case int(types.ArchAMD64):
				amd64Stable103 = m.InstancesCount
			case int(types.ArchAArch64):
				arm64Stable103 = m.InstancesCount
			}
		}
	}

	require.Equal(t, 4, amd64Stable103, "existing amd64 Stable/1.0.3 count must be unaffected")
	require.Equal(t, 1, arm64Stable103, "the new arm64 instance must show up as its own series")
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
