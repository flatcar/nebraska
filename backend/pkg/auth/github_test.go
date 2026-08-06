package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func newTestGithubAuth() *githubAuth {
	return &githubAuth{
		userSessionIDs: make(userSessionMap),
		teamToUsers:    make(teamToUsersMap),
	}
}

func teamPtr(team string) *string {
	return &team
}

// --- copyStringSlice ---

func TestCopyStringSlice(t *testing.T) {
	t.Run("nil input returns non-nil empty slice", func(t *testing.T) {
		got := copyStringSlice(nil)
		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("copy is independent of the original", func(t *testing.T) {
		original := []string{"a", "b", "c"}
		dup := copyStringSlice(original)
		require.Equal(t, original, dup)

		// Mutate the original after copying; the duplicate must not change.
		// This is the entire point of the helper — callers store this slice
		// long-term (readWriteTeams/readOnlyTeams) and must not alias the
		// caller's backing array.
		original[0] = "mutated"
		assert.Equal(t, "a", dup[0], "copyStringSlice must not alias the original backing array")
	})

	t.Run("empty input returns empty output", func(t *testing.T) {
		got := copyStringSlice([]string{})
		assert.Empty(t, got)
	})
}

// --- makeTeamName ---

func TestMakeTeamName(t *testing.T) {
	tests := []struct {
		name string
		org  string
		team string
		want string
	}{
		{name: "normal org and team", org: "flatcar", team: "maintainers", want: "flatcar/maintainers"},
		{name: "empty team", org: "flatcar", team: "", want: "flatcar/"},
		{name: "empty org", org: "", team: "maintainers", want: "/maintainers"},
		{name: "both empty", org: "", team: "", want: "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, makeTeamName(tt.org, tt.team))
		})
	}
}

// --- addSessionID ---

func TestAddSessionID(t *testing.T) {
	t.Run("session with a team registers user under that team", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-1", githubTeamData{org: "acme", team: teamPtr("dev")})

		require.Contains(t, gha.userSessionIDs, "alice")
		assert.Contains(t, gha.userSessionIDs["alice"], "sess-1")

		require.Contains(t, gha.teamToUsers, "acme/dev")
		assert.Contains(t, gha.teamToUsers["acme/dev"], "alice")
	})

	t.Run("org-level session (nil team) does not touch teamToUsers", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-1", githubTeamData{org: "acme", team: nil})

		require.Contains(t, gha.userSessionIDs, "alice")
		assert.Empty(t, gha.teamToUsers, "org-level (team=nil) sessions must not create a teamToUsers entry")
	})

	t.Run("multiple sessions for the same user accumulate", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-1", githubTeamData{org: "acme", team: teamPtr("dev")})
		gha.addSessionID("alice", "sess-2", githubTeamData{org: "acme", team: teamPtr("ops")})

		assert.Len(t, gha.userSessionIDs["alice"], 2)
		assert.Contains(t, gha.teamToUsers, "acme/dev")
		assert.Contains(t, gha.teamToUsers, "acme/ops")
	})

	t.Run("two users on the same team both appear in teamToUsers", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-1", githubTeamData{org: "acme", team: teamPtr("dev")})
		gha.addSessionID("bob", "sess-2", githubTeamData{org: "acme", team: teamPtr("dev")})

		assert.Contains(t, gha.teamToUsers["acme/dev"], "alice")
		assert.Contains(t, gha.teamToUsers["acme/dev"], "bob")
	})
}

// --- stealUserSessionIDs ---

func TestStealUserSessionIDs(t *testing.T) {
	t.Run("steals all sessions for a user across multiple teams", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-1", githubTeamData{org: "acme", team: teamPtr("dev")})
		gha.addSessionID("alice", "sess-2", githubTeamData{org: "acme", team: teamPtr("ops")})
		gha.addSessionID("bob", "sess-3", githubTeamData{org: "acme", team: teamPtr("dev")})

		stolen := gha.stealUserSessionIDs("alice")

		assert.ElementsMatch(t, []string{"sess-1", "sess-2"}, stolen)
		assert.NotContains(t, gha.userSessionIDs, "alice", "alice must be fully removed after stealing all her sessions")

		// bob was also on acme/dev; alice's removal must not affect him.
		assert.Contains(t, gha.teamToUsers["acme/dev"], "bob")
		assert.NotContains(t, gha.teamToUsers["acme/dev"], "alice")
	})

	t.Run("team entry is deleted entirely when the stolen user was its last member", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-1", githubTeamData{org: "acme", team: teamPtr("solo-team")})

		gha.stealUserSessionIDs("alice")

		assert.NotContains(t, gha.teamToUsers, "acme/solo-team", "team entry must be deleted, not left as an empty set")
	})

	t.Run("stealing an unknown user returns nil and does not panic", func(t *testing.T) {
		gha := newTestGithubAuth()
		stolen := gha.stealUserSessionIDs("nobody")
		assert.Nil(t, stolen)
	})
}

// --- stealUserSessionIDsForOrg ---

