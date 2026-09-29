// Package dbconn owns the database connection shared by the api stack.
package dbconn

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// PasswordFunc returns the password for a connection authenticating as user. It
// runs before every physical connection, so the password it returns is allowed
// to expire.
type PasswordFunc func(ctx context.Context, user string) (string, error)

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
// pool limits. When password is non-nil it supplies the credential for every
// physical connection, in place of the one carried in url.
func Open(driver string, url string, pool PoolConfig, password PasswordFunc) (*Conn, error) {
	db, err := open(driver, url, password)
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

// open builds the handle. The plain path leaves the URL to pgx, which re-reads
// it for each connection; the provider path asks for the credential instead.
func open(driver string, url string, password PasswordFunc) (*sqlx.DB, error) {
	if password == nil {
		return sqlx.Open(driver, url)
	}

	config, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDB(*config, stdlib.OptionBeforeConnect(func(ctx context.Context, connConfig *pgx.ConnConfig) error {
		// pgx starts connect_timeout after this callback returns, so without this
		// the credential call is the one step of opening a connection that nothing
		// bounds.
		if connConfig.ConnectTimeout > 0 {
			var cancel context.CancelFunc

			ctx, cancel = context.WithTimeout(ctx, connConfig.ConnectTimeout)
			defer cancel()
		}

		value, err := password(ctx, connConfig.User)
		if err != nil {
			return err
		}

		connConfig.Password = value

		return nil
	}))

	return sqlx.NewDb(db, driver), nil
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
