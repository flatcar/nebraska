package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// Lookups must see rows written to the database by something other than this
// process.
func TestLookupsSeeRowsWrittenOutsideTheProcess(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, err := as.AddTeam(&types.Team{Name: "test_team"})
	require.NoError(t, err)
	tApp, err := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID, ProductID: null.StringFrom("io.test.Outside")})
	require.NoError(t, err)
	tPkg, err := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: types.ArchAMD64})
	require.NoError(t, err)
	tChannel, err := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID), Arch: types.ArchAMD64})
	require.NoError(t, err)
	tGroup, err := as.AddGroup(&types.Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "before"})
	require.NoError(t, err)

	// Resolve once first, so a cache would be warm.
	_, err = a.GetAppID("io.test.Outside")
	require.NoError(t, err)
	_, err = a.GetGroupID(tApp.ID, "before", types.ArchAMD64)
	require.NoError(t, err)

	newAppID := uuid.New().String()
	_, err = a.db().Exec("insert into application (id, name, team_id, product_id) values ($1, 'outside_app', $2, 'io.test.Created')", newAppID, tTeam.ID)
	require.NoError(t, err)
	_, err = a.db().Exec("update application set product_id = 'io.test.Renamed' where id = $1", tApp.ID)
	require.NoError(t, err)
	_, err = a.db().Exec("update groups set track = 'after' where id = $1", tGroup.ID)
	require.NoError(t, err)

	appID, err := a.GetAppID("io.test.Created")
	assert.NoError(t, err)
	assert.Equal(t, newAppID, appID)

	appID, err = a.GetAppID("io.test.Renamed")
	assert.NoError(t, err)
	assert.Equal(t, tApp.ID, appID)

	_, err = a.GetAppID("io.test.Outside")
	assert.Error(t, err, "the old product id must stop resolving")

	groupID, err := a.GetGroupID(tApp.ID, "after", types.ArchAMD64)
	assert.NoError(t, err)
	assert.Equal(t, tGroup.ID, groupID)

	_, err = a.GetGroupID(tApp.ID, "before", types.ArchAMD64)
	assert.Error(t, err, "the old track must stop resolving")

	_, err = a.db().Exec("delete from groups where id = $1", tGroup.ID)
	require.NoError(t, err)
	_, err = a.GetGroupID(tApp.ID, "after", types.ArchAMD64)
	assert.Error(t, err, "a deleted group must stop resolving")
}