func TestStealUserSessionIDsForOrg(t *testing.T) {
	t.Run("only steals org-level sessions (team == nil), leaves team sessions intact", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "org-sess", githubTeamData{org: "acme", team: nil})
		gha.addSessionID("alice", "team-sess", githubTeamData{org: "acme", team: teamPtr("dev")})

		stolen := gha.stealUserSessionIDsForOrg("alice", "acme")

		assert.Equal(t, []string{"org-sess"}, stolen)
		require.Contains(t, gha.userSessionIDs, "alice", "alice still has a remaining team-scoped session")
		assert.Contains(t, gha.userSessionIDs["alice"], "team-sess")
		assert.NotContains(t, gha.userSessionIDs["alice"], "org-sess")
	})

	t.Run("does not steal org-level sessions from a different org", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-other-org", githubTeamData{org: "other", team: nil})

		stolen := gha.stealUserSessionIDsForOrg("alice", "acme")

		assert.Empty(t, stolen)
		assert.Contains(t, gha.userSessionIDs["alice"], "sess-other-org")
	})

	t.Run("user fully removed when their last session is stolen", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "org-sess", githubTeamData{org: "acme", team: nil})

		gha.stealUserSessionIDsForOrg("alice", "acme")

		assert.NotContains(t, gha.userSessionIDs, "alice")
	})
}

// --- stealUserSessionIDsForOrgAndTeam ---

func TestStealUserSessionIDsForOrgAndTeam(t *testing.T) {
	t.Run("only steals the matching org+team session, leaves org-level and other-team sessions", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "org-sess", githubTeamData{org: "acme", team: nil})
		gha.addSessionID("alice", "dev-sess", githubTeamData{org: "acme", team: teamPtr("dev")})
		gha.addSessionID("alice", "ops-sess", githubTeamData{org: "acme", team: teamPtr("ops")})

		stolen := gha.stealUserSessionIDsForOrgAndTeam("alice", "acme", "dev")

		assert.Equal(t, []string{"dev-sess"}, stolen)
		assert.NotContains(t, gha.userSessionIDs["alice"], "dev-sess")
		assert.Contains(t, gha.userSessionIDs["alice"], "org-sess")
		assert.Contains(t, gha.userSessionIDs["alice"], "ops-sess")
	})

	t.Run("removes the user from teamToUsers for that team only", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "dev-sess", githubTeamData{org: "acme", team: teamPtr("dev")})
		gha.addSessionID("bob", "dev-sess-2", githubTeamData{org: "acme", team: teamPtr("dev")})

		gha.stealUserSessionIDsForOrgAndTeam("alice", "acme", "dev")

		assert.NotContains(t, gha.teamToUsers["acme/dev"], "alice")
		assert.Contains(t, gha.teamToUsers["acme/dev"], "bob", "other team members must be unaffected")
	})

	t.Run("user fully removed from userSessionIDs when it was their only session", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "dev-sess", githubTeamData{org: "acme", team: teamPtr("dev")})

		gha.stealUserSessionIDsForOrgAndTeam("alice", "acme", "dev")

		assert.NotContains(t, gha.userSessionIDs, "alice")
	})
}

// --- stealSessionIDsForOrgAndTeam ---

func TestStealSessionIDsForOrgAndTeam(t *testing.T) {
	t.Run("steals sessions from every user on the team, not just one", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "alice-dev-sess", githubTeamData{org: "acme", team: teamPtr("dev")})
		gha.addSessionID("bob", "bob-dev-sess", githubTeamData{org: "acme", team: teamPtr("dev")})
		// alice also has an unrelated session on a different team.
		gha.addSessionID("alice", "alice-ops-sess", githubTeamData{org: "acme", team: teamPtr("ops")})

		stolen := gha.stealSessionIDsForOrgAndTeam("acme", "dev")

		assert.ElementsMatch(t, []string{"alice-dev-sess", "bob-dev-sess"}, stolen)

		// alice's unrelated ops session must survive.
		require.Contains(t, gha.userSessionIDs, "alice")
		assert.Contains(t, gha.userSessionIDs["alice"], "alice-ops-sess")
		assert.NotContains(t, gha.userSessionIDs["alice"], "alice-dev-sess")

		// bob had only the dev session, so he's fully removed.
		assert.NotContains(t, gha.userSessionIDs, "bob")
	})

	t.Run("team entry is deleted from teamToUsers unconditionally", func(t *testing.T) {
		gha := newTestGithubAuth()
		gha.addSessionID("alice", "sess-1", githubTeamData{org: "acme", team: teamPtr("dev")})

		gha.stealSessionIDsForOrgAndTeam("acme", "dev")

		assert.NotContains(t, gha.teamToUsers, "acme/dev")
	})

	t.Run("stealing from a team with no members returns nil and does not panic", func(t *testing.T) {
		gha := newTestGithubAuth()
		stolen := gha.stealSessionIDsForOrgAndTeam("acme", "nonexistent-team")
		assert.Nil(t, stolen)
	})
}