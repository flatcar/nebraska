package api

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestRegisterEvent_InvalidParams(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	err := rs.RegisterEvent(uuid.New().String(), tApp.ID, tGroup.ID, types.EventUpdateComplete, types.ResultSuccessReboot, "", "")
	assert.Equal(t, types.ErrInvalidInstance, err)

	err = rs.RegisterEvent(tInstance.ID, uuid.New().String(), tGroup.ID, types.EventUpdateComplete, types.ResultSuccessReboot, "", "")
	assert.Equal(t, types.ErrInvalidApplicationOrGroup, err)

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, uuid.New().String(), types.EventUpdateComplete, types.ResultSuccessReboot, "", "")
	assert.Equal(t, sql.ErrNoRows, err)

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateDownloadStarted, types.ResultSuccess, "", "")
	assert.Equal(t, types.ErrNoUpdateInProgress, err)

	_, _ = rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, 1000, types.ResultSuccess, "", "")
	assert.Equal(t, types.ErrInvalidEventTypeOrResult, err)

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, 1000, "", "")
	assert.Equal(t, types.ErrInvalidEventTypeOrResult, err)
}

func TestRegisterEvent_TriggerEventConsequences(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	tInstance2, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	_, err := rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)

	err = rs.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, types.EventUpdateDownloadStarted, types.ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloading)), instance.Application.Status)

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", types.EventUpdateDownloadFinished, types.ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloaded)), instance.Application.Status)

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateInstalled, types.ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(types.InstanceStatusInstalled)), instance.Application.Status)

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, types.ResultSuccessReboot, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(types.InstanceStatusComplete)), instance.Application.Status)

	_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance2.ID, IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)

	err = rs.RegisterEvent(tInstance2.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, types.ResultFailed, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance2.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(types.InstanceStatusError)), instance.Application.Status)
	group, _ := a.GetGroup(tGroup.ID)
	assert.Equal(t, true, group.PolicyUpdatesEnabled, "It wasn't the first update the one that failed.")
}

func TestRegisterEvent_TriggerEventConsequences_FirstUpdateAttemptFailed(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	_, err := rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)

	err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, types.ResultFailed, "", "")
	assert.NoError(t, err)
	instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, null.IntFrom(int64(types.InstanceStatusError)), instance.Application.Status)
	group, _ := a.GetGroup(tGroup.ID)
	assert.Equal(t, false, group.PolicyUpdatesEnabled, "First update attempt failed.")
}

func TestRegisterEvent_CheckSuccessResult(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	performUpdate := func(tApp *types.Application, tGroup *types.Group, resultType int) {
		tInstance, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
		assert.NoError(t, err)

		_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
		assert.NoError(t, err)

		err = rs.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, types.EventUpdateDownloadStarted, types.ResultSuccess, "", "")
		assert.NoError(t, err)
		instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloading)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", types.EventUpdateDownloadFinished, types.ResultSuccess, "", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloaded)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateInstalled, types.ResultSuccess, "", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusInstalled)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, resultType, "", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusComplete)), instance.Application.Status)
	}

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	performUpdate(tApp, tGroup, types.ResultSuccess)
	performUpdate(tApp, tGroup, types.ResultSuccessReboot)
}

func TestRegisterEvent_CheckFlatcarSuccessResult(t *testing.T) {
	// If it's a Flatcar application, then the instances' updates are only considered to
	// be complete if the instance has sent ResultSuccessReboot on completion.
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	performUpdate := func(tApp *types.Application, tGroup *types.Group, resultType, expectedInstanceStatus int) {
		tInstance, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
		assert.NoError(t, err)

		_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
		assert.NoError(t, err)

		err = rs.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, types.EventUpdateDownloadStarted, types.ResultSuccess, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloading)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", types.EventUpdateDownloadFinished, types.ResultSuccess, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloaded)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateInstalled, types.ResultSuccess, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusInstalled)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, resultType, "11.0.0", "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(expectedInstanceStatus)), instance.Application.Status)
	}

	tApp, _ := a.GetApp(flatcarAppID)
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group9", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	performUpdate(tApp, tGroup, types.ResultSuccess, types.InstanceStatusInstalled)
	performUpdate(tApp, tGroup, types.ResultSuccessReboot, types.InstanceStatusComplete)
}

func TestRegisterEvent_CheckFlatcarIgnoredUpdate(t *testing.T) {
	// If it's a Flatcar application, and the instance reports that it updated to "" or 0.0.0.0 as the version,
	// then the event is ignored.
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tApp, _ := a.GetApp(flatcarAppID)
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group9", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	performUpdate := func(previousVersion string) {
		tInstance, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
		assert.NoError(t, err)

		_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
		assert.NoError(t, err)

		err = rs.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, types.EventUpdateDownloadStarted, types.ResultSuccess, previousVersion, "")
		assert.NoError(t, err)
		instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloading)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", types.EventUpdateDownloadFinished, types.ResultSuccess, previousVersion, "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusDownloaded)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateInstalled, types.ResultSuccess, previousVersion, "")
		assert.NoError(t, err)
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusInstalled)), instance.Application.Status)

		err = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, types.ResultSuccessReboot, previousVersion, "")
		assert.Error(t, err, "Received unexpected error: \nnebraska: flatcar event ignored")
		instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
		assert.Equal(t, null.IntFrom(int64(types.InstanceStatusUndefined)), instance.Application.Status)
	}

	performUpdate("0.0.0.0")
	performUpdate("")
}

func TestRegisterEvent_GetEvent(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	_, err := rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)

	_, err = a.GetEvent(tInstance.ID, tApp.ID, time.Now())
	assert.Error(t, err, "sql: no rows in result set")

	err = rs.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, types.EventUpdateDownloadStarted, types.ResultSuccess, "", "")
	assert.NoError(t, err)

	errCode, err := a.GetEvent(tInstance.ID, tApp.ID, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, errCode, null.StringFrom(""))

	err = rs.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, types.EventUpdateDownloadFinished, types.ResultSuccess, "", "")
	assert.NoError(t, err)

	errCode, err = a.GetEvent(tInstance.ID, tApp.ID, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, errCode, null.StringFrom(""))
}
