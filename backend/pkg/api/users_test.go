package api

import (
	"crypto/md5" //nolint:gosec // test fixture reproducing the legacy secret format on purpose
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
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
	as := adminSvc(a)

	err := as.UpdateUserPassword("non-existent", "new-password")
	assert.Error(t, err)

	err = as.UpdateUserPassword("admin", "new-password")
	assert.NoError(t, err)

	user, err := a.GetUser("admin")
	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, defaultTeamID, user.TeamID)
	// The stored secret is no longer the legacy md5 value, and is a
	// bcrypt hash (which, unlike md5, embeds a random salt, so it can't
	// be asserted against a fixed string).
	assert.NotEqual(t, "8b31292d4778582c0e5fa96aee5513f1", user.Secret)
	assert.True(t, strings.HasPrefix(user.Secret, "$2"), "expected a bcrypt hash, got %q", user.Secret)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Secret), []byte("new-password")))

	ok, isLegacy, err := admin.VerifyUserPassword("admin", "new-password", user.Secret)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.False(t, isLegacy)

	ok, _, err = admin.VerifyUserPassword("admin", "wrong-password", user.Secret)
	assert.NoError(t, err)
	assert.False(t, ok)
}

// TestGenerateUserSecretUsesRandomSalt confirms two users with the same
// password no longer get an identical stored secret, unlike with the
// previous unsalted md5 scheme.
func TestGenerateUserSecretUsesRandomSalt(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	secret1, err := as.GenerateUserSecret("alice", "same-password")
	assert.NoError(t, err)
	secret2, err := as.GenerateUserSecret("bob", "same-password")
	assert.NoError(t, err)

	assert.NotEqual(t, secret1, secret2)

	ok, _, err := admin.VerifyUserPassword("alice", "same-password", secret1)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = admin.VerifyUserPassword("bob", "same-password", secret2)
	assert.NoError(t, err)
	assert.True(t, ok)
}

// TestVerifyUserPasswordLegacyMD5 checks that an account whose secret is
// still in the pre-bcrypt md5(username:realm:password) format can be
// verified, so existing accounts keep working until they log in again.
func TestVerifyUserPasswordLegacyMD5(t *testing.T) {
	legacySecret := legacyMD5Secret(t, "chandler", "correct-password")

	ok, isLegacy, err := admin.VerifyUserPassword("chandler", "correct-password", legacySecret)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, isLegacy)

	ok, isLegacy, err = admin.VerifyUserPassword("chandler", "wrong-password", legacySecret)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.True(t, isLegacy)
}

// TestAuthenticateUpgradesLegacySecret checks that a successful login
// against a legacy md5 secret transparently rewrites it as a bcrypt
// hash, so the weaker scheme is only ever checked once per account.
func TestAuthenticateUpgradesLegacySecret(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	username, password := "chandler", "correct-password"
	user := &User{
		Username: username,
		Secret:   legacyMD5Secret(t, username, password),
		TeamID:   defaultTeamID,
	}
	_, err := as.AddUser(user)
	assert.NoError(t, err)

	authed, err := as.Authenticate(username, password)
	assert.NoError(t, err)
	assert.Equal(t, username, authed.Username)

	upgraded, err := a.GetUser(username)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(upgraded.Secret, "$2"), "expected the secret to be upgraded to bcrypt, got %q", upgraded.Secret)
	assert.NotEqual(t, user.Secret, upgraded.Secret)

	// A second login goes through the bcrypt path and doesn't touch the
	// secret again.
	_, err = as.Authenticate(username, password)
	assert.NoError(t, err)
	unchanged, err := a.GetUser(username)
	assert.NoError(t, err)
	assert.Equal(t, upgraded.Secret, unchanged.Secret)

	_, err = as.Authenticate(username, "wrong-password")
	assert.ErrorIs(t, err, types.ErrInvalidCredentials)
}

// legacyMD5Secret reproduces the pre-bcrypt secret format so tests can
// exercise the legacy verification and upgrade paths without depending
// on the actual password behind any seeded fixture data.
func legacyMD5Secret(t *testing.T, username, password string) string {
	t.Helper()
	h := md5.New() //nolint:gosec // test fixture reproducing the legacy format on purpose
	_, err := io.WriteString(h, username+":"+admin.Realm+":"+password)
	assert.NoError(t, err)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestAddUser(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	user := &User{
		Username: "chandler",
		Secret:   "shhhhh",
		TeamID:   defaultTeamID,
	}

	chandler, err := as.AddUser(user)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, chandler.Username)

	_, err = as.AddUser(user)
	assert.Error(t, err)
}

func TestGetUsersInTeam(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	_, err := a.GetUsersInTeam("non-existent")
	assert.Error(t, err)

	users, err := a.GetUsersInTeam(defaultTeamID)
	assert.NoError(t, err)
	assert.Equal(t, len(users), 1)

	teams, err := a.GetTeams()
	assert.NoError(t, err)
	assert.Equal(t, len(teams), 1)

	teamRoss, _ := as.AddTeam(&Team{Name: "team-ross"})
	assert.NoError(t, err)
	assert.Equal(t, teamRoss.Name, "team-ross")

	user := &User{
		Username: "chandler",
		Secret:   "shhhhh",
		TeamID:   teamRoss.ID,
	}

	chandler, err := as.AddUser(user)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, chandler.Username)

	defaultUsers, err := a.GetUsersInTeam(defaultTeamID)
	assert.NoError(t, err)
	assert.Equal(t, len(defaultUsers), 1, "Should still be one.")

	newTeamUsers, err := a.GetUsersInTeam(teamRoss.ID)
	assert.NoError(t, err)
	assert.Equal(t, len(newTeamUsers), 1, "Should also be one.")
}
