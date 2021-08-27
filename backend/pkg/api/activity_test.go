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
	tInstance, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	tInstance2, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.2", "1.0.0", tApp.ID, tGroup2.ID)
	tFakeInstance, _ := a.RegisterInstance("{"+uuid.New().String()+"}", "", "10.0.0.2", "1.0.0", tApp.ID, tGroup2.ID)

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

	hasRecentActivity := a.hasRecentActivity(activityInstanceUpdateFailed, ActivityQueryParams{Severity: activitySuccess, AppID: tApp.ID, Version: tVersion, GroupID: tGroup.ID})
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
