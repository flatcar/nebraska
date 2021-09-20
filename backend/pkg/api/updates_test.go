package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

const testDuration = "1d"

func TestGetUpdatePackage(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp2.ID})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "group2", ApplicationID: tApp2.ID, ChannelID: null.StringFrom(tChannel2.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, err := a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "1.0.0", "invalidApplicationID", tGroup.ID)
	assert.Error(t, ErrInvalidApplicationOrGroup, err, "Invalid application id.")

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, "invalidGroupID")
	assert.Error(t, err, "Invalid group id.")

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "1.0.0", uuid.New().String(), tGroup.ID)
	assert.Error(t, err, "Non existent application id.")

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, uuid.New().String())
	assert.Error(t, err, "Non existent group id.")

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup2.ID)
	assert.Error(t, err, "Group doesn't belong to the application provided.")

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp2.ID, tGroup2.ID)
	assert.Equal(t, ErrNoPackageFound, err, "Group's channel has no package bound.")

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "12.1.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrNoUpdatePackageAvailable, err, "Instance version is up to date.")

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "1010.5.0+2016-05-27-1832", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrNoUpdatePackageAvailable, err, "Instance version is up to date.")
}

func TestGetUpdatePackage_GroupNoChannel(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, PolicyUpdatesEnabled: false, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, _ = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.Error(t, ErrNoPackageFound)
}

func TestGetUpdatePackage_UpdatesDisabled(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: false, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, err := a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrUpdatesDisabled, err)
}

func TestGetUpdatePackage_MaxUpdatesPerPeriodLimitReached_SafeMode(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	safeMode := true

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: safeMode, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, err := a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxUpdatesPerPeriodLimitReached, err, "Safe mode is enabled, first update should be completed before letting more through.")
}

func TestGetUpdatePackage_MaxUpdatesPerPeriodLimitReached_LimitUpdated(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 1, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.New().String()
	_, err := a.GetUpdatePackage(instanceID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxUpdatesPerPeriodLimitReached, err, "Max 1 update per period, limit reached")

	tGroup.PolicyMaxUpdatesPerPeriod = 2
	_ = a.UpdateGroup(tGroup)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)
}

func TestGetUpdatePackage_MaxUpdatesLimitsReached(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	maxUpdatesPerPeriod := 2
	periodInterval := 500 * time.Millisecond
	periodIntervalSetting := fmt.Sprintf("%d milliseconds", periodInterval.Milliseconds())
	extraWaitPeriod := 10 * time.Millisecond // to avoid a race

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: periodIntervalSetting, PolicyMaxUpdatesPerPeriod: maxUpdatesPerPeriod, PolicyUpdateTimeout: "60 minutes"})

	newInstance1ID := uuid.New().String()

	_, err := a.GetUpdatePackage(newInstance1ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxUpdatesPerPeriodLimitReached, err)

	time.Sleep(periodInterval + extraWaitPeriod) // ensure that period interval is over but update timeout isn't

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxConcurrentUpdatesLimitReached, err, "Period interval is over, but there are still two updates not completed or failed.")

	_ = a.updateInstanceStatus(newInstance1ID, tApp.ID, InstanceStatusComplete)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)
}

func TestGetUpdatePackage_MaxTimedOutUpdatesLimitReached_SafeMode(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	periodInterval := 10 * time.Millisecond
	periodIntervalSetting := fmt.Sprintf("%d milliseconds", periodInterval.Milliseconds())
	updateTimeout := 500 * time.Millisecond
	updateTimeoutSetting := fmt.Sprintf("%d milliseconds", updateTimeout.Milliseconds())
	extraWaitPeriod := 10 * time.Millisecond // to avoid a race

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: periodIntervalSetting, PolicyMaxUpdatesPerPeriod: 1, PolicyUpdateTimeout: updateTimeoutSetting})

	_, err := a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	time.Sleep(periodInterval + extraWaitPeriod) // ensure that period interval is over but update timeout isn't

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxConcurrentUpdatesLimitReached, err)

	time.Sleep(updateTimeout - periodInterval + extraWaitPeriod) // ensure that update timeout is over

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxTimedOutUpdatesLimitReached, err)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrUpdatesDisabled, err)
}

