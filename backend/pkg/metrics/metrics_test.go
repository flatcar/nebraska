package metrics

import (
	"os"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
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

// resetMetricsState clears both GaugeVecs and their label-tracking maps, so
// each test starts from a clean slate despite these being package-level
// singletons shared across the whole test binary.
func resetMetricsState(t *testing.T) {
	t.Helper()

	appInstancePerChannelGaugeMetric.Reset()
	failedUpdatesGaugeMetric.Reset()
	lastAipcLabelSets = map[string][]string{}
	lastFuLabelSets = map[string][]string{}
}

// collectLabelValues returns every value seen for labelName across all
// series currently exported by vec.
func collectLabelValues(t *testing.T, vec *prometheus.GaugeVec, labelName string) []string {
	t.Helper()

	ch := make(chan prometheus.Metric, 64)
	go func() {
		vec.Collect(ch)
		close(ch)
	}()

	var values []string

	for m := range ch {
		var dtoM dto.Metric
		require.NoError(t, m.Write(&dtoM))

		for _, lp := range dtoM.GetLabel() {
			if lp.GetName() == labelName {
				values = append(values, lp.GetValue())
			}
		}
	}

	return values
}

// TestCalculateMetricsRemovesStaleSeries checks that calculateMetrics removes
// Prometheus series whose underlying data has disappeared (e.g. every
// instance for a version/channel/arch went stale, or an application has no
// more failed updates) instead of leaving their last exported value in
// place forever, which would undermine the activity-window filtering added
// to GetAppInstancesPerChannelMetrics and make the metric disagree with the
// UI again. Regression test for the review discussion on
// https://github.com/flatcar/nebraska/pull/1580.
func TestCalculateMetricsRemovesStaleSeries(t *testing.T) {
	resetMetricsState(t)

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
		"series with no more matching instances must be synced away, not left exporting their last value")
	require.Equal(t, 0, testutil.CollectAndCount(failedUpdatesGaugeMetric),
		"series with no more matching instances must be synced away, not left exporting their last value")
}

// TestCalculateMetricsUsesNoneSentinelForNoChannel checks that instances
// whose group has no channel assigned are exported with an explicit
// channel="none" label, never channel="" — an empty label value is easy to
// miss or hard to filter for in PromQL. Regression test for the review
// discussion on https://github.com/flatcar/nebraska/pull/1580.
func TestCalculateMetricsUsesNoneSentinelForNoChannel(t *testing.T) {
	resetMetricsState(t)

	a := newAPIForTest(t)
	defer a.Close()

	db, err := sqlx.Connect("pgx", os.Getenv("NEBRASKA_DB_URL"))
	require.NoError(t, err)
	defer db.Close()

	// "Prod EC2 us-west-2" (bcaa68bc-...) has active instances; drop its
	// channel assignment so they fall into the no-channel bucket.
	_, err = db.Exec(`UPDATE groups SET channel_id = NULL WHERE id = 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f'`)
	require.NoError(t, err)

	require.NoError(t, calculateMetrics(a))

	channelValues := collectLabelValues(t, appInstancePerChannelGaugeMetric, "channel")
	require.Contains(t, channelValues, noChannelLabel)
	require.NotContains(t, channelValues, "")
}
