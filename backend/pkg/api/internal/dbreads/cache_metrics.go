package dbreads

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nebraska",
			Name:      "cache_hits_total",
			Help:      "Total number of cache hits",
		},
		[]string{"cache"},
	)

	cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nebraska",
			Name:      "cache_misses_total",
			Help:      "Total number of cache misses",
		},
		[]string{"cache"},
	)

	cacheInvalidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nebraska",
			Name:      "cache_invalidations_total",
			Help:      "Total number of cache invalidations",
		},
		[]string{"cache"},
	)

	cacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nebraska",
			Name:      "cache_size",
			Help:      "Current size of cache (number of entries)",
		},
		[]string{"cache"},
	)
)
