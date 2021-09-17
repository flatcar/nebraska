package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestAddFlatcarAction(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeFlatcar, URL: null.StringFrom("http://sample.url/pkg"), Version: "12.1.0", ApplicationID: tApp.ID, Filename: null.StringFrom("fname.txt"), Hash: null.StringFrom("2222222"), Size: null.StringFrom("22")})

	flatcarAction, err := a.AddFlatcarAction(&FlatcarAction{Event: "postinstall", Sha256: "fsdkjjfghsdakjfgaksdjfasd", PackageID: tPkg.ID})
	assert.NoError(t, err)

	flatcarActionX, err := a.GetFlatcarAction(tPkg.ID)
	assert.NoError(t, err)

	assert.Equal(t, flatcarAction.Event, flatcarActionX.Event)
	assert.Equal(t, flatcarAction.Sha256, flatcarActionX.Sha256)
}
