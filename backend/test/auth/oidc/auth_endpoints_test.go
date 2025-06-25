package auth_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/server"
)

func TestOIDCAuthModeSetup(t *testing.T) {
	t.Run("oidc_server_not_reachable_on_server_setup", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		server, err := server.New(conf, db)
		assert.Nil(t, server)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error setting up oidc provider")
		assert.Contains(t, err.Error(), "connect: connection refused")
	})

	t.Run("invalid_oidc_server_url", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		var testConfig config.Config
		err := copier.Copy(&testConfig, conf)
		require.NoError(t, err)

		testConfig.OidcIssuerURL = "http://127.0.0.1:8080/"

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		server, err := server.New(&testConfig, db)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "error setting up oidc provider")
		assert.Contains(t, err.Error(), "404 page not found")

		err = oidcServer.Shutdown()
		require.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		server, err := server.New(conf, db)
		assert.NotNil(t, server)
		assert.NoError(t, err)

		err = oidcServer.Shutdown()
		require.NoError(t, err)
	})
}

var ErrOutOfRetries = errors.New("test: out of retries")

func waitServerReady() (bool, error) {
	retries := 5
	for i := 0; i < retries; i++ {
		if i != 0 {
			time.Sleep(100 * time.Millisecond)
		}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/health", testServerURL), nil)
		if err != nil {
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		if (http.StatusOK == resp.StatusCode) && ("OK" == string(bodyBytes)) {
			return true, nil
		}
	}
	return false, ErrOutOfRetries
}
func TestOIDCEndpointBehavior(t *testing.T) {
	t.Run("login_routes_fall_back_to_spa", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		// Test /login endpoint - not registered in OIDC mode, falls back to SPA
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Returns 200 because it serves the SPA frontend (index.html)
		// In OIDC mode, frontend handles authentication directly with OIDC provider
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Test /login/cb endpoint - not registered in OIDC mode, falls back to SPA  
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/login/cb", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Returns 200 because it serves the SPA frontend (index.html)
		// In OIDC mode, frontend handles OAuth callback directly with OIDC provider
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestValidateTokenEndpoint(t *testing.T) {
	t.Run("validate_token_with_invalid_token", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login/validate_token", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		// Test with an invalid JWT token
		req.Header.Set("Authorization", "Bearer invalid-jwt-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 401 Unauthorized for invalid JWT token
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
