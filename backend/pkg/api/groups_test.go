package api

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestAddGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tPkg2, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp2.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "yellow", ApplicationID: tApp2.ID, PackageID: null.StringFrom(tPkg2.ID)})

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
	group, err := a.AddGroup(group)
	assert.NoError(t, err)
	assert.Equal(t, true, group.PolicyUpdatesEnabled)

	_, err = a.getGroupUpdatesStats(group)
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

	_, err = a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel2.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	assert.Equal(t, ErrInvalidChannel, err, "Channel id used doesn't belong to the application id that this group will be bound to and it should.")
}

func TestUpdateGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp1, _ := a.AddApp(&Application{Name: "test_app1", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp1.ID})
	tChannel1, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp1.ID, PackageID: null.StringFrom(tPkg.ID)})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp1.ID})
	tChannel3, _ := a.AddChannel(&Channel{Name: "test_channel3", Color: "red", ApplicationID: tApp2.ID})

	group, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp1.ID, ChannelID: null.StringFrom(tChannel1.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	group.Name = "group1_updated"
	group.PolicyUpdatesEnabled = true
	group.ChannelID = null.StringFrom(tChannel2.ID)
	err := a.UpdateGroup(group)
	assert.NoError(t, err)

	_, err = a.getGroupUpdatesStats(group)
	assert.NoError(t, err)

	groupX, _ := a.GetGroup(group.ID)
	assert.Equal(t, group.Name, groupX.Name)
	assert.Equal(t, group.PolicyPeriodInterval, groupX.PolicyPeriodInterval)
	assert.Equal(t, group.PolicyUpdatesEnabled, groupX.PolicyUpdatesEnabled)
	assert.Equal(t, tChannel2.Name, groupX.Channel.Name)

	groupX.ApplicationID = tApp2.ID
	err = a.UpdateGroup(groupX)
	assert.NoError(t, err, "Application id cannot be updated, but it won't produce an error.")

	groupX, _ = a.GetGroup(group.ID)
	assert.Equal(t, tApp1.ID, groupX.ApplicationID)

	groupX.ChannelID = null.StringFrom(tChannel3.ID)
	err = a.UpdateGroup(groupX)
	assert.Equal(t, ErrInvalidChannel, err, "Channel id used doesn't belong to the application id that this group is bound to and it should.")
}

func TestDeleteGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	err := a.DeleteGroup(tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetGroup(tGroup.ID)
	assert.Error(t, err, "Trying to get deleted group.")
}

func TestGetGroup(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

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

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup1, _ := a.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "test_group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

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

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	realInstanceID := uuid.New().String()
	fakeInstanceID1 := "{" + uuid.New().String() + "}"
	fakeInstanceID2 := "{" + uuid.New().String() + "}"
	_, _ = a.RegisterInstance(realInstanceID, "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	_, _ = a.RegisterInstance(fakeInstanceID1, "", "10.0.0.1", "2.0.0", tApp.ID, tGroup.ID)
	_, _ = a.RegisterInstance(fakeInstanceID2, "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

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

func TestGetVersionCountTimeline(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	version := "4.0.0"
	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	instanceID := uuid.New().String()

	_, _ = a.RegisterInstance(instanceID, "", "10.0.0.1", version, tApp.ID, tGroup.ID)

	instance, err := a.GetInstance(instanceID, tApp.ID)
	assert.NoError(t, err)
	_ = a.grantUpdate(instance, version)
	_ = a.updateInstanceStatus(instanceID, tApp.ID, InstanceStatusComplete)

	var versionTimelineMap map[time.Time](VersionCountMap)
	var isCache bool

	// get VersionCountTimeline from 1 hr before now
	_, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)

	// the first time the cache is not hit
	assert.Equal(t, false, isCache)

	time.Sleep(time.Second * 10)
	_, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)

	// the cache must be hit
	assert.Equal(t, true, isCache)

	time.Sleep(time.Second * 60)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)

	// the cache must be stale as we wait for the timespan
	assert.Equal(t, false, isCache)

	var totalInstances uint64
	for _, versionMap := range versionTimelineMap {
		totalInstances += versionMap[version]
	}
	assert.Equal(t, totalInstances, uint64(1))
	// for 1h we generate timestamp for every 15 minute so total timeline should have 5 timestamps
	assert.Equal(t, len(versionTimelineMap), 5)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1d")
	assert.NoError(t, err)
	// for 1d we generate timestamp for each hour so total timeline should have 25 timestamps
	assert.Equal(t, len(versionTimelineMap), 25)
	// the first time the cache is not hit
	assert.Equal(t, false, isCache)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "7d")
	assert.NoError(t, err)
	// for 7d we generate timestamp for each day so total timeline should have 8 timestamps
	assert.Equal(t, len(versionTimelineMap), 8)
	// the first time the cache is not hit
	assert.Equal(t, false, isCache)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "30d")
	assert.NoError(t, err)
	// for 30d we generate timestamp after each 3days so total timeline should have 11 timestamps
	assert.Equal(t, len(versionTimelineMap), 11)
	// the first time the cache is not hit
	assert.Equal(t, false, isCache)
}

