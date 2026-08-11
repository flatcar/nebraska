package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

// TestAdminActivityRouting verifies that the admin-side activity writer
// (newChannelActivityEntry) writes only to the admin_activity table and never
// leaks into the runtime activity table. The runtime side is covered by the
// runtime package's TestRuntimeActivityRouting. Together they preserve the routing
// invariant introduced when the activity table was split (PR #1398).
func TestAdminActivityRouting(t *testing.T) {
	a, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()
	svc := NewService(a.Conn(), a.Reads())

	tVersion := "12.1.0"
	tTeam, _ := svc.AddTeam(&types.Team{Name: "test_team_routing"})
	tApp, _ := svc.AddApp(&types.Application{Name: "test_app_routing", TeamID: tTeam.ID})
	tPkg, _ := svc.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: tVersion, ApplicationID: tApp.ID})
	tChannel, _ := svc.AddChannel(&types.Channel{Name: "test_channel_routing", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})

	err = svc.newChannelActivityEntry(types.ActivityChannelPackageUpdated, types.ActivityInfo, tVersion, tApp.ID, tChannel.ID)
	require.NoError(t, err)

	db := svc.db
	var adminCount, adminRuntimeLeak, runtimeCount int
	_ = db.QueryRow("select count(*) from admin_activity where application_id = $1", tApp.ID).Scan(&adminCount)
	_ = db.QueryRow("select count(*) from admin_activity where class <> 6 and application_id = $1", tApp.ID).Scan(&adminRuntimeLeak)
	_ = db.QueryRow("select count(*) from activity where application_id = $1", tApp.ID).Scan(&runtimeCount)
	assert.Equal(t, 1, adminCount, "admin writer must write to admin_activity")
	assert.Equal(t, 0, adminRuntimeLeak, "admin_activity must only hold class 6 (admin) rows")
	assert.Equal(t, 0, runtimeCount, "admin writer must not write to the runtime activity table")

	// Checking the admin activity is visible through the GetActivity as well.
	entries, err := a.GetActivity(tTeam.ID, api.ActivityQueryParams{AppID: tApp.ID})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, types.ActivityChannelPackageUpdated, entries[0].Class, "admin activity must be visible through GetActivity/all_activity")
}
