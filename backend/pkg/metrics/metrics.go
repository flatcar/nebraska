package metrics

import (
	"fmt"
	"os"
	"time"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultMetricsUpdateInterval = 5 * time.Second
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
	logger = util.NewLogger("nebraska")
)

// registerNebraskaMetrics registers the application metrics collector with the DefaultRegistrer.
func registerNebraskaMetrics() error {
	err := prometheus.Register(appInstancePerChannelGaugeMetric)
	if err != nil {
		return err
	}
	err = prometheus.Register(failedUpdatesGaugeMetric)
	if err != nil {
		return err
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
		logger.Warn().Str("value", refreshIntervalEnvValue).Msg("invalid NEBRASKA_METRICS_UPDATE_INTERVAL, it must be acceptable by time.ParseDuration and positive value")
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

	metricsTicker := time.Tick(refreshInterval)

	go func() {
		for {
			<-metricsTicker
			err := calculateMetrics(api)
			if err != nil {
				logger.Error().Err(err).Msg("registerAndInstrumentMetrics updating the metrics")
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

	return nil
}
