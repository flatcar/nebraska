package metrics

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

// noChannelArchLabel is the "arch" label value used for instances whose
// group has no channel (or no group) assigned, since there is no
// architecture to report in that case.
const noChannelArchLabel = "none"

// noChannelLabel is the "channel" label value used for instances whose
// group has no channel (or no group) assigned. GetAppInstancesPerChannelMetrics
// reports this case as an empty ChannelName (channel.name can never itself be
// empty, per the check constraint requiring a non-empty name), which we map
// here to an explicit sentinel instead of exporting channel="" - an empty
// label value is easy to miss/hard to filter for in PromQL.
const noChannelLabel = "none"

const (
	defaultMetricsUpdateInterval = 15 * time.Second
)

var (
	appInstancePerChannelGaugeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "application_instances_per_channel",
			Help:      "Number of applications from specific channel running on instances",
		},
		[]string{
			"application",
			"version",
			"channel",
			// arch distinguishes channels that share the same name across
			// architectures (e.g. "stable" for amd64 and aarch64). This is a
			// new label: existing dashboards/alerts that group solely by
			// application/version/channel should add `by (arch)` or sum()
			// over the new label to keep prior totals.
			"arch",
		},
	)

	failedUpdatesGaugeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "failed_updates",
			Help:      "Number of failed updates of an application",
		},
		[]string{
			"application",
		},
	)

	openConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "open_db_connections",
			Help:      "Number of established connections both in use and idle",
		},
	)

	inUseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "in_use_db_connections",
			Help:      "Number of connections currently in use",
		},
	)

	idleConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "idle_db_connections",
			Help:      "Number of idle connections",
		},
	)

	l = logger.New("nebraska")

	// lastAipcLabelSets and lastFuLabelSets track the label combinations
	// exported on the previous calculateMetrics run, keyed by labelKey.
	// They're only ever read/written from the single ticker goroutine
	// started in RegisterAndInstrument (calculateMetrics is never called
	// concurrently with itself outside of tests, which also call it
	// sequentially), so no extra locking is needed.
	lastAipcLabelSets = map[string][]string{}
	lastFuLabelSets   = map[string][]string{}
)

// labelKey joins label values into a single map key so removed label
// combinations can be detected between calculateMetrics runs. It doesn't
// need to be collision-proof against label values containing the
// separator: a spurious collision would only cause an extra
// DeleteLabelValues + re-Set on the next run, never an incorrectly
// exported value.
func labelKey(labels []string) string {
	return strings.Join(labels, "\x00")
}

// syncGaugeVec upserts every entry in newLabelSets (using values for each
// series' new value) into vec, then removes any label combination present
// in prevLabelSets but absent from newLabelSets. Unlike Reset() followed by
// repopulating, this never makes the whole metric family briefly empty or
// partial, so a concurrent Prometheus scrape always observes either the
// previous or the new value for every series, never a gap.
func syncGaugeVec(vec *prometheus.GaugeVec, prevLabelSets, newLabelSets map[string][]string, values map[string]float64) {
	for key, labels := range newLabelSets {
		vec.WithLabelValues(labels...).Set(values[key])
	}
	for key, labels := range prevLabelSets {
		if _, stillPresent := newLabelSets[key]; !stillPresent {
			vec.DeleteLabelValues(labels...)
		}
	}
}

// registerNebraskaMetrics registers the application metrics collector with the DefaultRegistrer.
func registerNebraskaMetrics() error {
	collectors := []prometheus.Collector{
		appInstancePerChannelGaugeMetric,
		failedUpdatesGaugeMetric,
		openConnections,
		inUseConnections,
		idleConnections,
	}

	for _, collector := range collectors {
		err := prometheus.Register(collector)
		if err != nil {
			return err
		}
	}
	return nil
}

