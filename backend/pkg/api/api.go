package api

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strconv"

	//register "pgx" sql driver
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
	"github.com/flatcar/nebraska/backend/pkg/logger"

	// PostgreSQL Driver and Toolkit
	_ "github.com/jackc/pgx/v5/stdlib"

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

var l = logger.New("api")

const migrationsTable = "database_migrations"

// API represents an api instance used to interact with Nebraska entities.
type API struct {
	conn            *dbconn.Conn
	dbDriver        string
	dbURL           string
	migrationsDBURL string

	*dbreads.Queries
}

// db returns the handle for the statements api runs itself.
func (api *API) db() *sqlx.DB {
	return dbconn.DB(api.conn)
}

// withMigrationsDB runs fn against the connection that owns the schema. When
// NEBRASKA_MIGRATIONS_DB_URL is unset the serving connection is used, which is
// what a single-instance deployment does today.
func (api *API) withMigrationsDB(fn func(*sqlx.DB) error) error {
	if api.migrationsDBURL == "" {
		return fn(api.db())
	}

	conn, err := dbconn.Open(api.dbDriver, api.migrationsDBURL, dbconn.PoolConfig{
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})
	if err != nil {
		return fmt.Errorf("opening the migrations database connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	migrationsDB := dbconn.DB(conn)

	if err := checkSameDeployment(api.db(), migrationsDB); err != nil {
		return err
	}

	return fn(migrationsDB)
}

// New creates a new API instance, creates the underlying db connection.
func New(options ...func(*API) error) (*API, error) {
	api := &API{
		dbDriver:        "pgx",
		dbURL:           os.Getenv("NEBRASKA_DB_URL"),
		migrationsDBURL: os.Getenv("NEBRASKA_MIGRATIONS_DB_URL"),
	}

	if api.dbURL == "" {
		api.dbURL = defaultDbURL
	}

	var err error

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

	api.conn, err = dbconn.Open(api.dbDriver, api.dbURL, dbconn.PoolConfig{
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: time.Duration(connMaxLifetime) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// Load max floors per response configuration
	maxFloorsPerResponse, err := strconv.Atoi(os.Getenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE"))
	if err != nil || maxFloorsPerResponse <= 0 {
		maxFloorsPerResponse = dbreads.DefaultMaxFloorsPerResponse
	}

	api.Queries = dbreads.New(api.conn, maxFloorsPerResponse)

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

	err = api.withMigrationsDB(func(db *sqlx.DB) error {
		migrate.SetTable(migrationsTable)
		if _, err := migrate.Exec(db.DB, "postgres", migrationAssets(), migrate.Up); err != nil {
			return fmt.Errorf("applying database migrations: %w", err)
		}
		return api.setupServingRoles(db)
	})
	if err != nil {
		return nil, err
	}

	api.UpdateCachedGroups()
	api.ClearCachedAppIDs()

	return api, nil
}

type migration struct {
	ID        string    `db:"id"`
	AppliedAt time.Time `db:"applied_at"`
}

func (api *API) MigrateDown(version string) (int, error) {
	var count int

	err := api.withMigrationsDB(func(db *sqlx.DB) error {
		var err error
		count, err = migrateDown(db, version)
		return err
	})

	return count, err
}

func migrateDown(db *sqlx.DB, version string) (int, error) {
	migrate.SetTable(migrationsTable)
	migrations := migrationAssets()

	// find version based on input string
	query, _, err := goqu.Select("*").From(migrationsTable).Where(goqu.C("id").Like(fmt.Sprintf("%s%%", version))).ToSQL()
	if err != nil {
		return 0, err
	}

	var mig migration
	err = db.QueryRowx(query).StructScan(&mig)
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

	err = db.QueryRowx(query).MapScan(countMap)
	if err != nil {
		return 0, err
	}

	levels := countMap["count"].(int64)
	l.Info().Msgf("migrating down %d levels", levels)
	count, err := migrate.ExecMax(db.DB, "postgres", migrations, migrate.Down, int(levels))
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

	err = api.withMigrationsDB(func(db *sqlx.DB) error {
		_, err := db.Exec(string(sqlFile))
		return err
	})
	if err != nil {
		return err
	}
	api.UpdateCachedGroups()

	return nil
}

// Close releases the connections to the database.
func (api *API) Close() {
	_ = api.conn.Close()
}

// Conn returns the shared database connection (dbconn.Conn) owned by this API
// instance.
func (api *API) Conn() *dbconn.Conn {
	return api.conn
}

// Reads returns the shared read queries (dbreads.Queries) owned by this API
// instance.
func (api *API) Reads() *dbreads.Queries {
	return api.Queries
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

	_, err = a.db().Exec(string(sqlFile))
	if err != nil {
		return nil, err
	}
	a.UpdateCachedGroups()
	a.ClearCachedAppIDs()

	return a, nil
}