func TestGetUpdatePackage_ResumeUpdates(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	maxUpdatesPerPeriod := 2
	periodInterval := 10 * time.Millisecond
	periodIntervalSetting := fmt.Sprintf("%d milliseconds", periodInterval.Milliseconds())
	updateTimeout := 500 * time.Millisecond
	updateTimeoutSetting := fmt.Sprintf("%d milliseconds", updateTimeout.Milliseconds())
	extraWaitPeriod := 10 * time.Millisecond // to avoid a race

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: periodIntervalSetting, PolicyMaxUpdatesPerPeriod: maxUpdatesPerPeriod, PolicyUpdateTimeout: updateTimeoutSetting})

	_, err := a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	time.Sleep(periodInterval + extraWaitPeriod) // ensure that period interval is over but update timeout isn't

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxConcurrentUpdatesLimitReached, err)

	time.Sleep(updateTimeout - periodInterval + extraWaitPeriod) // ensure that update timeout is over

	_, err = a.GetUpdatePackage(uuid.New().String(), "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)
}

func TestGetUpdatePackage_RolloutStats(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 4, PolicyUpdateTimeout: "60 minutes"})

	instance1, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	instance2, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	instance3, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, _ = a.GetUpdatePackage(instance1.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	_, _ = a.GetUpdatePackage(instance2.ID, "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	_, _ = a.GetUpdatePackage(instance3.ID, "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)

	group, _ := a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	stats, _ := a.GetGroupInstancesStats(group.ID, testDuration)
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, int64(1), stats.UpdateGranted.Int64)
	assert.Equal(t, int64(2), stats.OnHold.Int64)

	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")

	group, _ = a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	stats, _ = a.GetGroupInstancesStats(group.ID, testDuration)
	assert.Equal(t, int64(1), stats.Downloading.Int64)
	assert.Equal(t, int64(2), stats.OnHold.Int64)

	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	_, _ = a.GetUpdatePackage(instance2.ID, "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	_, _ = a.GetUpdatePackage(instance3.ID, "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)

	group, _ = a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	stats, _ = a.GetGroupInstancesStats(group.ID, testDuration)
	assert.Equal(t, int64(1), stats.Complete.Int64)
	assert.Equal(t, int64(2), stats.UpdateGranted.Int64)

	_ = a.RegisterEvent(instance2.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	_ = a.RegisterEvent(instance3.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultFailed, "", "")

	group, _ = a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	stats, _ = a.GetGroupInstancesStats(group.ID, testDuration)
	assert.Equal(t, int64(2), stats.Complete.Int64)
	assert.Equal(t, int64(1), stats.Error.Int64)

	_, _ = a.GetUpdatePackage(instance3.ID, "", "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	_ = a.RegisterEvent(instance3.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")

	group, _ = a.GetGroup(tGroup.ID)
	assert.False(t, group.RolloutInProgress)
	stats, _ = a.GetGroupInstancesStats(group.ID, testDuration)
	assert.Equal(t, int64(3), stats.Complete.Int64)
}

