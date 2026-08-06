// Package runtime provides operations for runtime-managed tables
// (instance, instance_application, instance_status_history, event, activity, group_local, instance_stats).
package runtime

import (
	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

var l = logger.New("runtime")

// Service provides runtime operations. It embeds the api's shared
// dbreads.Queries for read access using the same DB connection.
type Service struct {
	*dbreads.Queries
	db *sqlx.DB

	// disableUpdatesOnFailedRollout defines whether to disable updates
	// after a first rollout attempt failed (ResultFailed)
	disableUpdatesOnFailedRollout bool
}

// Config holds construction-time options for the runtime Service.
type Config struct {
	// DisableUpdatesOnFailedRollout, when true, trips the safe-mode brake that
	// disables updates after a first rollout attempt fails (ResultFailed).
	DisableUpdatesOnFailedRollout bool
}

// NewService creates a new runtime Service that reuses the given read queries
// (and their underlying DB connection).
func NewService(q *dbreads.Queries, cfg Config) *Service {
	return &Service{
		Queries:                       q,
		db:                            dbreads.DB(q),
		disableUpdatesOnFailedRollout: cfg.DisableUpdatesOnFailedRollout,
	}
}
