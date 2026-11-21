package metrics

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsRefreshInterval(t *testing.T) {
	testCases := []struct {
		name        string
		envValue    string
		expectedDur time.Duration
	}{
		{
			name:        "default when env var is not set",
			envValue:    "",
			expectedDur: defaultMetricsUpdateInterval,
		},
		{
			name:        "valid custom interval",
			envValue:    "30s",
			expectedDur: 30 * time.Second,
		},
		{
			name:        "valid custom interval in minutes",
			envValue:    "2m",
			expectedDur: 2 * time.Minute,
		},
		{
			name:        "fallback on invalid format",
			envValue:    "invalid-duration",
			expectedDur: defaultMetricsUpdateInterval,
		},
		{
			name:        "fallback on negative duration",
			envValue:    "-10s",
			expectedDur: defaultMetricsUpdateInterval,
		},
		{
			name:        "fallback on zero duration",
			envValue:    "0s",
			expectedDur: defaultMetricsUpdateInterval,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue == "" {
				os.Unsetenv("NEBRASKA_METRICS_UPDATE_INTERVAL")
			} else {
				os.Setenv("NEBRASKA_METRICS_UPDATE_INTERVAL", tc.envValue)
			}
			defer os.Unsetenv("NEBRASKA_METRICS_UPDATE_INTERVAL")

			interval := getMetricsRefreshInterval()
			assert.Equal(t, tc.expectedDur, interval)
		})
	}
}
