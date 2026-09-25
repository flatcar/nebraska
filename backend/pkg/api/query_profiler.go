package api

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/flatcar/nebraska/backend/pkg/logger"
)

const (
	// DefaultSlowQueryThreshold is the default threshold for considering a query slow
	DefaultSlowQueryThreshold = 100 * time.Millisecond
)

var (
	queryDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "nebraska",
			Name:      "db_query_duration_seconds",
			Help:      "Database query execution duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"query_type"},
	)

	slowQueryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nebraska",
			Name:      "db_slow_queries_total",
			Help:      "Total number of slow database queries",
		},
		[]string{"query_type"},
	)

	queryErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nebraska",
			Name:      "db_query_errors_total",
			Help:      "Total number of database query errors",
		},
		[]string{"query_type"},
	)
)

// QueryProfiler wraps sqlx.DB to provide query performance profiling
type QueryProfiler struct {
	db                  *sqlx.DB
	logger              zerolog.Logger
	slowQueryThreshold  time.Duration
	enableSlowQueryLog  bool
	enableMetrics       bool
}

// QueryProfilerConfig holds configuration for the query profiler
type QueryProfilerConfig struct {
	SlowQueryThreshold time.Duration
	EnableSlowQueryLog bool
	EnableMetrics      bool
}

// NewQueryProfiler creates a new QueryProfiler wrapping the provided database
func NewQueryProfiler(db *sqlx.DB, config QueryProfilerConfig) *QueryProfiler {
	if config.SlowQueryThreshold == 0 {
		config.SlowQueryThreshold = DefaultSlowQueryThreshold
	}

	return &QueryProfiler{
		db:                  db,
		logger:              logger.New("query-profiler"),
		slowQueryThreshold:  config.SlowQueryThreshold,
		enableSlowQueryLog:  config.EnableSlowQueryLog,
		enableMetrics:       config.EnableMetrics,
	}
}

// RegisterMetrics registers the query profiling metrics with Prometheus
func RegisterMetrics() error {
	collectors := []prometheus.Collector{
		queryDurationHistogram,
		slowQueryCounter,
		queryErrorCounter,
	}

	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			// Ignore if already registered
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				return err
			}
		}
	}
	return nil
}

// profileQuery wraps query execution with profiling
func (qp *QueryProfiler) profileQuery(queryType string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record metrics if enabled
	if qp.enableMetrics {
		queryDurationHistogram.WithLabelValues(queryType).Observe(duration.Seconds())
		if err != nil {
			queryErrorCounter.WithLabelValues(queryType).Inc()
		}
	}

	// Log slow queries if enabled
	if qp.enableSlowQueryLog && duration > qp.slowQueryThreshold {
		if qp.enableMetrics {
			slowQueryCounter.WithLabelValues(queryType).Inc()
		}
		
		logEvent := qp.logger.Warn().
			Str("query_type", queryType).
			Dur("duration_ms", duration).
			Float64("duration_seconds", duration.Seconds()).
			Float64("threshold_seconds", qp.slowQueryThreshold.Seconds())
		
		if err != nil {
			logEvent = logEvent.Err(err)
		}
		
		logEvent.Msg("slow query detected")
	}

	return err
}

// Queryx executes a query with profiling
func (qp *QueryProfiler) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	var rows *sqlx.Rows
	var err error

	profErr := qp.profileQuery("select", func() error {
		rows, err = qp.db.Queryx(query, args...)
		return err
	})

	if profErr != nil {
		return nil, profErr
	}
	return rows, nil
}

// QueryRowx executes a single-row query with profiling
func (qp *QueryProfiler) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	row := qp.db.QueryRowx(query, args...)
	duration := time.Since(start)

	// Record metrics
	if qp.enableMetrics {
		queryDurationHistogram.WithLabelValues("select_one").Observe(duration.Seconds())
	}

	// Log slow queries
	if qp.enableSlowQueryLog && duration > qp.slowQueryThreshold {
		if qp.enableMetrics {
			slowQueryCounter.WithLabelValues("select_one").Inc()
		}
		qp.logger.Warn().
			Str("query_type", "select_one").
			Dur("duration_ms", duration).
			Float64("duration_seconds", duration.Seconds()).
			Msg("slow query detected")
	}

	return row
}

