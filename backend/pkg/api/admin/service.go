// Package admin provides write operations for admin-managed tables
// (application, channel, package, groups, team, users).
package admin

import (
	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

var l = logger.New("admin")

// Service provides admin write operations. It embeds the api's shared
// dbreads.Queries for read access using the same DB connection.
type Service struct {
	*dbreads.Queries
	db *sqlx.DB
}

// NewService creates a new admin Service that reuses the given read queries
// (and their underlying DB connection).
func NewService(q *dbreads.Queries) *Service {
	return &Service{
		Queries: q,
		db:      dbreads.DB(q),
	}
}
