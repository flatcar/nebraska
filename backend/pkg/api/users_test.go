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
