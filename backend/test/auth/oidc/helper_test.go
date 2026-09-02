package auth_test

import (
	"net"
	"testing"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
)

// newDBForTest is a helper function that
// establishes connection with test db and returns the db struct.
func newDBForTest(t *testing.T) *api.API {
	t.Helper()
	db, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	require.NotNil(t, db)
	return db
}

// adminSvc builds an admin.Service over the given API's shared read queries,
// matching how the composition root injects it into the server.
func adminSvc(db *api.API) *admin.Service {
	return admin.NewService(db.Conn(), db.Reads())
}

// runtimeSvc builds a runtime.Service over the given API's shared read queries,
// matching how the composition root injects it into the server.
func runtimeSvc(db *api.API) *runtime.Service {
	return runtime.NewService(db.Conn(), db.Reads(), runtime.Config{})
}

// newOIDCMockServer creates a new mockoidc server and returns it.
func newOIDCMockServer(t *testing.T) *mockoidc.MockOIDC {
	t.Helper()
	m, err := mockoidc.NewServer(nil)
	require.NoError(t, err)
	return m
}

// startOIDCMockServer starts the mockoidc server at 127.0.0.1:8080.
func startOIDCMockServer(t *testing.T, m *mockoidc.MockOIDC) {
	t.Helper()
	// create listener
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	require.NoError(t, err)
	// start server
	err = m.Start(ln, nil)
	require.NoError(t, err)
}
