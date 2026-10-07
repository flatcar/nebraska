package api

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryProfiler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name           string
		config         QueryProfilerConfig
		expectedThreshold time.Duration
	}{
		{
			name: "default threshold",
			config: QueryProfilerConfig{
				EnableSlowQueryLog: true,
				EnableMetrics:      true,
			},
			expectedThreshold: DefaultSlowQueryThreshold,
		},
		{
			name: "custom threshold",
			config: QueryProfilerConfig{
				SlowQueryThreshold: 50 * time.Millisecond,
				EnableSlowQueryLog: true,
				EnableMetrics:      true,
			},
			expectedThreshold: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiler := NewQueryProfiler(sqlxDB, tt.config)
			assert.NotNil(t, profiler)
			assert.Equal(t, tt.expectedThreshold, profiler.slowQueryThreshold)
			assert.Equal(t, tt.config.EnableSlowQueryLog, profiler.enableSlowQueryLog)
			assert.Equal(t, tt.config.EnableMetrics, profiler.enableMetrics)
		})
	}
}

func TestQueryProfiler_Queryx(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 10 * time.Millisecond,
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test1").
		AddRow(2, "test2")

	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	result, err := profiler.Queryx("SELECT * FROM users WHERE active = ?", true)
	require.NoError(t, err)
	assert.NotNil(t, result)
	defer result.Close()

	count := 0
	for result.Next() {
		count++
	}
	assert.Equal(t, 2, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryProfiler_QueryRowx(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 10 * time.Millisecond,
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test")
	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").WillReturnRows(rows)

	row := profiler.QueryRowx("SELECT * FROM users WHERE id = ?", 1)
	assert.NotNil(t, row)

	var id int
	var name string
	err = row.Scan(&id, &name)
	require.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, "test", name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryProfiler_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 10 * time.Millisecond,
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	mock.ExpectExec("UPDATE users SET name = ?").
		WithArgs("updated").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := profiler.Exec("UPDATE users SET name = ?", "updated")
	require.NoError(t, err)
	assert.NotNil(t, result)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryProfiler_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 10 * time.Millisecond,
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test")
	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").WillReturnRows(rows)

	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var user User
	err = profiler.Get(&user, "SELECT * FROM users WHERE id = ?", 1)
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryProfiler_Select(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 10 * time.Millisecond,
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test1").
		AddRow(2, "test2")
	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var users []User
	err = profiler.Select(&users, "SELECT * FROM users")
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "test1", users[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryProfiler_SlowQueryDetection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Set a very low threshold to ensure we detect "slow" queries
	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 1 * time.Nanosecond, // Extremely low threshold
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT (.+) FROM users").
		WillDelayFor(2 * time.Millisecond). // Ensure some delay
		WillReturnRows(rows)

	// This should be detected as slow
	result, err := profiler.Queryx("SELECT * FROM users")
	require.NoError(t, err)
	assert.NotNil(t, result)
	result.Close()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryProfiler_ErrorHandling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 10 * time.Millisecond,
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	mock.ExpectQuery("SELECT (.+) FROM users").
		WillReturnError(sql.ErrNoRows)

	_, err = profiler.Queryx("SELECT * FROM users")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryProfiler_SetSlowQueryThreshold(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 100 * time.Millisecond,
	})

	assert.Equal(t, 100*time.Millisecond, profiler.GetSlowQueryThreshold())

	profiler.SetSlowQueryThreshold(200 * time.Millisecond)
	assert.Equal(t, 200*time.Millisecond, profiler.GetSlowQueryThreshold())
}

func TestQueryProfiler_Stats(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{})

	stats := profiler.Stats()
	assert.NotNil(t, stats)
	// Just verify we can call Stats without error
}

func TestQueryProfiler_Transactions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{})

	mock.ExpectBegin()
	tx, err := profiler.Beginx()
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryStats_String(t *testing.T) {
	stats := &QueryStats{
		QueryType:    "select",
		Count:        100,
		TotalTime:    5 * time.Second,
		AverageTime:  50 * time.Millisecond,
		MaxTime:      200 * time.Millisecond,
		MinTime:      10 * time.Millisecond,
		SlowCount:    5,
		ErrorCount:   2,
	}

	str := stats.String()
	assert.Contains(t, str, "select")
	assert.Contains(t, str, "100")
	assert.Contains(t, str, "5s")
}

func TestRegisterMetrics(t *testing.T) {
	// This test verifies that RegisterMetrics can be called multiple times
	// without error due to already registered collectors
	err := RegisterMetrics()
	assert.NoError(t, err)

	err = RegisterMetrics()
	assert.NoError(t, err) // Should not error on re-registration
}

func TestQueryProfiler_NamedExec(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	profiler := NewQueryProfiler(sqlxDB, QueryProfilerConfig{
		SlowQueryThreshold: 10 * time.Millisecond,
		EnableSlowQueryLog: true,
		EnableMetrics:      true,
	})

	mock.ExpectExec("UPDATE users SET name = ?").
		WithArgs("test").
		WillReturnResult(sqlmock.NewResult(0, 1))

	type Update struct {
		Name string `db:"name"`
	}

	result, err := profiler.NamedExec("UPDATE users SET name = :name", Update{Name: "test"})
	require.NoError(t, err)
	assert.NotNil(t, result)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}
