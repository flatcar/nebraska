package runtime

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestGetUpdatePackage_MaxUpdatesLimitsReached(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	maxUpdatesPerPeriod := 2
	periodInterval := 500 * time.Millisecond
	periodIntervalSetting := fmt.Sprintf("%d milliseconds", periodInterval.Milliseconds())
	extraWaitPeriod := 10 * time.Millisecond // to avoid a race

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: periodIntervalSetting, PolicyMaxUpdatesPerPeriod: maxUpdatesPerPeriod, PolicyUpdateTimeout: "60 minutes"})

	newInstance1ID := uuid.New().String()

	_, err := rs.GetUpdatePackage(types.Instance{ID: newInstance1ID, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)

	_, err = rs.GetUpdatePackage(types.Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)

	_, err = rs.GetUpdatePackage(types.Instance{ID: uuid.New().String(), IP: "10.0.0.3"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.Equal(t, types.ErrMaxUpdatesPerPeriodLimitReached, err)

	time.Sleep(periodInterval + extraWaitPeriod) // ensure that period interval is over but update timeout isn't

	_, err = rs.GetUpdatePackage(types.Instance{ID: uuid.New().String(), IP: "10.0.0.3"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.Equal(t, types.ErrMaxConcurrentUpdatesLimitReached, err, "Period interval is over, but there are still two updates not completed or failed.")

	_ = rs.updateInstanceStatus(newInstance1ID, tApp.ID, types.InstanceStatusComplete)

	_, err = rs.GetUpdatePackage(types.Instance{ID: uuid.New().String(), IP: "10.0.0.3"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)
}

func TestGetUpdatePackage_UpdateInProgressOnInstance(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.New().String()

	p1, err := rs.GetUpdatePackage(types.Instance{ID: instanceID, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)

	p2, err := rs.GetUpdatePackage(types.Instance{ID: instanceID, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.NoError(t, err)
	assert.Equal(t, p1, p2)

	instance, err := a.GetInstance(instanceID, tApp.ID)
	assert.NoError(t, err)
	assert.True(t, instance.Application.UpdateInProgress)
	assert.Equal(t, "12.1.0", instance.Application.LastUpdateVersion.String)

	err = rs.updateInstanceStatus(instanceID, tApp.ID, types.InstanceStatusDownloading)
	assert.NoError(t, err)
	_, err = rs.GetUpdatePackage(types.Instance{ID: instanceID, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "12.0.0"))
	assert.Equal(t, types.ErrUpdateInProgressOnInstance, err)
}
