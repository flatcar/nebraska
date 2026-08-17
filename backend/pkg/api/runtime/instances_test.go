package runtime

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestGetInstanceStatusHistory(t *testing.T) {
	// Update instance status several times and see if the history matches.

	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	newInstance1ID := uuid.New().String()
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: newInstance1ID, Alias: "analias", IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	assert.Equal(t, tInstance.Alias, "analias")

	instance, err := a.GetInstance(tInstance.ID, tApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", instance.IP)

	err = rs.grantUpdate(tInstance, "1.0.1")
	assert.NoError(t, err)

	err = rs.updateInstanceStatus(tInstance.ID, tApp.ID, types.InstanceStatusInstalled)
	assert.NoError(t, err)

	err = rs.updateInstanceStatus(tInstance.ID, tApp.ID, types.InstanceStatusComplete)
	assert.NoError(t, err)

	err = rs.grantUpdate(tInstance, "1.0.2")
	assert.NoError(t, err)

	history, err := a.GetInstanceStatusHistory(tInstance.ID, tApp.ID, tGroup.ID, 100)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(history))
	assert.Equal(t, history[0].Status, types.InstanceStatusUpdateGranted)
	assert.Equal(t, history[0].Version, "1.0.2")
	assert.Equal(t, history[1].Status, types.InstanceStatusComplete)
	assert.Equal(t, history[1].Version, "1.0.1")
	assert.Equal(t, history[2].Status, types.InstanceStatusInstalled)
	assert.Equal(t, history[2].Version, "1.0.1")
	assert.Equal(t, history[3].Status, types.InstanceStatusUpdateGranted)
	assert.Equal(t, history[3].Version, "1.0.1")
}
