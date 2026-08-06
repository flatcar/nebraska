package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
)

func TestAddGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := as.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tPkg2, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp2.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tChannel2, _ := as.AddChannel(&Channel{Name: "test_channel2", Color: "yellow", ApplicationID: tApp2.ID, PackageID: null.StringFrom(tPkg2.ID)})

	group := &Group{
		Name:                      "group1",
		Description:               "description",
		ApplicationID:             tApp.ID,
		ChannelID:                 null.StringFrom(tChannel.ID),
		PolicyUpdatesEnabled:      true,
		PolicySafeMode:            true,
		PolicyPeriodInterval:      "15 minutes",
		PolicyMaxUpdatesPerPeriod: 2,
		PolicyUpdateTimeout:       "60 minutes",
	}
	group, err := as.AddGroup(group)
	assert.NoError(t, err)
	assert.Equal(t, true, group.PolicyUpdatesEnabled)

	_, err = a.GetGroupUpdatesStats(group)
	assert.NoError(t, err)

	groupX, err := a.GetGroup(group.ID)
	assert.NoError(t, err)
	assert.Equal(t, group.Name, groupX.Name)
	assert.Equal(t, group.Description, groupX.Description)
	assert.Equal(t, group.PolicyUpdatesEnabled, groupX.PolicyUpdatesEnabled)
	assert.Equal(t, group.PolicySafeMode, groupX.PolicySafeMode)
	assert.Equal(t, group.PolicyPeriodInterval, groupX.PolicyPeriodInterval)
	assert.Equal(t, group.PolicyMaxUpdatesPerPeriod, groupX.PolicyMaxUpdatesPerPeriod)
	assert.Equal(t, group.PolicyUpdateTimeout, groupX.PolicyUpdateTimeout)
	assert.Equal(t, tApp.ID, groupX.ApplicationID)
	assert.Equal(t, null.StringFrom(tChannel.ID), groupX.ChannelID)
	assert.Equal(t, tChannel.Name, groupX.Channel.Name)
	assert.Equal(t, tPkg.Version, groupX.Channel.Package.Version)

	_, err = as.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel2.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	assert.Equal(t, ErrInvalidChannel, err, "Channel id used doesn't belong to the application id that this group will be bound to and it should.")
}

func TestUpdateGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp1, _ := as.AddApp(&Application{Name: "test_app1", TeamID: tTeam.ID})
	tApp2, _ := as.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp1.ID})
	tChannel1, _ := as.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp1.ID, PackageID: null.StringFrom(tPkg.ID)})
	tChannel2, _ := as.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp1.ID})
	tChannel3, _ := as.AddChannel(&Channel{Name: "test_channel3", Color: "red", ApplicationID: tApp2.ID})

	group, _ := as.AddGroup(&Group{Name: "group1", ApplicationID: tApp1.ID, ChannelID: null.StringFrom(tChannel1.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	group.Name = "group1_updated"
	group.PolicyUpdatesEnabled = true
	group.ChannelID = null.StringFrom(tChannel2.ID)
	err := as.UpdateGroup(group)
	assert.NoError(t, err)

	_, err = a.GetGroupUpdatesStats(group)
	assert.NoError(t, err)

	groupX, _ := a.GetGroup(group.ID)
	assert.Equal(t, group.Name, groupX.Name)
	assert.Equal(t, group.PolicyPeriodInterval, groupX.PolicyPeriodInterval)
	assert.Equal(t, group.PolicyUpdatesEnabled, groupX.PolicyUpdatesEnabled)
	assert.Equal(t, tChannel2.Name, groupX.Channel.Name)

	groupX.ApplicationID = tApp2.ID
	err = as.UpdateGroup(groupX)
	assert.NoError(t, err, "Application id cannot be updated, but it won't produce an error.")

	groupX, _ = a.GetGroup(group.ID)
	assert.Equal(t, tApp1.ID, groupX.ApplicationID)

	groupX.ChannelID = null.StringFrom(tChannel3.ID)
	err = as.UpdateGroup(groupX)
	assert.Equal(t, ErrInvalidChannel, err, "Channel id used doesn't belong to the application id that this group is bound to and it should.")
}

func TestDeleteGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	err := as.DeleteGroup(tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetGroup(tGroup.ID)
	assert.Error(t, err, "Trying to get deleted group.")
}

func TestGetGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	group, err := a.GetGroup(tGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, tGroup.Name, group.Name)
	assert.Equal(t, tApp.ID, group.ApplicationID)
	assert.Equal(t, tChannel.Name, group.Channel.Name)
	assert.Equal(t, tPkg.Version, group.Channel.Package.Version)

	_, err = a.GetGroup(uuid.New().String())
	assert.Error(t, err, "Trying to get non existent group.")
}

func TestGetGroups(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup1, _ := as.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := as.AddGroup(&Group{Name: "test_group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	groups, err := a.GetGroups(tApp.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, tGroup2.Name, groups[0].Name)
	assert.Equal(t, tGroup1.Name, groups[1].Name)
	assert.Equal(t, tChannel.Name, groups[1].Channel.Name)
	assert.Equal(t, tPkg.ID, groups[1].Channel.PackageID.String)
	assert.Equal(t, tPkg.Version, groups[1].Channel.Package.Version)
}

func TestGetGroupsFiltered(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	realInstanceID := uuid.New().String()
	fakeInstanceID1 := "{" + uuid.New().String() + "}"
	fakeInstanceID2 := "{" + uuid.New().String() + "}"
	_, _ = rs.RegisterInstance(Instance{ID: realInstanceID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	_, _ = rs.RegisterInstance(Instance{ID: fakeInstanceID1, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "2.0.0"))
	_, _ = rs.RegisterInstance(Instance{ID: fakeInstanceID2, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	groups, err := a.GetGroups(tApp.ID, 0, 0)
	assert.NoError(t, err)
	if assert.Len(t, groups, 1) {
		g := groups[0]
		stats, err := a.GetGroupInstancesStats(g.ID, testDuration)
		assert.NoError(t, err)
		assert.Equal(t, 1, stats.Total)
		versionBreakdown, vbErr := a.GetGroupVersionBreakdown(g.ID)
		assert.NoError(t, vbErr)
		if assert.Len(t, versionBreakdown, 1) {
			vb := versionBreakdown[0]
			assert.Equal(t, "1.0.0", vb.Version)
			assert.Equal(t, 1, vb.Instances)
		}
	}
}

func TestVersionBreakDownEmpty(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	_, err := as.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	assert.NoError(t, err)

	groups, err := a.GetGroups(tApp.ID, 0, 0)
	assert.NoError(t, err)
	g := groups[0]

	versionBreakdown, vbErr := a.GetGroupVersionBreakdown(g.ID)
	assert.NoError(t, vbErr)
	assert.Len(t, versionBreakdown, 0)
}

func TestGroupTrackName(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	trackName := "production"
	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup1, _ := as.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: trackName})

	id, err := a.GetGroupID(tApp.ID, trackName, ArchAll)
	assert.NoError(t, err)
	assert.Equal(t, tGroup1.ID, id)

	_, err = a.GetGroupID(tApp.ID, "", ArchAll)
	assert.Error(t, err, "no group found for track  and architecture amd64")

	_, err = a.GetGroupID(tApp.ID, "Phony", ArchAll)
	assert.Error(t, err, "no group found for track Phony and architecture amd64")

	tGroup2, err := as.AddGroup(&Group{Name: "test_group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	assert.NoError(t, err)
	assert.Equal(t, tGroup2.Track, tGroup2.ID)

	// Check adding two groups with the same track name, in different apps, and getting them
	tApp2, err := as.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	assert.NoError(t, err)
	tPkgApp2, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp2.ID})
	tChannelApp2, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp2.ID, PackageID: null.StringFrom(tPkgApp2.ID)})
	tGroupApp2, err := as.AddGroup(&Group{Name: "beta", ApplicationID: tApp2.ID, ChannelID: null.StringFrom(tChannelApp2.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "beta"})
	assert.NoError(t, err)

	tGroup3, err := as.AddGroup(&Group{Name: "beta", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "beta"})
	assert.NoError(t, err)

	betaGroupID, err := a.GetGroupID(tApp2.ID, "beta", ArchAll)
	assert.NoError(t, err)
	assert.Equal(t, tGroupApp2.ID, betaGroupID)

	betaGroupID, err = a.GetGroupID(tApp.ID, "beta", ArchAll)
	assert.NoError(t, err)
	assert.Equal(t, tGroup3.ID, betaGroupID)

	// Test group with a track name but no arch (because there's no channel assigned to it)
	tGroupNoChannel, err := as.AddGroup(&Group{Name: "unknown", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "unknown"})
	assert.NoError(t, err)

	_, err = a.GetGroupID(tApp.ID, tGroupNoChannel.Track, ArchAll)
	assert.Error(t, err, "no group found")
}
