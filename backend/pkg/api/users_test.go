package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	defaultTeamID = "d89342dc-9214-441d-a4af-bdd837a3b239"
)

func TestGetUser(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	_, err := a.GetUser("non-existent")
	assert.Error(t, err)

	user, err := a.GetUser("admin")
	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, defaultTeamID, user.TeamID)
	assert.Equal(t, "8b31292d4778582c0e5fa96aee5513f1", user.Secret)
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
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, defaultTeamID, user.TeamID)
	assert.NotEqual(t, "8b31292d4778582c0e5fa96aee5513f1", user.Secret)
	assert.Equal(t, "c01e8daa7a6c135909f218ff2bea1cfe", user.Secret)
}

func TestAddUser(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	user := &User{
		Username: "chandler",
		Secret:   "shhhhh",
		TeamID:   defaultTeamID,
	}

	chandler, err := a.AddUser(user)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, chandler.Username)

	_, err = a.AddUser(user)
	assert.Error(t, err)
}

func TestGetUsersInTeam(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	_, err := a.GetUsersInTeam("non-existent")
	assert.Error(t, err)

	users, err := a.GetUsersInTeam(defaultTeamID)
	assert.NoError(t, err)
	assert.Equal(t, len(users), 1)

	teams, err := a.GetTeams()
	assert.NoError(t, err)
	assert.Equal(t, len(teams), 1)

	teamRoss, _ := a.AddTeam(&Team{Name: "team-ross"})
	assert.NoError(t, err)
	assert.Equal(t, teamRoss.Name, "team-ross")

	user := &User{
		Username: "chandler",
		Secret:   "shhhhh",
		TeamID:   teamRoss.ID,
	}

	chandler, err := a.AddUser(user)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, chandler.Username)

	defaultUsers, err := a.GetUsersInTeam(defaultTeamID)
	assert.NoError(t, err)
	assert.Equal(t, len(defaultUsers), 1, "Should still be one.")

	newTeamUsers, err := a.GetUsersInTeam(teamRoss.ID)
	assert.NoError(t, err)
	assert.Equal(t, len(newTeamUsers), 1, "Should also be one.")
}
