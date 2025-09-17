package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/server"
)

type oidcTestSetup struct {
	nebraskaServer   interface{ Shutdown(context.Context) error }
	mockOIDCProvider interface{ Shutdown() error }
}

func startWithOIDC(t *testing.T) oidcTestSetup {
	// establish db connection
	db := newDBForTest(t)

	// setup and run mock OIDC provider
	mockOIDCProvider := newOIDCMockServer(t)
	startOIDCMockServer(t, mockOIDCProvider)

	// start nebraska server
	nebraskaServer, err := server.New(conf, db)
	require.NotNil(t, nebraskaServer)
	require.NoError(t, err)

	//nolint:errcheck
	go nebraskaServer.Start(serverPortStr)

	_, err = waitServerReady()
	require.NoError(t, err)

	return oidcTestSetup{
		nebraskaServer:   nebraskaServer,
		mockOIDCProvider: mockOIDCProvider,
	}
}

func (s oidcTestSetup) shutdown() {
	_ = s.nebraskaServer.Shutdown(context.Background())
	_ = s.mockOIDCProvider.Shutdown()
}

func TestOIDCAuthorization(t *testing.T) {
	t.Run("authorize_with_invalid_token", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		// Try to access API with invalid token
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/apps", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorize_without_token", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		// Try to access API without authorization header
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/apps", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 401 Unauthorized - auth middleware rejects requests without Bearer token
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorize_with_malformed_header", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		// Try to access API with malformed authorization header
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/apps", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Authorization", "Malformed header")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 401 Unauthorized - auth middleware rejects malformed Bearer tokens
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestOIDCValidateTokenEndpoint(t *testing.T) {
	t.Run("validate_token_without_header", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		// Call validate token endpoint without Authorization header
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login/validate_token", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("validate_token_with_invalid_token", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		// Call validate token endpoint with invalid token
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login/validate_token", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
