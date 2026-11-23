package api

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestWithDefaultTimeZone(t *testing.T) {
	t.Run("adds UTC to a URL DSN that does not pin one", func(t *testing.T) {
		got := withDefaultTimeZone("postgres://postgres:nebraska@127.0.0.1:5432/nebraska?sslmode=disable&connect_timeout=10")

		u, err := url.Parse(got)
		require.NoError(t, err)
		assert.Equal(t, "UTC", u.Query().Get("timezone"))
		assert.Equal(t, "disable", u.Query().Get("sslmode"))
		assert.Equal(t, "10", u.Query().Get("connect_timeout"))
	})

	t.Run("leaves a URL DSN that already pins a time zone", func(t *testing.T) {
		dsn := "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?timezone=Asia%2FKolkata"
		assert.Equal(t, dsn, withDefaultTimeZone(dsn))
	})

	t.Run("leaves a time zone set through options", func(t *testing.T) {
		dsn := "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?options=-c%20timezone%3DAsia%2FKolkata"
		assert.Equal(t, dsn, withDefaultTimeZone(dsn))
	})

	t.Run("adds UTC to a keyword/value DSN", func(t *testing.T) {
		assert.Equal(t,
			"host=127.0.0.1 user=postgres dbname=nebraska timezone=UTC",
			withDefaultTimeZone("host=127.0.0.1 user=postgres dbname=nebraska"))
	})

	t.Run("leaves a keyword/value DSN that already pins one", func(t *testing.T) {
		dsn := "host=127.0.0.1 user=postgres timezone=Asia/Kolkata"
		assert.Equal(t, dsn, withDefaultTimeZone(dsn))
	})
}

// withSessionTimeZone returns dsn with the Postgres session time zone pinned
// to tz.
func withSessionTimeZone(t *testing.T, dsn, tz string) string {
	t.Helper()

	u, err := url.Parse(dsn)
	require.NoError(t, err)

	q := u.Query()
	q.Set("timezone", tz)
	u.RawQuery = q.Encode()

	return u.String()
}

// registerAged registers an instance and backdates its last check-in by age.
func registerAged(t *testing.T, a *API, rs *runtime.Service, appID, groupID, age string) string {
	t.Helper()

	id := uuid.New().String()
	_, err := rs.RegisterInstance(types.Instance{ID: id, IP: "10.0.0.1"},
		runtime.NewInstanceApplication(appID, groupID, "1.0.0"))
	require.NoError(t, err)

	_, err = a.db().Exec(
		"UPDATE instance_application SET last_check_for_updates = now() - $2::interval WHERE instance_id = $1",
		id, age)
	require.NoError(t, err)

	return id
}

func TestGroupInstancesStatsUnderNonUTCSession(t *testing.T) {
	// A managed Postgres often runs with a local session time zone. The DSN
	// pins one here on purpose: withDefaultTimeZone leaves an explicit time
	// zone alone, so this runs the stats queries the way they would run
	// against such a database.
	t.Setenv("NEBRASKA_DB_URL", withSessionTimeZone(t, testDBURL(), "Asia/Kolkata"))

	a := newForTest(t)
	defer a.Close()

	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, err := as.AddTeam(&types.Team{Name: "test_team"})
	require.NoError(t, err)
	tApp, err := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	require.NoError(t, err)
	tGroup, err := as.AddGroup(&types.Group{
		Name: "test_group", ApplicationID: tApp.ID,
		PolicyUpdatesEnabled: true, PolicyPeriodInterval: "15 minutes",
		PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes",
	})
	require.NoError(t, err)

	// One instance inside the one day window and one well outside it. With
	// the session five and a half hours ahead of UTC, the old instance is the
	// one the broken comparison pulls back in.
	registerAged(t, a, rs, tApp.ID, tGroup.ID, "2 hours")
	registerAged(t, a, rs, tApp.ID, tGroup.ID, "27 hours")

	stats, err := a.GetGroupInstancesStats(tGroup.ID, "1d")
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Total, "only the recent instance belongs in the one day window")
}
