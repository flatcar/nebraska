package api

import (
	"errors"
	"os"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"

	// Postgresql driver
	_ "github.com/lib/pq"
)

// To re-generate the bindata.go file, use go-bindata from
// github.com/kevinburke/go-bindata (a fork of the discontinued
// go-bindata project). Run the following command from the root of the
// repository:
//
//    make bindata

//go:generate go-bindata -ignore=\.swp -pkg=api -modtime=1 db db/migrations

const (
	defaultDbURL = "postgres://postgres@127.0.0.1:5432/nebraska?sslmode=disable&connect_timeout=10"
	nowUTC       = dat.UnsafeString("now() at time zone 'utc'")
)

var (
	// ErrNoRowsAffected indicates that no rows were affected in an update or
	// delete database operation.
	ErrNoRowsAffected = errors.New("nebraska: no rows affected")

	// ErrInvalidSemver indicates that the provided semver version is not valid.
	ErrInvalidSemver = errors.New("nebraska: invalid semver")
)

// API represents an api instance used to interact with Nebraska entities.
type API struct {
	db       *sqlx.DB
	dbR      *runner.DB
	dbDriver string
	dbURL    string

	// disableUpdatesOnFailedRollout defines wether to disable updates
	// after a first rollout attempt failed (ResultFailed)
	disableUpdatesOnFailedRollout bool
}

// New creates a new API instance, creating the underlying db connection and
// applying db migrations available.
func New(options ...func(*API) error) (*API, error) {
	api := &API{
		dbDriver: "postgres",
		dbURL:    os.Getenv("NEBRASKA_DB_URL"),
	}

	if api.dbURL == "" {
		api.dbURL = defaultDbURL
	}

	var err error
	api.db, err = sqlx.Open(api.dbDriver, api.dbURL)
	if err != nil {
		return nil, err
	}
	if err := api.db.Ping(); err != nil {
		return nil, err
	}

	dat.EnableInterpolation = true
	api.dbR = runner.NewDBFromSqlx(api.db)

	for _, option := range options {
		err := option(api)
		if err != nil {
			return nil, err
		}
	}

	migrate.SetTable("database_migrations")
	migrations := &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "db/migrations",
	}
	if _, err := migrate.Exec(api.db.DB, "postgres", migrations, migrate.Up); err != nil {
		return nil, err
	}

	return api, nil
}

// OptionInitDB will initialize the database during the API instance creation,
// dropping all existing tables, which will force all migration scripts to be
// re-executed. Use with caution, this will DESTROY ALL YOUR DATA.
func OptionInitDB(api *API) error {
	sqlFile, err := Asset("db/drop_all_tables.sql")
	if err != nil {
		return err
	}

	if _, err := api.db.Exec(string(sqlFile)); err != nil {
		return err
	}

	return nil
}

// OptionDisableUpdatesOnFailedRollout will modify API to disable
// updates on failed rollout.
func OptionDisableUpdatesOnFailedRollout(api *API) error {
	api.disableUpdatesOnFailedRollout = true

	return nil
}

// Close releases the connections to the database.
func (api *API) Close() {
	_ = api.db.DB.Close()
}

// NewForTest creates a new API instance with given options and fills
// the database with sample data for testing purposes.
func NewForTest(options ...func(*API) error) (*API, error) {
	a, err := New(options...)
	if err != nil {
		return nil, err
	}

	sqlFile, err := Asset("db/sample_data.sql")
	if err != nil {
		return nil, err
	}

	_, err = a.db.Exec(string(sqlFile))
	if err != nil {
		return nil, err
	}

	return a, nil
}
