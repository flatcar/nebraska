package runtime

import (
	"os"
	"testing"
)

const defaultTestDbURL = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}
	if _, ok := os.LookupEnv("NEBRASKA_DB_URL"); !ok {
		_ = os.Setenv("NEBRASKA_DB_URL", defaultTestDbURL)
	}
	os.Exit(m.Run())
}
