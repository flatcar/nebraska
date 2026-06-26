package api

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestGetActivity(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tVersion := "12.1.0"
	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: tVersion, ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	tInstance2, _ := a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, NewInstanceApplication(tApp.ID, tGroup2.ID, "1.0.0"))
	tFakeInstance, _ := a.RegisterInstance(Instance{ID: "{" + uuid.New().String() + "}", IP: "10.0.0.2"}, NewInstanceApplication(tApp.ID, tGroup2.ID, "1.0.0"))

	_ = a.newGroupActivityEntry(activityRolloutStarted, activitySuccess, tVersion, tApp.ID, tGroup.ID)
	_ = a.newGroupActivityEntry(activityRolloutStarted, activitySuccess, tVersion, tApp.ID, tGroup2.ID)
	_ = a.newInstanceActivityEntry(activityInstanceUpdateFailed, activityError, tVersion, tApp.ID, tGroup.ID, tInstance.ID)
	_ = a.newInstanceActivityEntry(activityInstanceUpdateFailed, activityError, tVersion, tApp.ID, tGroup2.ID, tInstance2.ID)
	_ = a.newGroupActivityEntry(activityInstanceUpdateFailed, activitySuccess, tVersion, tApp.ID, tGroup.ID)
	_ = a.newInstanceActivityEntry(activityInstanceUpdateFailed, activityError, tVersion, tApp.ID, tGroup.ID, tFakeInstance.ID)

	time.Sleep(10 * time.Millisecond)

	// this should ignore the entry for the fake instance
	activityEntries, err := a.GetActivity(tTeam.ID, ActivityQueryParams{AppID: tApp.ID, GroupID: tGroup.ID})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(activityEntries))

	activityEntries, err = a.GetActivity(tTeam.ID, ActivityQueryParams{Severity: activityError})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(activityEntries))

	activityEntries, err = a.GetActivity(tTeam.ID, ActivityQueryParams{InstanceID: tInstance2.ID})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(activityEntries))

	// when asked explicitly, fake instance won't be ignored
	activityEntries, err = a.GetActivity(tTeam.ID, ActivityQueryParams{InstanceID: tFakeInstance.ID})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(activityEntries))

	activityEntries, err = a.GetActivity(tTeam.ID, ActivityQueryParams{})
	assert.NoError(t, err)
	assert.Equal(t, 5, len(activityEntries))
	anActivity := activityEntries[0]

	hasRecentActivity := a.hasRecentRuntimeActivity(activityInstanceUpdateFailed, ActivityQueryParams{Severity: activitySuccess, AppID: tApp.ID, Version: tVersion, GroupID: tGroup.ID})
	assert.True(t, hasRecentActivity)

	_, err = a.GetActivity("invalidTeamID", ActivityQueryParams{})
	assert.Error(t, err, "Team id used must be a valid uuid.")

	activityEntries, err = a.GetActivity(uuid.New().String(), ActivityQueryParams{})
	assert.NoError(t, err)
	assert.Nil(t, activityEntries, "Team with this id doesn't exist")

	// We try counting with default Start==-3days, End==Now
	totalCount, err := a.GetActivityCount(tTeam.ID, ActivityQueryParams{})
	assert.NoError(t, err)
	assert.Equal(t, 5, totalCount)

	totalCount, err = a.GetActivityCount(tTeam.ID,
		ActivityQueryParams{
			Start: anActivity.CreatedTs.Add(time.Duration(-10) * time.Minute),
			End:   anActivity.CreatedTs.Add(time.Duration(10) * time.Minute),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 5, totalCount)

	// Can filter by GroupID, ChannelID, AppID, and InstanceID.
	totalCount, err = a.GetActivityCount(tTeam.ID,
		ActivityQueryParams{
			GroupID: tGroup.ID,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, totalCount)
}

func TestActivityRouting(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tVersion := "12.1.0"
	tTeam, _ := a.AddTeam(&Team{Name: "test_team_routing"})
	tApp, _ := a.AddApp(&Application{Name: "test_app_routing", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: tVersion, ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel_routing", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group_routing", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_ = a.newGroupActivityEntry(activityRolloutStarted, activitySuccess, tVersion, tApp.ID, tGroup.ID)
	_ = a.newChannelActivityEntry(activityChannelPackageUpdated, activityInfo, tVersion, tApp.ID, tChannel.ID)

	var runtimeCount, adminCount, runtimeAdminLeak, adminRuntimeLeak int
	_ = a.db.QueryRow("select count(*) from activity where application_id = $1", tApp.ID).Scan(&runtimeCount)
	_ = a.db.QueryRow("select count(*) from admin_activity where application_id = $1", tApp.ID).Scan(&adminCount)
	_ = a.db.QueryRow("select count(*) from activity where class = 6 and application_id = $1", tApp.ID).Scan(&runtimeAdminLeak)
	_ = a.db.QueryRow("select count(*) from admin_activity where class <> 6 and application_id = $1", tApp.ID).Scan(&adminRuntimeLeak)
	assert.Equal(t, 1, runtimeCount)
	assert.Equal(t, 1, adminCount)
	assert.Equal(t, 0, runtimeAdminLeak)
	assert.Equal(t, 0, adminRuntimeLeak)

	entries, err := a.GetActivity(tTeam.ID, ActivityQueryParams{AppID: tApp.ID})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entries))

	classes := []int{entries[0].Class, entries[1].Class}
	assert.ElementsMatch(t, []int{activityRolloutStarted, activityChannelPackageUpdated}, classes)

	assert.True(t, a.hasRecentRuntimeActivity(activityRolloutStarted, ActivityQueryParams{AppID: tApp.ID, Version: tVersion, GroupID: tGroup.ID}))
}
