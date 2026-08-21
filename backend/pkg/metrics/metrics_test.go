package metrics

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" sql driver used below

	"github.com/flatcar/nebraska/backend/pkg/api"
)

const defaultTestDbURL string = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"

func newAPIForTest(t *testing.T) *api.API {
	t.Helper()

	if _, ok := os.LookupEnv("NEBRASKA_DB_URL"); !ok {
		require.NoError(t, os.Setenv("NEBRASKA_DB_URL", defaultTestDbURL))
	}

	a, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	require.NotNil(t, a)

	return a
}

// TestCalculateMetricsResetsStaleSeries checks that calculateMetrics removes
// Prometheus series whose underlying data has disappeared (e.g. every
// instance for a version/channel/arch went stale, or an application has no
// more failed updates) instead of leaving their last exported value in
// place forever, which would undermine the activity-window filtering added
// to GetAppInstancesPerChannelMetrics and make the metric disagree with the
// UI again. Regression test for the review discussion on
// https://github.com/flatcar/nebraska/pull/1580.
func TestCalculateMetricsResetsStaleSeries(t *testing.T) {
	a := newAPIForTest(t)
	defer a.Close()

	// Independent connection to the same test database, since api.API's
	// own connection is not exported outside pkg/api.
	db, err := sqlx.Connect("pgx", os.Getenv("NEBRASKA_DB_URL"))
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, calculateMetrics(a))
	require.Positive(t, testutil.CollectAndCount(appInstancePerChannelGaugeMetric),
		"sample_data.sql has active instances, so the gauge should have at least one series after the first calculation")
	require.Positive(t, testutil.CollectAndCount(failedUpdatesGaugeMetric),
		"sample_data.sql has instances with failed updates, so the gauge should have at least one series after the first calculation")

	// Make every sample instance look like it hasn't checked in for 400
	// days, and remove the one "failed update" event in the sample data,
	// so both underlying queries now return nothing.
	longAgo := time.Now().UTC().AddDate(0, 0, -400)
	_, err = db.Exec(`UPDATE instance_application SET last_check_for_updates = $1`, longAgo)
	require.NoError(t, err)
	_, err = db.Exec(`
		DELETE FROM event
		USING event_type
		WHERE event.event_type_id = event_type.id AND event_type.type = 3 AND event_type.result = 0`)
	require.NoError(t, err)

	require.NoError(t, calculateMetrics(a))

	require.Equal(t, 0, testutil.CollectAndCount(appInstancePerChannelGaugeMetric),
		"series with no more matching instances must be removed by Reset(), not left exporting their last value")
	require.Equal(t, 0, testutil.CollectAndCount(failedUpdatesGaugeMetric),
		"series with no more matching instances must be removed by Reset(), not left exporting their last value")
}
