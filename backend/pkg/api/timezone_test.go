package api

import (
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
)

// dbSessionTimeZones holds the Postgres session time zones the stats queries
// are exercised with: UTC plus one zone west and one east of it, so a query
// that mixes timestamptz values with a plain timestamp shifts its time window
// in both directions.
var dbSessionTimeZones = []string{"UTC", "America/New_York", "Asia/Kolkata"}

func TestWithDefaultTimeZone(t *testing.T) {
	tests := []struct {
		name     string
		dbURL    string
		expected string
	}{
		{
			name:     "url dsn without parameters",
			dbURL:    "postgres://postgres:nebraska@127.0.0.1:5432/nebraska",
			expected: "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?timezone=UTC",
		},
		{
			name:     "url dsn with parameters",
			dbURL:    "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?sslmode=disable",
			expected: "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?sslmode=disable&timezone=UTC",
		},
		{
			name:     "url dsn pinning a time zone already",
			dbURL:    "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?TimeZone=Europe/Berlin",
			expected: "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?TimeZone=Europe/Berlin",
		},
		{
			name:     "url dsn passing a time zone through options",
			dbURL:    "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?options=-c%20timezone%3DEurope%2FBerlin",
			expected: "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?options=-c%20timezone%3DEurope%2FBerlin",
		},
		{
			name:     "keyword/value dsn",
			dbURL:    "host=127.0.0.1 user=postgres dbname=nebraska",
			expected: "host=127.0.0.1 user=postgres dbname=nebraska timezone=UTC",
		},
		{
			name:     "keyword/value dsn pinning a time zone already",
			dbURL:    "host=127.0.0.1 user=postgres dbname=nebraska timezone=Europe/Berlin",
			expected: "host=127.0.0.1 user=postgres dbname=nebraska timezone=Europe/Berlin",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, withDefaultTimeZone(test.dbURL))
		})
	}
}

// TestInstanceStatsIgnoreSessionTimeZone checks that the instance counters
// shown in the dashboard only depend on how long ago an instance checked in,
// and not on the time zone the Postgres session happens to run in.
func TestInstanceStatsIgnoreSessionTimeZone(t *testing.T) {
	for _, timeZone := range dbSessionTimeZones {
		t.Run(timeZone, func(t *testing.T) {
			a := newForTestWithTimeZone(t, timeZone)
			defer a.Close()
			as := adminSvc(a)

			tTeam, err := as.AddTeam(&Team{Name: "test_team"})
			require.NoError(t, err)
			tApp, err := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
			require.NoError(t, err)
			tPkg, err := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "1.0.0", ApplicationID: tApp.ID})
			require.NoError(t, err)
			tChannel, err := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
			require.NoError(t, err)
			tGroup, err := as.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
			require.NoError(t, err)

			// The instance counters use a one day validity window, so only the
			// first two instances may be counted. The second and the third one
			// are close enough to the edge of the window to be misplaced by it
			// shrinking or growing by the offset of the session time zone.
			registerInstanceCheckedInAgo(t, a, tApp.ID, tGroup.ID, "2 hours")
			registerInstanceCheckedInAgo(t, a, tApp.ID, tGroup.ID, "22 hours")
			registerInstanceCheckedInAgo(t, a, tApp.ID, tGroup.ID, "27 hours")

			app, err := a.GetApp(tApp.ID)
			require.NoError(t, err)
			assert.Equal(t, 2, app.Instances.Count)

			instancesCount, err := a.GetInstancesCount(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID}, testDuration)
			require.NoError(t, err)
			assert.Equal(t, 2, instancesCount)

			instancesStats, err := a.GetGroupInstancesStats(tGroup.ID, testDuration)
			require.NoError(t, err)
			assert.Equal(t, 2, instancesStats.Total)

			versionBreakdown, err := a.GetGroupVersionBreakdown(tGroup.ID)
			require.NoError(t, err)
			require.Len(t, versionBreakdown, 1)
			assert.Equal(t, 2, versionBreakdown[0].Instances)
		})
	}
}