func TestGetUpdatePackage_CompletionStats(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 4, PolicyUpdateTimeout: "60 minutes"})

	addAndUpdateInstance := func() {
		tInstance, err := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
		assert.NoError(t, err)

		_, err = a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
		assert.NoError(t, err)

		err = a.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloading)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", EventUpdateDownloadFinished, ResultSuccess, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloaded)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateInstalled, ResultSuccess, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusInstalled)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusComplete)), instance.Application.Status)
	}

	addAndUpdateInstance()

	stats, _ := a.GetGroupInstancesStats(tGroup.ID, testDuration)
	assert.Equal(t, 1, stats.Total)

	// This instance has the group's current package's version already and reports no status.
	// We need to make sure it doesn't show up as undefined.
	instance1, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", tPkg.Version, tApp.ID, tGroup.ID)

	stats, _ = a.GetGroupInstancesStats(tGroup.ID, testDuration)
	assert.Equal(t, int64(0), stats.Undefined.Int64)
	assert.Equal(t, int64(2), stats.Complete.Int64)

	// Just ensuring that a call for getting an update in an already up to date instance won't change its status
	_, err := a.GetUpdatePackage(instance1.ID, "", "10.0.0.1", tPkg.Version, tApp.ID, tGroup.ID)
	assert.Error(t, err, "nebraska: no update package available")

	// This version has a version different from the group's current one, and reports no status, so the
	// status should be undefined.
	_, err = a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "0.1.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	stats, _ = a.GetGroupInstancesStats(tGroup.ID, testDuration)
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, int64(1), stats.Undefined.Int64)
	assert.Equal(t, int64(2), stats.Complete.Int64)

	// Remove channel from group
	tGroup.ChannelID = null.StringFromPtr(nil)
	err = a.UpdateGroup(tGroup)
	assert.NoError(t, err)

	// Without the channel set to the group, we cannot know the version pointed to by that group,
	// so all instances that don't have the explicit "complete" status will be reported as undefined.
	stats, _ = a.GetGroupInstancesStats(tGroup.ID, testDuration)
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, int64(2), stats.Undefined.Int64)
	assert.Equal(t, int64(1), stats.Complete.Int64)
}

func TestGetUpdatePackage_UpdateInProgressOnInstance(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.New().String()

	p1, err := a.GetUpdatePackage(instanceID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	p2, err := a.GetUpdatePackage(instanceID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, p1, p2)

	instance, err := a.GetInstance(instanceID, tApp.ID)
	assert.NoError(t, err)
	assert.True(t, instance.Application.UpdateInProgress)
	assert.Equal(t, "12.1.0", instance.Application.LastUpdateVersion.String)

	err = a.updateInstanceStatus(instanceID, tApp.ID, InstanceStatusDownloading)
	assert.NoError(t, err)
	_, err = a.GetUpdatePackage(instanceID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrUpdateInProgressOnInstance, err)
}

func TestGetUpdatePackage_CheckVersionForGrantedUpdate(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.New().String()

	_, err := a.GetUpdatePackage(instanceID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	instance, err := a.GetInstance(instanceID, tApp.ID)
	assert.NoError(t, err)
	assert.True(t, instance.Application.UpdateInProgress)
	assert.Equal(t, int64(InstanceStatusUpdateGranted), instance.Application.Status.Int64)
	assert.Equal(t, "12.1.0", instance.Application.LastUpdateVersion.String)
	assert.Equal(t, "12.0.0", instance.Application.Version)

	_, err = a.GetUpdatePackage(instanceID, "", "10.0.0.1", "12.1.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrNoUpdatePackageAvailable, err)

	instanceStatusHistory, err := a.GetInstanceStatusHistory(instanceID, tApp.ID, tGroup.ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, InstanceStatusComplete, instanceStatusHistory[0].Status)
	assert.Equal(t, "12.1.0", instanceStatusHistory[0].Version)
}

func TestGetUpdatePackage_InstanceStatusHistory(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 3, PolicyUpdateTimeout: "60 minutes"})

	instance1, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, _ = a.GetUpdatePackage(instance1.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")

	instanceStatusHistory, err := a.GetInstanceStatusHistory(instance1.ID, tApp.ID, tGroup.ID, 5)
	assert.NoError(t, err)
	assert.Equal(t, InstanceStatusComplete, instanceStatusHistory[0].Status)
	assert.Equal(t, tPkg.Version, instanceStatusHistory[0].Version)
	assert.Equal(t, InstanceStatusDownloading, instanceStatusHistory[1].Status)
	assert.Equal(t, tPkg.Version, instanceStatusHistory[1].Version)
	assert.Equal(t, InstanceStatusUpdateGranted, instanceStatusHistory[2].Status)
	assert.Equal(t, tPkg.Version, instanceStatusHistory[2].Version)
}
