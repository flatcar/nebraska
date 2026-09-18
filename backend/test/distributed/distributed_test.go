package distributed_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

const (
	seededAppID   = types.FlatcarAppID
	seededGroupID = "9a2deb70-37be-4026-853f-bfdd6b347bbe"
	unknownID     = "00000000-0000-0000-0000-000000000000"
)

// Migrations and role provisioning run on first start, so a node can take a
// while to serve after its container starts.
const nodeReadyTimeout = 90 * time.Second

var nodeURLEnv = []string{
	"NEBRASKA_TEST_SINGLE_URL",
	"NEBRASKA_TEST_CONTROL_URL",
	"NEBRASKA_TEST_EDGE_URL",
}

var defaultTestEnv = map[string]string{
	"NEBRASKA_TEST_SINGLE_URL":     "http://localhost:8012",
	"NEBRASKA_TEST_CONTROL_URL":    "http://localhost:8013",
	"NEBRASKA_TEST_EDGE_URL":       "http://localhost:8014",
	"NEBRASKA_TEST_CONTROL_DB_URL": "postgres://postgres:nebraska@localhost:8011/nebraska_control?sslmode=disable",
	"NEBRASKA_TEST_EDGE_DB_URL":    "postgres://postgres:nebraska@localhost:8011/nebraska_edge?sslmode=disable",
}

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" || os.Getenv("NEBRASKA_RUN_DISTRIBUTED_TESTS") == "" {
		return
	}

	for env, value := range defaultTestEnv {
		if _, ok := os.LookupEnv(env); !ok {
			log.Printf("%s not set, setting to default %q\n", env, value)
			_ = os.Setenv(env, value)
		}
	}

	for _, env := range nodeURLEnv {
		if err := waitNodeReady(os.Getenv(env), nodeReadyTimeout); err != nil {
			log.Printf("%s never served /health: %v\n", env, err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}