// getMetricsRefreshInterval returns the metrics update Interval key is set in the environment as time.Duration,
// NEBRASKA_METRICS_UPDATE_INTERVAL. The variable must be a string acceptable by time.ParseDuration
// If not returns the default update interval.
func getMetricsRefreshInterval() time.Duration {
	refreshIntervalEnvValue := os.Getenv("NEBRASKA_METRICS_UPDATE_INTERVAL")
	if refreshIntervalEnvValue == "" {
		return defaultMetricsUpdateInterval
	}

	refreshInterval, err := time.ParseDuration(refreshIntervalEnvValue)
	if err != nil || refreshInterval <= 0 {
		l.Warn().Str("value", refreshIntervalEnvValue).Msg("invalid NEBRASKA_METRICS_UPDATE_INTERVAL, it must be acceptable by time.ParseDuration and positive value")
		return defaultMetricsUpdateInterval
	}
	return refreshInterval
}

// registerAndInstrumentMetrics registers the application metrics and instruments them in configurable intervals.
func RegisterAndInstrument(api *api.API) error {
	// register application metrics
	err := registerNebraskaMetrics()
	if err != nil {
		return err
	}

	refreshInterval := getMetricsRefreshInterval()

	metricsTicker := time.NewTicker(refreshInterval)

	go func() {
		for {
			<-metricsTicker.C
			err := calculateMetrics(api)
			if err != nil {
				l.Error().Err(err).Msg("registerAndInstrumentMetrics updating the metrics")
			}
		}
	}()

	return nil
}

// calculateMetrics calculates the application metrics and updates the respective metric.
func calculateMetrics(api *api.API) error {
	aipcMetrics, err := api.GetAppInstancesPerChannelMetrics()
	if err != nil {
		return fmt.Errorf("failed to get app instances per channel metrics: %w", err)
	}

	// Sync (rather than Reset()+repopulate) so a series whose instances
	// all went stale, or whose channel was deleted, stops being exported
	// instead of keeping its last non-zero count forever — without ever
	// making the whole metric family briefly empty for a concurrent
	// scrape (see syncGaugeVec).
	newAipcLabelSets := make(map[string][]string, len(aipcMetrics))
	aipcValues := make(map[string]float64, len(aipcMetrics))
	for _, metric := range aipcMetrics {
		archLabel := noChannelArchLabel
		if metric.Arch >= 0 {
			archLabel = types.Arch(uint(metric.Arch)).String()
		}
		channelLabel := metric.ChannelName
		if channelLabel == "" {
			channelLabel = noChannelLabel
		}
		labels := []string{metric.ApplicationName, metric.Version, channelLabel, archLabel}
		key := labelKey(labels)
		newAipcLabelSets[key] = labels
		aipcValues[key] = float64(metric.InstancesCount)
	}
	syncGaugeVec(appInstancePerChannelGaugeMetric, lastAipcLabelSets, newAipcLabelSets, aipcValues)
	lastAipcLabelSets = newAipcLabelSets

	fuMetrics, err := api.GetFailedUpdatesMetrics()
	if err != nil {
		return fmt.Errorf("failed to get failed update metrics: %w", err)
	}

	// Same reasoning as above: an application with no failed updates left
	// should stop being exported instead of keeping its last count.
	newFuLabelSets := make(map[string][]string, len(fuMetrics))
	fuValues := make(map[string]float64, len(fuMetrics))
	for _, metric := range fuMetrics {
		labels := []string{metric.ApplicationName}
		key := labelKey(labels)
		newFuLabelSets[key] = labels
		fuValues[key] = float64(metric.FailureCount)
	}
	syncGaugeVec(failedUpdatesGaugeMetric, lastFuLabelSets, newFuLabelSets, fuValues)
	lastFuLabelSets = newFuLabelSets

	// db stats
	dbStats := api.DbStats()
	openConnections.Set(float64(dbStats.OpenConnections))
	inUseConnections.Set(float64(dbStats.InUse))
	idleConnections.Set(float64(dbStats.Idle))

	return nil
}
