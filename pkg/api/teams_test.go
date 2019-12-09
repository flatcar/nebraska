package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTeams(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	teams, err := a.GetTeams()
	assert.NoError(t, err)
	if assert.Len(t, teams, 1) && assert.NotNil(t, teams[0]) {
		team := teams[0]
		assert.Equal(t, "d89342dc-9214-441d-a4af-bdd837a3b239", team.ID)
		assert.Equal(t, "default", team.Name)
	}
}

func TestGetTeam(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	team, err := a.GetTeam()
	assert.NoError(t, err)
	if assert.NotNil(t, team) {
		assert.Equal(t, "d89342dc-9214-441d-a4af-bdd837a3b239", team.ID)
		assert.Equal(t, "default", team.Name)
	}
}

func TestUpdateTeam(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	{
		team := mustGetTeam(t, a)

		timestamp := team.CreatedAt
		team.Name = "The A-Team"
		err := a.UpdateTeam(team)
		assert.NoError(t, err)

		team = mustGetTeam(t, a)

		assert.Equal(t, "The A-Team", team.Name)
		timestamp2 := team.CreatedAt
		assert.Equal(t, timestamp, timestamp2)
	}

	{
		team := Team{
			// some nonexistent team id
			ID:   "cd24398d-4129-d144-fa4a-932b3a738ddb",
			Name: "Yadda, yadda, won't be saved",
		}
		err := a.UpdateTeam(&team)
		if assert.Error(t, err) {
			assert.Equal(t, ErrNoRowsAffected, err)
		}
	}
}

func TestAddTeam(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	timeZero := time.Time{}

	{
		teamID := "cd24398d-4129-d144-fa4a-932b3a738ddb"
		teamName := "The A-Team"
		newTeam, err := a.AddTeam(&Team{
			ID:   teamID,
			Name: teamName,
		})
		assert.NoError(t, err)
		if assert.NotNil(t, newTeam) {
			assert.Equal(t, teamID, newTeam.ID)
			assert.Equal(t, teamName, newTeam.Name)
			assert.NotEqual(t, timeZero, newTeam.CreatedAt)
		}
	}

	{
		teamName := "Another team"
		newTeam, err := a.AddTeam(&Team{
			Name: teamName,
		})
		assert.NoError(t, err)
		if assert.NotNil(t, newTeam) {
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", newTeam.ID)
			assert.NotEqual(t, "", newTeam.ID)
			assert.Equal(t, teamName, newTeam.Name)
			assert.NotEqual(t, timeZero, newTeam.CreatedAt)
		}
	}
}
