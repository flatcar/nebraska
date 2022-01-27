package api_test

import (
	"log"
	"os"
	"testing"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

const (
	testServerURL    = "http://localhost:8000"
	defaultTestDbURL = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" || os.Getenv("NEBRASKA_RUN_SERVER_TESTS") == "" {
		return
	}

	if _, ok := os.LookupEnv("NEBRASKA_DB_URL"); !ok {
		log.Printf("NEBRASKA_DB_URL not set, setting to default %q\n", defaultTestDbURL)
		_ = os.Setenv("NEBRASKA_DB_URL", defaultTestDbURL)
	}

	a, err := api.New(api.OptionInitDB)
	if err != nil {
		log.Printf("Failed to init DB: %v\n", err)
		log.Println("These tests require PostgreSQL running and a tests database created, please adjust NEBRASKA_DB_URL as needed.")
		os.Exit(1)
	}
	a.Close()

	os.Exit(m.Run())
}
