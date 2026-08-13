package metrics

import (
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

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

	syncerLastSuccessTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "syncer_last_success_timestamp_seconds",
			Help:      "Unix timestamp of the last successful syncer check for a channel/arch",
		},
		[]string{
			"channel",
			"arch",
		},
	)
	syncerCheckFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nebraska",
			Name:      "syncer_check_failures_total",
			Help:      "Total number of failed syncer checks for a channel/arch",
		},
		[]string{
			"channel",
			"arch",
		},
	)

	syncerCheckDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "nebraska",
			Name:      "syncer_check_duration_seconds",
			Help:      "Duration of syncer Omaha requests to upstream Flatcar servers",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{
			"channel",
			"arch",
		},
	)

	syncerPackagesCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nebraska",
			Name:      "syncer_packages_created_total",
			Help:      "Total number of new packages created by the syncer for a channel/arch",
		},
		[]string{
			"channel",
			"arch",
		},
	)


	l = logger.New("nebraska")
)
// SyncerLastSuccessTimestamp records the last successful check time for a channel/arch.
func SyncerLastSuccessTimestamp(channel, arch string) {
	syncerLastSuccessTimestamp.WithLabelValues(channel, arch).SetToCurrentTime()
}

// SyncerCheckFailure increments the failure counter for a channel/arch.
func SyncerCheckFailure(channel, arch string) {
	syncerCheckFailuresTotal.WithLabelValues(channel, arch).Inc()
}

// SyncerCheckDuration observes how long an Omaha check took for a channel/arch.
func SyncerCheckDuration(channel, arch string, seconds float64) {
	syncerCheckDurationSeconds.WithLabelValues(channel, arch).Observe(seconds)
}

// SyncerPackageCreated increments the packages-created counter for a channel/arch.
func SyncerPackageCreated(channel, arch string) {
	syncerPackagesCreatedTotal.WithLabelValues(channel, arch).Inc()
}

// registerNebraskaMetrics registers the application metrics collector with the DefaultRegistrer.
func registerNebraskaMetrics() error {
	collectors := []prometheus.Collector{
		appInstancePerChannelGaugeMetric,
		failedUpdatesGaugeMetric,
		openConnections,
		inUseConnections,
		idleConnections,
		syncerLastSuccessTimestamp,
		syncerCheckFailuresTotal,
		syncerCheckDurationSeconds,
		syncerPackagesCreatedTotal,
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

	for _, metric := range aipcMetrics {
		appInstancePerChannelGaugeMetric.WithLabelValues(metric.ApplicationName, metric.Version, metric.ChannelName).Set(float64(metric.InstancesCount))
	}

	fuMetrics, err := api.GetFailedUpdatesMetrics()
	if err != nil {
		return fmt.Errorf("failed to get failed update metrics: %w", err)
	}

	for _, metric := range fuMetrics {
		failedUpdatesGaugeMetric.WithLabelValues(metric.ApplicationName).Set(float64(metric.FailureCount))
	}

	// db stats
	dbStats := api.DbStats()
	openConnections.Set(float64(dbStats.OpenConnections))
	inUseConnections.Set(float64(dbStats.InUse))
	idleConnections.Set(float64(dbStats.Idle))

	return nil
}
