package metrics

import (
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/logger"
)

const (
	defaultMetricsUpdateInterval = 15 * time.Second
)

var (
	InstanceMetricsRegistry = prometheus.NewRegistry()
	InstanceMetricsHandler  = promhttp.HandlerFor(InstanceMetricsRegistry, promhttp.HandlerOpts{})

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

	latestInstanceStatsGaugeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "instance_count",
			Help:      "Number of instances per channel, version, and architecture",
		},
		[]string{
			"channel",
			"version",
			"arch",
			"timestamp",
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
)

// registerNebraskaMetrics registers the application metrics collector with the DefaultRegistrer.
func registerNebraskaMetrics() error {
	collectors := []prometheus.Collector{
		appInstancePerChannelGaugeMetric,
		failedUpdatesGaugeMetric,
		latestInstanceStatsGaugeMetric,
		openConnections,
		inUseConnections,
		idleConnections,
	}

	for _, collector := range collectors {
		if collector == latestInstanceStatsGaugeMetric {
			if err := InstanceMetricsRegistry.Register(collector); err != nil {
				return fmt.Errorf("registering instance stats collector: %w", err)
			}
		} else {
			err := prometheus.Register(collector)
			if err != nil {
				return fmt.Errorf("registering Prometheus collector: %w", err)
			}
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
	// reset instance stats on each refresh
	latestInstanceStatsGaugeMetric.Reset()

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

	lisMetrics, err := api.GetLatestInstanceStatsMetrics()
	if err != nil {
		return fmt.Errorf("failed to get latest instance stats metrics: %w", err)
	}

	for _, metric := range lisMetrics {
		latestInstanceStatsGaugeMetric.WithLabelValues(metric.ChannelName, metric.Version, metric.Arch, metric.Timestamp).Set(float64(metric.InstancesCount))
	}

	// db stats
	dbStats := api.DbStats()
	openConnections.Set(float64(dbStats.OpenConnections))
	inUseConnections.Set(float64(dbStats.InUse))
	idleConnections.Set(float64(dbStats.Idle))

	return nil
}