func TestGetStatusCountTimeline(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	version := "4.0.0"
	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	instanceID1 := uuid.New().String()
	instanceID2 := uuid.New().String()

	_, _ = a.RegisterInstance(instanceID1, "", "10.0.0.1", version, tApp.ID, tGroup.ID)

	instance1, err := a.GetInstance(instanceID1, tApp.ID)
	assert.NoError(t, err)

	_ = a.grantUpdate(instance1, version)
	_ = a.updateInstanceStatus(instanceID1, tApp.ID, InstanceStatusComplete)
	_, _ = a.RegisterInstance(instanceID2, "", "10.0.0.2", version, tApp.ID, tGroup.ID)

	instance2, err := a.GetInstance(instanceID2, tApp.ID)
	assert.NoError(t, err)

	_ = a.grantUpdate(instance2, version)
	_ = a.updateInstanceStatus(instanceID2, tApp.ID, InstanceStatusDownloading)

	// get StatusCountTimeline from 1 hr before now
	statusTimelineMap, err := a.GetGroupStatusCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)
	// for 1h we generate timestamp for every 15 minute so total timeline should have 5 timestamps
	assert.Equal(t, len(statusTimelineMap), 5)
	var statusInstanceCountMap = make(map[int]uint64)
	for _, statusMap := range statusTimelineMap {
		statusInstanceCountMap[InstanceStatusComplete] += statusMap[InstanceStatusComplete][version]
		statusInstanceCountMap[InstanceStatusDownloading] += statusMap[InstanceStatusDownloading][version]
	}
	// as we registered two instances with version 4.0.0 with statuses 4 and 7
	// so our status breakdown should have count 1 for both status 4 and 7
	assert.Equal(t, statusInstanceCountMap[InstanceStatusComplete], uint64(1))
	assert.Equal(t, statusInstanceCountMap[InstanceStatusDownloading], uint64(1))

	statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "1d")
	assert.NoError(t, err)
	// for 1d we generate timestamp for each hour so total timeline should have 25 timestamps
	assert.Equal(t, len(statusTimelineMap), 25)

	statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "7d")
	assert.NoError(t, err)
	// for 7d we generate timestamp for each day so total timeline should have 8 timestamps
	assert.Equal(t, len(statusTimelineMap), 8)

	statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "30d")
	assert.NoError(t, err)
	// for 30d we generate timestamp after each 3days so total timeline should have 11 timestamps
	assert.Equal(t, len(statusTimelineMap), 11)
}

func TestGroupTrackName(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	trackName := "production"
	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup1, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: trackName})

	id, err := a.GetGroupID(tApp.ID, trackName, ArchAll)
	assert.NoError(t, err)
	assert.Equal(t, tGroup1.ID, id)

	_, err = a.GetGroupID(tApp.ID, "", ArchAll)
	assert.Error(t, err, "no group found for track  and architecture amd64")

	_, err = a.GetGroupID(tApp.ID, "Phony", ArchAll)
	assert.Error(t, err, "no group found for track Phony and architecture amd64")

	tGroup2, err := a.AddGroup(&Group{Name: "test_group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	assert.NoError(t, err)
	assert.Equal(t, tGroup2.Track, tGroup2.ID)

	// Check adding two groups with the same track name, in different apps, and getting them
	tApp2, err := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	assert.NoError(t, err)
	tPkgApp2, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp2.ID})
	tChannelApp2, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp2.ID, PackageID: null.StringFrom(tPkgApp2.ID)})
	tGroupApp2, err := a.AddGroup(&Group{Name: "beta", ApplicationID: tApp2.ID, ChannelID: null.StringFrom(tChannelApp2.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "beta"})
	assert.NoError(t, err)

	tGroup3, err := a.AddGroup(&Group{Name: "beta", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "beta"})
	assert.NoError(t, err)

	betaGroupID, err := a.GetGroupID(tApp2.ID, "beta", ArchAll)
	assert.NoError(t, err)
	assert.Equal(t, tGroupApp2.ID, betaGroupID)

	betaGroupID, err = a.GetGroupID(tApp.ID, "beta", ArchAll)
	assert.NoError(t, err)
	assert.Equal(t, tGroup3.ID, betaGroupID)

	// Test group with a track name but no arch (because there's no channel assigned to it)
	tGroupNoChannel, err := a.AddGroup(&Group{Name: "unknown", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "unknown"})
	assert.NoError(t, err)

	_, err = a.GetGroupID(tApp.ID, tGroupNoChannel.Track, ArchAll)
	assert.Error(t, err, "no group found")
}