// Exec executes a non-query statement with profiling
func (qp *QueryProfiler) Exec(query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	profErr := qp.profileQuery("exec", func() error {
		result, err = qp.db.Exec(query, args...)
		return err
	})

	if profErr != nil {
		return nil, profErr
	}
	return result, nil
}

// ExecContext executes a non-query statement with context and profiling
func (qp *QueryProfiler) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	profErr := qp.profileQuery("exec", func() error {
		result, err = qp.db.ExecContext(ctx, query, args...)
		return err
	})

	if profErr != nil {
		return nil, profErr
	}
	return result, nil
}

// Get executes a query and scans result into dest with profiling
func (qp *QueryProfiler) Get(dest interface{}, query string, args ...interface{}) error {
	return qp.profileQuery("select_one", func() error {
		return qp.db.Get(dest, query, args...)
	})
}

// Select executes a query and scans results into dest with profiling
func (qp *QueryProfiler) Select(dest interface{}, query string, args ...interface{}) error {
	return qp.profileQuery("select", func() error {
		return qp.db.Select(dest, query, args...)
	})
}

// NamedExec executes a named query with profiling
func (qp *QueryProfiler) NamedExec(query string, arg interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	profErr := qp.profileQuery("exec", func() error {
		result, err = qp.db.NamedExec(query, arg)
		return err
	})

	if profErr != nil {
		return nil, profErr
	}
	return result, nil
}

// Begin starts a transaction
func (qp *QueryProfiler) Begin() (*sql.Tx, error) {
	return qp.db.Begin()
}

// Beginx starts a transaction returning an sqlx.Tx
func (qp *QueryProfiler) Beginx() (*sqlx.Tx, error) {
	return qp.db.Beginx()
}

// GetDB returns the underlying sqlx.DB for direct access when needed
func (qp *QueryProfiler) GetDB() *sqlx.DB {
	return qp.db
}

// Stats returns database statistics
func (qp *QueryProfiler) Stats() sql.DBStats {
	return qp.db.Stats()
}

// SetSlowQueryThreshold updates the slow query threshold
func (qp *QueryProfiler) SetSlowQueryThreshold(threshold time.Duration) {
	qp.slowQueryThreshold = threshold
	qp.logger.Info().
		Dur("threshold_ms", threshold).
		Float64("threshold_seconds", threshold.Seconds()).
		Msg("slow query threshold updated")
}

// GetSlowQueryThreshold returns the current slow query threshold
func (qp *QueryProfiler) GetSlowQueryThreshold() time.Duration {
	return qp.slowQueryThreshold
}

// LogQuerySummary logs a summary of query performance statistics
func (qp *QueryProfiler) LogQuerySummary() {
	stats := qp.db.Stats()
	qp.logger.Info().
		Int("max_open_connections", stats.MaxOpenConnections).
		Int("open_connections", stats.OpenConnections).
		Int("in_use", stats.InUse).
		Int("idle", stats.Idle).
		Int64("wait_count", stats.WaitCount).
		Dur("wait_duration", stats.WaitDuration).
		Int64("max_idle_closed", stats.MaxIdleClosed).
		Int64("max_idle_time_closed", stats.MaxIdleTimeClosed).
		Int64("max_lifetime_closed", stats.MaxLifetimeClosed).
		Msg("database connection pool stats")
}

// QueryStats holds statistics about a single query type
type QueryStats struct {
	QueryType    string
	Count        int64
	TotalTime    time.Duration
	AverageTime  time.Duration
	MaxTime      time.Duration
	MinTime      time.Duration
	SlowCount    int64
	ErrorCount   int64
}

// String returns a formatted string representation of query stats
func (qs *QueryStats) String() string {
	return fmt.Sprintf(
		"QueryType: %s, Count: %d, Total: %v, Avg: %v, Max: %v, Min: %v, Slow: %d, Errors: %d",
		qs.QueryType, qs.Count, qs.TotalTime, qs.AverageTime, qs.MaxTime, qs.MinTime, qs.SlowCount, qs.ErrorCount,
	)
}
