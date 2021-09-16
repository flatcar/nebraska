package api

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestRegisterEvent_InvalidParams(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	err := a.RegisterEvent(uuid.New().String(), tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.Equal(t, ErrInvalidInstance, err)

	err = a.RegisterEvent(tInstance.ID, uuid.New().String(), tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.Equal(t, ErrInvalidApplicationOrGroup, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, uuid.New().String(), EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.Equal(t, sql.ErrNoRows, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
	assert.Equal(t, ErrNoUpdateInProgress, err)

	_, _ = a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, 1000, ResultSuccess, "", "")
	assert.Equal(t, ErrInvalidEventTypeOrResult, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, 1000, "", "")
	assert.Equal(t, ErrInvalidEventTypeOrResult, err)
}

func TestRegisterEvent_TriggerEventConsequences(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	tInstance2, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.2", "1.0.0", tApp.ID, tGroup.ID)

	_, err := a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	err = a.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloading)), instance.Application.Status)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", EventUpdateDownloadFinished, ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloaded)), instance.Application.Status)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateInstalled, ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(InstanceStatusInstalled)), instance.Application.Status)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(InstanceStatusComplete)), instance.Application.Status)

	_, err = a.GetUpdatePackage(tInstance2.ID, "", "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	err = a.RegisterEvent(tInstance2.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultFailed, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance2.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(InstanceStatusError)), instance.Application.Status)
	group, _ := a.GetGroup(tGroup.ID)
	assert.Equal(t, true, group.PolicyUpdatesEnabled, "It wasn't the first update the one that failed.")
}

func TestRegisterEvent_TriggerEventConsequences_FirstUpdateAttemptFailed(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, err := a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultFailed, "", "")
	assert.NoError(t, err)
	instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(InstanceStatusError)), instance.Application.Status)
	group, _ := a.GetGroup(tGroup.ID)
	assert.Equal(t, false, group.PolicyUpdatesEnabled, "First update attempt failed.")
}

func TestRegisterEvent_CheckSuccessResult(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	performUpdate := func(tApp *Application, tGroup *Group, resultType int) {
		tInstance, err := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
		assert.NoError(t, err)

		_, err = a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
		assert.NoError(t, err)

		err = a.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
		assert.NoError(t, err)
		instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloading)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", EventUpdateDownloadFinished, ResultSuccess, "", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloaded)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateInstalled, ResultSuccess, "", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusInstalled)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, resultType, "", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusComplete)), instance.Application.Status)
	}

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	performUpdate(tApp, tGroup, ResultSuccess)
	performUpdate(tApp, tGroup, ResultSuccessReboot)
}

func TestRegisterEvent_CheckFlatcarSuccessResult(t *testing.T) {
	// If it's a Flatcar application, then the instances' updates are only considered to
	// be complete if the instance has sent ResultSuccessReboot on completion.
	a := newForTest(t)
	defer a.Close()

	performUpdate := func(tApp *Application, tGroup *Group, resultType, expectedInstanceStatus int) {
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

		err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, resultType, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(expectedInstanceStatus)), instance.Application.Status)
	}

	tApp, _ := a.GetApp(flatcarAppID)
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group9", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	performUpdate(tApp, tGroup, ResultSuccess, InstanceStatusInstalled)
	performUpdate(tApp, tGroup, ResultSuccessReboot, InstanceStatusComplete)
}

func TestRegisterEvent_CheckFlatcarIgnoredUpdate(t *testing.T) {
	// If it's a Flatcar application, and the instance reports that it updated to "" or 0.0.0.0 as the version,
	// then the event is ignored.
	a := newForTest(t)
	defer a.Close()

	tApp, _ := a.GetApp(flatcarAppID)
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group9", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	performUpdate := func(previousVersion string) {
		tInstance, err := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
		assert.NoError(t, err)

		_, err = a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
		assert.NoError(t, err)

		err = a.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, previousVersion, "")
		assert.NoError(t, err)
		instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloading)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", EventUpdateDownloadFinished, ResultSuccess, previousVersion, "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusDownloaded)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateInstalled, ResultSuccess, previousVersion, "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusInstalled)), instance.Application.Status)

		err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, previousVersion, "")
		assert.Error(t, err, "Received unexpected error: \nnebraska: flatcar event ignored")
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(InstanceStatusUndefined)), instance.Application.Status)
	}

	performUpdate("0.0.0.0")
	performUpdate("")
}

func TestRegisterEvent_GetEvent(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, err := a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetEvent(tInstance.ID, tApp.ID, time.Now())
	assert.Error(t, err, "sql: no rows in result set")

	err = a.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
	assert.NoError(t, err)

	errCode, err := a.GetEvent(tInstance.ID, tApp.ID, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, errCode, null.StringFrom(""))

	err = a.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, EventUpdateDownloadFinished, ResultSuccess, "", "")
	assert.NoError(t, err)

	errCode, err = a.GetEvent(tInstance.ID, tApp.ID, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, errCode, null.StringFrom(""))
}
