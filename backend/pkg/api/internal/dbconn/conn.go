// Package dbconn owns the database connection shared by the api stack.
package dbconn

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// PoolConfig holds the connection pool limits applied when the connection is opened.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Conn owns the database connection shared by the api stack.
type Conn struct {
	db *sqlx.DB
}

// Open opens the database connection, verifies it is reachable and applies the
// pool limits.
func Open(driver string, url string, pool PoolConfig) (*Conn, error) {
	db, err := sqlx.Open(driver, url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(pool.MaxOpenConns)
	db.SetMaxIdleConns(pool.MaxIdleConns)
	db.SetConnMaxLifetime(pool.ConnMaxLifetime)

	return &Conn{db: db}, nil
}

// Close releases the connection.
func (c *Conn) Close() error {
	return c.db.Close()
}

// DB returns the underlying database handle. It is a package function rather
// than a method so the handle stays unreachable outside pkg/api.
func DB(c *Conn) *sqlx.DB {
	return c.db
}
