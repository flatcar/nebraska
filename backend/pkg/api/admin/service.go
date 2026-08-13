// Package admin provides write operations for admin-managed tables
// (application, channel, package, groups, team, users).
package admin

import (
	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
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

// NewService creates a new admin Service that writes over the given connection
// and reuses the given read queries.
// conn must be the same connection used to construct q.
func NewService(conn *dbconn.Conn, q *dbreads.Queries) *Service {
	return &Service{
		Queries: q,
		db:      dbconn.DB(conn),
	}
}
