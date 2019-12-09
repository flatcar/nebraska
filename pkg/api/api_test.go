package api

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testsDbURL string = "postgres://postgres@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"
)

func newForTest(t *testing.T) *API {
	values := map[string]bool{
		"1":     true,
		"t":     true,
		"true":  true,
		"0":     false,
		"f":     false,
		"false": false,
	}
	useGORM := false
	if v, ok := values[os.Getenv("USE_GORM")]; ok {
		useGORM = v
	}
	if useGORM {
		return newForTestGORM(t)
	}
	return newForTestDat(t)
}

func newForTestDat(t *testing.T) *API {
	return newForTestWithOpts(t, OptionInitDB, OptionDisableUpdatesOnFailedRollout)
}

func newForTestGORM(t *testing.T) *API {
	return newForTestWithOpts(t, OptionInitDB, OptionDisableUpdatesOnFailedRollout, ConnectWithGORM)
}

func newForTestWithOpts(t *testing.T, options ...func(*API) error) *API {
	a, err := NewForTest(options...)

	require.NoError(t, err)
	require.NotNil(t, a)

	return a
}

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	_ = os.Setenv("NEBRASKA_DB_URL", testsDbURL)

	a, err := New(OptionInitDB)
	if err != nil {
		log.Println("These tests require PostgreSQL running and a tests database created, please adjust testsDbUrl as needed.")
		log.Println("Default: postgres://postgres@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10")
		os.Exit(1)
	}
	a.Close()

	os.Exit(m.Run())
}
