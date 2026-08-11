// Package dbreads holds the shared read queries used by the api stack.
package dbreads

import (
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

var l = logger.New("dbreads")

// ErrNoPackageFound indicates that the group doesn't have a channel
// assigned or that the channel doesn't have a package assigned.
var ErrNoPackageFound = errors.New("nebraska: no package found")

type Queries struct {
	db *sqlx.DB

	// maxFloorsPerResponse defines the maximum number of floor versions
	// to return in a single update response to syncers
	maxFloorsPerResponse int
}

func New(conn *dbconn.Conn, maxFloorsPerResponse int) *Queries {
	if maxFloorsPerResponse <= 0 {
		maxFloorsPerResponse = DefaultMaxFloorsPerResponse
	}
	return &Queries{db: dbconn.DB(conn), maxFloorsPerResponse: maxFloorsPerResponse}
}
