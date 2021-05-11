package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

func main() {
	a, err := api.New()
	if err != nil {
		fail("failed to get API object: %v", err)
	}
	listTeamsF := flag.Bool("list-teams", false, "List teams")
	listUsersInTeamF := flag.String("list-users-in-team", "", "List users in team")
	addTeamF := flag.String("add-team", "", "Add team")
	addUserToTeamF := flag.String("add-user-to-team", "", "Add user to team (must use -add-team to specify the team, the team must exist)")
	changeTeamNameToF := flag.String("change-team-name-to", "", "Change team name to the passed value (must use -add-team to specify the team to change)")
	flag.Parse()

	if *listTeamsF {
		teams := getTeams(a)
		say("teams:")
		for _, team := range teams {
			say("  %s", team.Name)
		}
		return
	}
	if *listUsersInTeamF != "" {
		users := getUsersInTeam(a, *listUsersInTeamF)
		say("users in the %q team", *listUsersInTeamF)
		for _, user := range users {
			say("  %s", user.Username)
		}
		return
	}
	if *addUserToTeamF != "" {
		if *addTeamF == "" {
			fail(`use the "-add-team" flag to specify the team where the user should be added`)
		}
		addUserToTeam(a, *addUserToTeamF, *addTeamF)
		say("user %q added to team %q", *addUserToTeamF, *addTeamF)
		return
	}
	if *changeTeamNameToF != "" {
		if *addTeamF == "" {
			fail(`use the "-add-team" flag to specify the team to be changed`)
		}
		changeTeamName(a, *addTeamF, *changeTeamNameToF)
		say("team %q renamed to %q", *addTeamF, *changeTeamNameToF)
		return
	}
	if *addTeamF != "" {
		addTeam(a, *addTeamF)
		say("team %q added", *addTeamF)
		return
	}

	say("nothing to do")
}

func getUsersInTeam(a *api.API, teamName string) []*api.User {
	teamID := getTeamIDFor(a, teamName)
	users, err := a.GetUsersInTeam(teamID)
	if err != nil && err != sql.ErrNoRows {
		fail("failed to get users in team %q: %v", teamName, err)
	}
	return users
}

func addUserToTeam(a *api.API, userToAdd, teamName string) *api.User {
	_, err := a.GetUser(userToAdd)
	if err != sql.ErrNoRows {
		if err == nil {
			fail("user %q already exists", userToAdd)
		}
		fail("failed to check if user %q exists", userToAdd)
	}
	teamID := getTeamIDFor(a, teamName)
	secret, err := a.GenerateUserSecret(userToAdd, userToAdd)
	if err != nil {
		fail("failed to generate secret for user %q: %v", userToAdd, err)
	}
	user := &api.User{
		Username: userToAdd,
		Secret:   secret,
		TeamID:   teamID,
	}
	user, err = a.AddUser(user)
	if err != nil {
		fail("failed to add user %q to team %q: %v", userToAdd, teamName, err)
	}

	return user
}

func addTeam(a *api.API, teamName string) *api.Team {
	ensureNoTeam(a, teamName)
	team := &api.Team{
		Name: teamName,
	}
	team, err := a.AddTeam(team)
	if err != nil {
		fail("failed to add team %q: %v", teamName, err)
	}
	return team
}

func changeTeamName(a *api.API, oldName, newName string) {
	teamID := getTeamIDFor(a, oldName)
	ensureNoTeam(a, newName)
	team := api.Team{
		ID:   teamID,
		Name: newName,
	}
	if err := a.UpdateTeam(&team); err != nil {
		fail("failed to change team %q name to %q: %v", oldName, newName, err)
	}
}

func getTeamIDFor(a *api.API, teamName string) string {
	teams := getTeams(a)
	teamID := ""
	for _, team := range teams {
		if team.Name == teamName {
			teamID = team.ID
		}
	}
	if teamID == "" {
		fail("team %q not found", teamName)
	}
	return teamID
}

func ensureNoTeam(a *api.API, teamName string) {
	teams := getTeams(a)
	for _, team := range teams {
		if team.Name == teamName {
			fail("team %q already exists", teamName)
		}
	}
}

func getTeams(a *api.API) []*api.Team {
	teams, err := a.GetTeams()
	if err != nil && err != sql.ErrNoRows {
		fail("failed to get teams: %v", err)
	}
	return teams
}

func fail(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func say(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", a...)
}
