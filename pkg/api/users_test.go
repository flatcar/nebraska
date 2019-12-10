package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	defaultTeamID = "d89342dc-9214-441d-a4af-bdd837a3b239"
)

func TestAddUser(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	team := mustGetTeam(t, a)

	secret, err := a.GenerateUserSecret("Foo", "Password")
	assert.NoError(t, err)

	user := &User{
		Username: "Foo",
		TeamID:   team.ID,
		Secret:   secret,
	}

	newUser, err := a.AddUser(user)
	assert.NoError(t, err)
	if assert.NotNil(t, newUser) {
		assert.Equal(t, "Foo", newUser.Username)
		assert.Equal(t, team.ID, newUser.TeamID)
		assert.Equal(t, secret, newUser.Secret)
	}
}

func TestGetUser(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	_, err := a.GetUser("non-existent")
	assert.Error(t, err)

	user, err := a.GetUser("admin")
	assert.NoError(t, err)
	if assert.NotNil(t, user) {
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, defaultTeamID, user.TeamID)
		assert.Equal(t, "8b31292d4778582c0e5fa96aee5513f1", user.Secret)
	}
}

func TestGetUsersInTeam(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	team1 := mustAddTeam(t, a, &Team{Name: "T1"})
	team2 := mustAddTeam(t, a, &Team{Name: "T2"})

	user11 := mustAddUser(t, a, &User{Username: "U11", TeamID: team1.ID, Secret: "foo1"})
	user12 := mustAddUser(t, a, &User{Username: "U12", TeamID: team1.ID, Secret: "foo1"})
	user13 := mustAddUser(t, a, &User{Username: "U13", TeamID: team1.ID, Secret: "foo1"})
	user21 := mustAddUser(t, a, &User{Username: "U21", TeamID: team2.ID, Secret: "foo2"})
	user22 := mustAddUser(t, a, &User{Username: "U22", TeamID: team2.ID, Secret: "foo2"})

	{
		users, err := a.GetUsersInTeam(team1.ID)
		assert.NoError(t, err)
		if assert.Len(t, users, 3) {
			names := map[string]struct{}{
				user11.Username: {},
				user12.Username: {},
				user13.Username: {},
			}
			for _, user := range users {
				assert.Contains(t, names, user.Username)
				delete(names, user.Username)
				assert.Equal(t, team1.ID, user.TeamID)
				assert.Equal(t, "foo1", user.Secret)
			}
		}
	}

	{
		users, err := a.GetUsersInTeam(team2.ID)
		assert.NoError(t, err)
		if assert.Len(t, users, 2) {
			names := map[string]struct{}{
				user21.Username: {},
				user22.Username: {},
			}
			for _, user := range users {
				assert.Contains(t, names, user.Username)
				delete(names, user.Username)
				assert.Equal(t, team2.ID, user.TeamID)
				assert.Equal(t, "foo2", user.Secret)
			}
		}
	}
}

func TestUpdateUserPassword(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	err := a.UpdateUserPassword("non-existent", "new-password")
	assert.Error(t, err)

	err = a.UpdateUserPassword("admin", "new-password")
	assert.NoError(t, err)

	user, err := a.GetUser("admin")
	assert.NoError(t, err)
	if assert.NotNil(t, user) {
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, defaultTeamID, user.TeamID)
		assert.NotEqual(t, "8b31292d4778582c0e5fa96aee5513f1", user.Secret)
		assert.Equal(t, "c01e8daa7a6c135909f218ff2bea1cfe", user.Secret)
	}
}
