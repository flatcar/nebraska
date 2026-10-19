package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
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
	return admin.NewService(a.Conn(), a.Reads())
}

// runtimeSvc returns a runtime.Service that reuses a's shared read queries so
// tests can exercise the runtime operations under test.
func runtimeSvc(a *api.API) *Service {
	return NewService(a.Conn(), a.Reads(), Config{DisableUpdatesOnFailedRollout: true})
}

const day = 24 * time.Hour

// backdateInstanceStatusHistory rewinds the created_ts of an instance's status history records
// by timeAgo relative to the database current timestamp (now()).
//
// Status transitions written via updateInstanceStatus always register with created_ts = now().
// To verify multi-bucket timeline aggregation (e.g. 1d, 7d, 30d queries) and historical window
// filtering, this helper simulates past events by updating created_ts.
func backdateInstanceStatusHistory(t *testing.T, a *api.API, instanceID string, timeAgo time.Duration) {
	t.Helper()
	query := `UPDATE instance_status_history SET created_ts = now() - ($2 * interval '1 second') WHERE instance_id = $1`
	_, err := dbconn.DB(a.Conn()).Exec(query, instanceID, timeAgo.Seconds())
	require.NoError(t, err)
}
