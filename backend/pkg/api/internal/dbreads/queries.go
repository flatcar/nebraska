// Package dbreads holds the shared read queries used by the api stack.
package dbreads

import (
	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

var l = logger.New("dbreads")

type Queries struct {
	db *sqlx.DB
}

func New(conn *dbconn.Conn) *Queries {
	return &Queries{db: dbconn.DB(conn)}
}
