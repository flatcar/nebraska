package auth_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/config"
)

const (
	testServerURL    = "http://localhost:6000"
	defaultTestDbURL = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"
	clientID         = "clientID"
	clientSecret     = "clientSecret"
	issuerURL        = "http://127.0.0.1:8080/oidc"
	serverPort       = uint(6000)
)

var serverPortStr = fmt.Sprintf(":%d", serverPort)

var conf = &config.Config{
	EnableSyncer:    true,
	NebraskaURL:     testServerURL,
	HTTPLog:         true,
	AuthMode:        "oidc",
	Debug:           true,
	ServerPort:      serverPort,
	OidcClientID:    clientID,
	OidcIssuerURL:   issuerURL,
	OidcAdminRoles:  "nebraska-admin",
	OidcViewerRoles: "nebraska-member",
	OidcRolesPath:   "groups",
	OidcScopes:      "openid,profile,email,groups",
}

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
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
