package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func mustGetTeam(t *testing.T, a *API) *Team {
	t.Helper()
	tTeam, err := a.GetTeam()
	require.NoError(t, err)
	require.NotNil(t, tTeam)
	return tTeam
}

func mustAddTeam(t *testing.T, a *API, team *Team) *Team {
	t.Helper()
	tTeam, err := a.AddTeam(team)
	require.NoError(t, err)
	require.NotNil(t, tTeam)
	return tTeam
}

func mustGetApp(t *testing.T, a *API, appID string) *Application {
	t.Helper()
	tApp, err := a.GetApp(appID)
	require.NoError(t, err)
	require.NotNil(t, tApp)
	return tApp
}

func mustAddApp(t *testing.T, a *API, app *Application) *Application {
	t.Helper()
	tApp, err := a.AddApp(app)
	require.NoError(t, err)
	require.NotNil(t, tApp)
	return tApp
}

func mustAddPackage(t *testing.T, a *API, pkg *Package) *Package {
	t.Helper()
	tPkg, err := a.AddPackage(pkg)
	require.NoError(t, err)
	require.NotNil(t, tPkg)
	return tPkg
}

func mustAddChannel(t *testing.T, a *API, channel *Channel) *Channel {
	t.Helper()
	tChannel, err := a.AddChannel(channel)
	require.NoError(t, err)
	require.NotNil(t, tChannel)
	return tChannel
}

func mustAddGroup(t *testing.T, a *API, group *Group) *Group {
	t.Helper()
	tGroup, err := a.AddGroup(group)
	require.NoError(t, err)
	require.NotNil(t, tGroup)
	return tGroup
}

func mustAddUser(t *testing.T, a *API, user *User) *User {
	t.Helper()
	tUser, err := a.AddUser(user)
	require.NoError(t, err)
	require.NotNil(t, tUser)
	return tUser
}
