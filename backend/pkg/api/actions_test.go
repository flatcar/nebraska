package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestAddFlatcarAction(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeFlatcar, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})

	flatcarAction, err := as.AddFlatcarAction(&types.FlatcarAction{Event: "postinstall", Sha256: "fsdkjjfghsdakjfgaksdjfasd", PackageID: tPkg.ID})
	assert.NoError(t, err)

	flatcarActionX, err := a.GetFlatcarAction(tPkg.ID)
	assert.NoError(t, err)

	assert.Equal(t, flatcarAction.Event, flatcarActionX.Event)
	assert.Equal(t, flatcarAction.Sha256, flatcarActionX.Sha256)
}
