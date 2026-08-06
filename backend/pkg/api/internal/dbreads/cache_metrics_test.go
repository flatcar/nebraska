package dbreads

import (
	"testing"
)

// TestCacheMetricsBasic verifies that cache metrics can be recorded without errors
func TestCacheMetricsBasic(t *testing.T) {
	// Test that metrics can be incremented without panicking
	t.Run("CacheHits", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Cache hits metric panicked: %v", r)
			}
		}()
		cacheHits.WithLabelValues("groups").Inc()
		cacheHits.WithLabelValues("app_ids").Inc()
	})

	t.Run("CacheMisses", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Cache misses metric panicked: %v", r)
			}
		}()
		cacheMisses.WithLabelValues("groups").Inc()
		cacheMisses.WithLabelValues("app_ids").Inc()
	})

	t.Run("CacheInvalidations", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Cache invalidations metric panicked: %v", r)
			}
		}()
		cacheInvalidations.WithLabelValues("groups").Inc()
		cacheInvalidations.WithLabelValues("app_ids").Inc()
	})

	t.Run("CacheSize", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Cache size metric panicked: %v", r)
			}
		}()
		cacheSize.WithLabelValues("groups").Set(100)
		cacheSize.WithLabelValues("app_ids").Set(50)
	})
}

// TestCacheMetricsLabels verifies different cache labels work independently
func TestCacheMetricsLabels(t *testing.T) {
	// Verify both cache labels are accepted
	labels := []string{"groups", "app_ids"}
	
	for _, label := range labels {
		t.Run(label, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Metric with label %s panicked: %v", label, r)
				}
			}()
			
			cacheHits.WithLabelValues(label).Inc()
			cacheMisses.WithLabelValues(label).Inc()
			cacheInvalidations.WithLabelValues(label).Inc()
			cacheSize.WithLabelValues(label).Set(42)
		})
	}
}