// TestUpdatesStatsIgnoreSessionTimeZone checks that the statistics driving the
// rollout policy classify granted updates by their real age, and not by an age
// shifted by the offset of the Postgres session time zone. Getting this wrong
// either times updates out early (disabling the updates of a whole group when
// the safe mode is on) or keeps them in progress forever, which holds new
// instances back.
func TestUpdatesStatsIgnoreSessionTimeZone(t *testing.T) {
	for _, timeZone := range dbSessionTimeZones {
		t.Run(timeZone, func(t *testing.T) {
			a := newForTestWithTimeZone(t, timeZone)
			defer a.Close()
			as := adminSvc(a)

			tTeam, err := as.AddTeam(&Team{Name: "test_team"})
			require.NoError(t, err)
			tApp, err := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
			require.NoError(t, err)
			tPkg, err := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "1.0.0", ApplicationID: tApp.ID})
			require.NoError(t, err)
			tChannel, err := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
			require.NoError(t, err)
			tGroup, err := as.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 10, PolicyUpdateTimeout: "60 minutes"})
			require.NoError(t, err)

			// One update granted well inside the group update timeout and one
			// granted well past it. Both were granted before the start of the
			// period the rate limit looks at.
			instanceInProgress := registerInstanceCheckedInAgo(t, a, tApp.ID, tGroup.ID, "5 minutes")
			grantUpdateAgo(t, a, instanceInProgress, "30 minutes")
			instanceTimedOut := registerInstanceCheckedInAgo(t, a, tApp.ID, tGroup.ID, "5 minutes")
			grantUpdateAgo(t, a, instanceTimedOut, "90 minutes")

			group, err := a.GetGroup(tGroup.ID)
			require.NoError(t, err)

			updatesStats, err := a.GetGroupUpdatesStats(group)
			require.NoError(t, err)
			assert.Equal(t, 1, updatesStats.UpdatesInProgress)
			assert.Equal(t, 1, updatesStats.UpdatesTimedOut)
			assert.Equal(t, 0, updatesStats.UpdatesGrantedInLastPeriod)
		})
	}
}

// newForTestWithTimeZone returns an API instance connected with the given
// Postgres session time zone.
func newForTestWithTimeZone(t *testing.T, timeZone string) *API {
	t.Helper()

	dbURL := os.Getenv("NEBRASKA_DB_URL")
	if dbURL == "" {
		dbURL = defaultTestDbURL
	}
	u, err := url.Parse(dbURL)
	require.NoError(t, err)
	params := u.Query()
	params.Set("timezone", timeZone)
	u.RawQuery = params.Encode()
	t.Setenv("NEBRASKA_DB_URL", u.String())

	a := newForTest(t)

	var sessionTimeZone string
	require.NoError(t, a.db().QueryRow("SHOW timezone").Scan(&sessionTimeZone))
	require.Equal(t, timeZone, sessionTimeZone)

	return a
}

// registerInstanceCheckedInAgo registers a new instance whose last update check
// happened the given interval ago, and returns its id. The timestamp is
// computed by Postgres from plain now(), which is not affected by the session
// time zone.
func registerInstanceCheckedInAgo(t *testing.T, a *API, appID, groupID, ago string) string {
	t.Helper()

	instanceID := uuid.New().String()
	_, err := runtimeSvc(a).RegisterInstance(Instance{ID: instanceID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(appID, groupID, "1.0.0"))
	require.NoError(t, err)

	_, err = a.db().Exec("UPDATE instance_application SET last_check_for_updates = now() - $1::interval WHERE instance_id = $2", ago, instanceID)
	require.NoError(t, err)

	return instanceID
}

// grantUpdateAgo marks an instance as having been granted an update the given
// interval ago, with the update still in progress.
func grantUpdateAgo(t *testing.T, a *API, instanceID, ago string) {
	t.Helper()

	_, err := a.db().Exec("UPDATE instance_application SET last_update_granted_ts = now() - $1::interval, last_update_version = '1.0.0', update_in_progress = true WHERE instance_id = $2", ago, instanceID)
	require.NoError(t, err)
}
