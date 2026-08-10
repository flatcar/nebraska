package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
)

// newForTest creates an *api.API backed by a freshly initialised test DB,
// mirroring the api package's own newForTest helper.
func newForTest(t *testing.T) *api.API {
	a, err := api.NewForTest(api.OptionInitDB)

	require.NoError(t, err)
	require.NotNil(t, a)

	return a
}

// adminSvc returns an admin.Service that reuses a's shared read queries so
// runtime tests can build the domain fixtures (teams, apps, channels, groups)
// that only the admin write surface can create.
func adminSvc(a *api.API) *admin.Service {
	return admin.NewService(a.Reads())
}

// runtimeSvc returns a runtime.Service that reuses a's shared read queries so
// tests can exercise the runtime operations under test.
func runtimeSvc(a *api.API) *Service {
	return NewService(a.Reads(), Config{DisableUpdatesOnFailedRollout: true})
}
