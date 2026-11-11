package api

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// TestSingleConnection fails if a read or package save asks for a second
// database connection while it still holds the first one.
func TestSingleConnection(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	setup := setupFloors(t, a, "single_connection", []string{"1000.0.0"}, "2000.0.0")
	channel, err := as.AddChannel(&types.Channel{Name: "blacklisted", ApplicationID: setup.AppID, Arch: types.ArchAMD64})
	require.NoError(t, err)
	instance, err := runtimeSvc(a).RegisterInstance(types.Instance{ID: uuid.NewString(), IP: "127.0.0.1"},
		runtime.NewInstanceApplication(setup.AppID, setup.Group.ID, "500.0.0"))
	require.NoError(t, err)
	// The status history only loads the event for error entries.
	_, err = a.db().Exec(`INSERT INTO event (instance_id, application_id, error_code, event_type_id, created_ts)
		VALUES ($1, $2, '2504', (SELECT id FROM event_type LIMIT 1), now() - interval '1 minute')`, instance.ID, setup.AppID)
	require.NoError(t, err)
	_, err = a.db().Exec(`INSERT INTO instance_status_history (instance_id, application_id, group_id, status, version)
		VALUES ($1, $2, $3, $4, '500.0.0')`, instance.ID, setup.AppID, setup.Group.ID, types.InstanceStatusError)
	require.NoError(t, err)

	a.db().SetMaxOpenConns(1)
	// A query waiting for a second connection waits forever. Closing the pool
	// makes it fail, so the test fails instead of hanging.
	timer := time.AfterFunc(5*time.Second, a.Close)
	defer func() { require.True(t, timer.Stop(), "blocked waiting for a second database connection") }()

	_, err = a.GetApps(defaultTeamID, 1, 100)
	require.NoError(t, err)
	groups, err := a.GetGroups(setup.AppID, 1, 100)
	require.NoError(t, err)
	require.NotEmpty(t, groups)
	channels, err := a.GetChannels(setup.AppID, 1, 100)
	require.NoError(t, err)
	require.NotEmpty(t, channels)
	floorChannels, err := a.GetPackageFloorChannels(setup.Floors[0].ID)
	require.NoError(t, err)
	require.Len(t, floorChannels, 1)
	history, err := a.GetInstanceStatusHistory(instance.ID, setup.AppID, setup.Group.ID, 100)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, "2504", history[0].ErrorCode.String)

	// AddPackage only checks the channels when the blacklist is not empty.
	pkg, err := as.AddPackage(&types.Package{Type: types.PkgTypeFlatcar, Arch: types.ArchAMD64, ApplicationID: setup.AppID,
		Version: "3000.0.0", URL: "http://sample.url/pkg", ChannelsBlacklist: types.StringArray{channel.ID}})
	require.NoError(t, err)
	// UpdatePackage only checks channels that are new in the blacklist.
	pkg.ChannelsBlacklist = types.StringArray{channel.ID, setup.Channel.ID}
	require.NoError(t, as.UpdatePackage(pkg))
}
