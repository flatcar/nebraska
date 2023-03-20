package api

import (
	"log"
	"os"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTestDbURL string = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"
)

func newForTest(t *testing.T) *API {
	a, err := NewForTest(OptionInitDB, OptionDisableUpdatesOnFailedRollout)

	require.NoError(t, err)
	require.NotNil(t, a)

	return a
}

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	if _, ok := os.LookupEnv("NEBRASKA_DB_URL"); !ok {
		log.Printf("NEBRASKA_DB_URL not set, setting to default %q\n", defaultTestDbURL)
		_ = os.Setenv("NEBRASKA_DB_URL", defaultTestDbURL)
	}

	a, err := NewWithMigrations(OptionInitDB)
	if err != nil {
		log.Printf("Failed to init DB: %v\n", err)
		log.Println("These tests require PostgreSQL running and a tests database created, please adjust NEBRASKA_DB_URL as needed.")
		os.Exit(1)
	}
	a.Close()

	os.Exit(m.Run())
}

func TestMigrateDown(t *testing.T) {
	// Create New DB
	db, err := NewWithMigrations(OptionInitDB)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.MigrateDown("0004")
	require.NoError(t, err)

	query, _, err := goqu.Select("*").From(migrationsTable).ToSQL()
	require.NoError(t, err)

	var migrations []migration
	rows, err := db.db.Queryx(query)
	require.NoError(t, err)

	defer rows.Close()
	for rows.Next() {
		var mig migration
		err := rows.StructScan(&mig)
		require.NoError(t, err)
		migrations = append(migrations, mig)
	}

	assert.Equal(t, 4, len(migrations))
}
