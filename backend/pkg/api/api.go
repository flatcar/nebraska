package api

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"strconv"

	//register "pgx" sql driver
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/flatcar/nebraska/backend/pkg/logger"

	// PostgreSQL Driver and Toolkit
	_ "github.com/jackc/pgx/v4/stdlib"

	// Postgresql driver
	_ "github.com/lib/pq"

	"time"
)

var (
	//go:embed db/*.sql
	sqlFolder embed.FS
	//go:embed db/migrations/*.sql
	migrationsFolder embed.FS
)

const (
	defaultDbURL          = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska?sslmode=disable&connect_timeout=10"
	maxOpenAndIdleDbConns = 25
	dBConnMaxLifetime     = 5 * 60 // seconds
)

func nowUTC() time.Time {
	return time.Now().UTC()
}

var (
	l = logger.New("api")

	// ErrNoRowsAffected indicates that no rows were affected in an update or
	// delete database operation.
	ErrNoRowsAffected = errors.New("nebraska: no rows affected")

	// ErrInvalidSemver indicates that the provided semver version is not valid.
	ErrInvalidSemver = errors.New("nebraska: invalid semver")

	// ErrInvalidArch indicates that the provided architecture is not valid/supported
	ErrInvalidArch = errors.New("nebraska: invalid/unsupported arch")

	// ErrArchMismatch indicates that arches of two objects didn't
	// match (for example, for a package and channel)
	ErrArchMismatch = errors.New("nebraska: mismatched arches")
)

const migrationsTable = "database_migrations"

// API represents an api instance used to interact with Nebraska entities.
type API struct {
	db       *sqlx.DB
	dbDriver string
	dbURL    string

	// disableUpdatesOnFailedRollout defines wether to disable updates
	// after a first rollout attempt failed (ResultFailed)
	disableUpdatesOnFailedRollout bool

	// maxFloorsPerResponse defines the maximum number of floor versions
	// to return in a single update response to syncers
	maxFloorsPerResponse int
}

// New creates a new API instance, creates the underlying db connection.
func New(options ...func(*API) error) (*API, error) {
	api := &API{
		dbDriver: "pgx",
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

	var (
		maxOpenConns    int
		maxIdleConns    int
		connMaxLifetime int
	)

	maxOpenConns, err = strconv.Atoi(os.Getenv("NEBRASKA_DB_MAX_OPEN_CONNS"))
	if err != nil {
		maxOpenConns = maxOpenAndIdleDbConns
	}
	maxIdleConns, err = strconv.Atoi(os.Getenv("NEBRASKA_DB_MAX_IDLE_CONNS"))
	if err != nil {
		maxIdleConns = maxOpenConns
	}

	connMaxLifetime, err = strconv.Atoi(os.Getenv("NEBRASKA_DB_CONN_MAX_LIFETIME"))
	if err != nil {
		connMaxLifetime = dBConnMaxLifetime
	}

	api.db.SetMaxOpenConns(maxOpenConns)
	api.db.SetMaxIdleConns(maxIdleConns)
	api.db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	// Load max floors per response configuration
	maxFloorsPerResponse, err := strconv.Atoi(os.Getenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE"))
	if err != nil || maxFloorsPerResponse <= 0 {
		api.maxFloorsPerResponse = 5
	} else {
		api.maxFloorsPerResponse = maxFloorsPerResponse
	}

	for _, option := range options {
		err := option(api)
		if err != nil {
			return nil, err
		}
	}
	return api, nil
}

// NewWithMigrations creates a new API instance, creates the underlying db connection and
// applies all available db migrations.
func NewWithMigrations(options ...func(*API) error) (*API, error) {
	api, err := New(options...)
	if err != nil {
		return nil, err
	}

	migrate.SetTable(migrationsTable)
	migrations := migrationAssets()

	if _, err := migrate.Exec(api.db.DB, "postgres", migrations, migrate.Up); err != nil {
		return nil, err
	}
	api.updateCachedGroups()
	api.clearCachedAppIDs()

	return api, nil
}

type migration struct {
	ID        string    `db:"id"`
	AppliedAt time.Time `db:"applied_at"`
}

func (api *API) MigrateDown(version string) (int, error) {
	migrate.SetTable(migrationsTable)
	migrations := migrationAssets()

	// find version based on input string
	query, _, err := goqu.Select("*").From(migrationsTable).Where(goqu.C("id").Like(fmt.Sprintf("%s%%", version))).ToSQL()
	if err != nil {
		return 0, err
	}

	var mig migration
	err = api.db.QueryRowx(query).StructScan(&mig)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no migrations found for: %s, err: %v", version, err)
		}
		return 0, err
	}

	// find  count of migrations that have been applied after the version
	query, _, err = goqu.Select(goqu.COUNT("*")).From(migrationsTable).Where(goqu.C("applied_at").Gt(mig.AppliedAt)).ToSQL()
	if err != nil {
		return 0, err
	}

	countMap := make(map[string]interface{})

	err = api.db.QueryRowx(query).MapScan(countMap)
	if err != nil {
		return 0, err
	}

	levels := countMap["count"].(int64)
	l.Info().Msgf("migrating down %d levels", levels)
	count, err := migrate.ExecMax(api.db.DB, "postgres", migrations, migrate.Down, int(levels))
	if err != nil {
		return 0, err
	}
	l.Info().Msg("successfully migrated down")
	return count, nil
}

func migrationAssets() *migrate.EmbedFileSystemMigrationSource {
	return &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFolder,
		Root:       "db/migrations",
	}
}

// OptionInitDB will initialize the database during the API instance creation,
// dropping all existing tables, which will force all migration scripts to be
// re-executed. Use with caution, this will DESTROY ALL YOUR DATA.
func OptionInitDB(api *API) error {
	sqlFile, err := sqlFolder.ReadFile("db/drop_all_tables.sql")
	if err != nil {
		return err
	}

	if _, err := api.db.Exec(string(sqlFile)); err != nil {
		return err
	}
	api.updateCachedGroups()

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
	_ = api.db.Close()
}

// NewForTest creates a new API instance with given options and fills
// the database with sample data for testing purposes.
func NewForTest(options ...func(*API) error) (*API, error) {
	a, err := NewWithMigrations(options...)
	if err != nil {
		return nil, err
	}

	sqlFile, err := sqlFolder.ReadFile("db/sample_data.sql")
	if err != nil {
		return nil, err
	}

	_, err = a.db.Exec(string(sqlFile))
	if err != nil {
		return nil, err
	}
	a.updateCachedGroups()
	a.clearCachedAppIDs()

	return a, nil
}
