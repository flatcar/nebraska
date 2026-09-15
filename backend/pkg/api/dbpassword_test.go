package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// rotatingPassword answers with a credential that changes between connections,
// which is what an expiring token looks like from Nebraska's side.
type rotatingPassword struct {
	mu       sync.Mutex
	password string
	err      error
	calls    int
	users    []string
}

func (r *rotatingPassword) DBPassword(_ context.Context, user string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls++
	r.users = append(r.users, user)

	return r.password, r.err
}

func (r *rotatingPassword) set(password string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.password = password
}

func (r *rotatingPassword) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.calls
}

func (r *rotatingPassword) seenUsers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.users...)
}

// loginURLFor creates a login role and returns a copy of the test connection URL
// that authenticates as it. The password in the URL is deliberately wrong, so a
// connection only succeeds if the provider was consulted.
func loginURLFor(t *testing.T, a *API, role, password string) string {
	t.Helper()

	_, err := a.db().Exec(fmt.Sprintf("drop role if exists %s", role))
	require.NoError(t, err)

	_, err = a.db().Exec(fmt.Sprintf("create role %s login password '%s'", role, password))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = a.db().Exec(fmt.Sprintf("drop role if exists %s", role))
	})

	parsed, err := url.Parse(os.Getenv("NEBRASKA_DB_URL"))
	require.NoError(t, err)

	parsed.User = url.UserPassword(role, "not-the-password")

	return parsed.String()
}

func selectOne(a *API) error {
	var one int

	return a.db().QueryRow("select 1").Scan(&one)
}

func TestDBPasswordProviderIsUsedForEveryConnection(t *testing.T) {
	admin, err := New()
	require.NoError(t, err)
	t.Cleanup(func() { admin.Close() })

	const role = "nebraska_password_provider_test"

	dbURL := loginURLFor(t, admin, role, "first")

	provider := &rotatingPassword{password: "first"}
	SetDBPasswordProvider(provider)
	t.Cleanup(func() { SetDBPasswordProvider(nil) })

	t.Setenv("NEBRASKA_DB_URL", dbURL)
	// Without idle connections every query below needs a fresh physical one.
	t.Setenv("NEBRASKA_DB_MAX_IDLE_CONNS", "0")

	serving, err := New()
	require.NoError(t, err)

	defer serving.Close()

	require.NoError(t, selectOne(serving))

	_, err = admin.db().Exec(fmt.Sprintf("alter role %s password 'second'", role))
	require.NoError(t, err)
	provider.set("second")

	require.NoError(t, selectOne(serving), "a rotated credential should be picked up by later connections")

	require.Greater(t, provider.callCount(), 1)
	for _, user := range provider.seenUsers() {
		require.Equal(t, role, user)
	}
}

func TestDBPasswordProviderErrorFailsTheConnection(t *testing.T) {
	SetDBPasswordProvider(&rotatingPassword{err: errors.New("credential unavailable")})
	t.Cleanup(func() { SetDBPasswordProvider(nil) })

	_, err := New()
	require.ErrorContains(t, err, "credential unavailable")
}

func TestWithoutDBPasswordProviderTheURLPasswordIsUsed(t *testing.T) {
	require.Nil(t, dbPasswordFunc())

	a, err := New()
	require.NoError(t, err)

	defer a.Close()

	require.NoError(t, selectOne(a))
}
