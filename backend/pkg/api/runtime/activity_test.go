package runtime

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

func TestGetActivity(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tVersion := "12.1.0"
	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: tVersion, ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := as.AddGroup(&types.Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	tInstance2, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, NewInstanceApplication(tApp.ID, tGroup2.ID, "1.0.0"))
	tFakeInstance, _ := rs.RegisterInstance(types.Instance{ID: "{" + uuid.New().String() + "}", IP: "10.0.0.2"}, NewInstanceApplication(tApp.ID, tGroup2.ID, "1.0.0"))

	_ = rs.newGroupActivityEntry(types.ActivityRolloutStarted, types.ActivitySuccess, tVersion, tApp.ID, tGroup.ID)
	_ = rs.newGroupActivityEntry(types.ActivityRolloutStarted, types.ActivitySuccess, tVersion, tApp.ID, tGroup2.ID)
	_ = rs.newInstanceActivityEntry(types.ActivityInstanceUpdateFailed, types.ActivityError, tVersion, tApp.ID, tGroup.ID, tInstance.ID)
	_ = rs.newInstanceActivityEntry(types.ActivityInstanceUpdateFailed, types.ActivityError, tVersion, tApp.ID, tGroup2.ID, tInstance2.ID)
	_ = rs.newGroupActivityEntry(types.ActivityInstanceUpdateFailed, types.ActivitySuccess, tVersion, tApp.ID, tGroup.ID)
	_ = rs.newInstanceActivityEntry(types.ActivityInstanceUpdateFailed, types.ActivityError, tVersion, tApp.ID, tGroup.ID, tFakeInstance.ID)

	time.Sleep(10 * time.Millisecond)

	// this should ignore the entry for the fake instance
	activityEntries, err := a.GetActivity(tTeam.ID, types.ActivityQueryParams{AppID: tApp.ID, GroupID: tGroup.ID})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(activityEntries))

	activityEntries, err = a.GetActivity(tTeam.ID, types.ActivityQueryParams{Severity: types.ActivityError})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(activityEntries))

	activityEntries, err = a.GetActivity(tTeam.ID, types.ActivityQueryParams{InstanceID: tInstance2.ID})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(activityEntries))

	// when asked explicitly, fake instance won't be ignored
	activityEntries, err = a.GetActivity(tTeam.ID, types.ActivityQueryParams{InstanceID: tFakeInstance.ID})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(activityEntries))

	activityEntries, err = a.GetActivity(tTeam.ID, types.ActivityQueryParams{})
	assert.NoError(t, err)
	assert.Equal(t, 5, len(activityEntries))
	anActivity := activityEntries[0]

	hasRecentActivity := a.HasRecentRuntimeActivity(types.ActivityInstanceUpdateFailed, types.ActivityQueryParams{Severity: types.ActivitySuccess, AppID: tApp.ID, Version: tVersion, GroupID: tGroup.ID})
	assert.True(t, hasRecentActivity)

	_, err = a.GetActivity("invalidTeamID", types.ActivityQueryParams{})
	assert.Error(t, err, "Team id used must be a valid uuid.")

	activityEntries, err = a.GetActivity(uuid.New().String(), types.ActivityQueryParams{})
	assert.NoError(t, err)
	assert.Nil(t, activityEntries, "Team with this id doesn't exist")

	// We try counting with default Start==-3days, End==Now
	totalCount, err := a.GetActivityCount(tTeam.ID, types.ActivityQueryParams{})
	assert.NoError(t, err)
	assert.Equal(t, 5, totalCount)

	totalCount, err = a.GetActivityCount(tTeam.ID,
		types.ActivityQueryParams{
			Start: anActivity.CreatedTs.Add(time.Duration(-10) * time.Minute),
			End:   anActivity.CreatedTs.Add(time.Duration(10) * time.Minute),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 5, totalCount)

	// Can filter by GroupID, ChannelID, AppID, and InstanceID.
	totalCount, err = a.GetActivityCount(tTeam.ID,
		types.ActivityQueryParams{
			GroupID: tGroup.ID,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, totalCount)
}

// TestRuntimeActivityRouting verifies that the runtime-side activity writer
// writes only to the runtime `activity` table and never leaks into `admin_activity`.
func TestRuntimeActivityRouting(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tVersion := "12.1.0"
	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team_routing"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app_routing", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: tVersion, ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel_routing", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group_routing", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_ = rs.newGroupActivityEntry(types.ActivityRolloutStarted, types.ActivitySuccess, tVersion, tApp.ID, tGroup.ID)

	var runtimeCount, adminCount, runtimeAdminLeak int
	_ = rs.db.QueryRow("select count(*) from activity where application_id = $1", tApp.ID).Scan(&runtimeCount)
	_ = rs.db.QueryRow("select count(*) from admin_activity where application_id = $1", tApp.ID).Scan(&adminCount)
	_ = rs.db.QueryRow("select count(*) from activity where class = 6 and application_id = $1", tApp.ID).Scan(&runtimeAdminLeak)
	assert.Equal(t, 1, runtimeCount, "runtime writer must write to the activity table")
	assert.Equal(t, 0, adminCount, "runtime writer must not write to admin_activity")
	assert.Equal(t, 0, runtimeAdminLeak, "activity must not hold class 6 (admin) rows")

	entries, err := a.GetActivity(tTeam.ID, types.ActivityQueryParams{AppID: tApp.ID})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, types.ActivityRolloutStarted, entries[0].Class)

	assert.True(t, a.HasRecentRuntimeActivity(types.ActivityRolloutStarted, types.ActivityQueryParams{AppID: tApp.ID, Version: tVersion, GroupID: tGroup.ID}))
}
