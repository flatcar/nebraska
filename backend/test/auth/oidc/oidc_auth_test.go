package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/server"
)

type oidcTestSetup struct {
	nebraskaServer   interface{ Shutdown(context.Context) error }
	mockOIDCProvider *mockoidc.MockOIDC
}

func startWithOIDC(t *testing.T, configure ...func(*config.Config)) oidcTestSetup {
	// establish db connection
	db := newDBForTest(t)

	// setup and run mock OIDC provider
	mockOIDCProvider := newOIDCMockServer(t)
	startOIDCMockServer(t, mockOIDCProvider)

	localConf := *conf
	for _, fn := range configure {
		fn(&localConf)
	}

	// start nebraska server
	nebraskaServer, err := server.New(&localConf, db)
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

func signedAccessToken(t *testing.T, provider *mockoidc.MockOIDC, audience []string) string {
	t.Helper()
	return signedTokenWithClaims(t, provider, func(c jwt.MapClaims) { c["aud"] = audience })
}

// signedTokenWithClaims builds a token with the expected audience, then lets the
// caller mutate the claims to model a specific provider's token shape.
func signedTokenWithClaims(t *testing.T, provider *mockoidc.MockOIDC, mutate func(jwt.MapClaims)) string {
	t.Helper()

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":    issuerURL,
		"sub":    "oidc-test-user",
		"aud":    []string{audienceID},
		"iat":    now.Unix(),
		"nbf":    now.Unix(),
		"exp":    now.Add(5 * time.Minute).Unix(),
		"groups": []string{"nebraska-member"},
	}
	mutate(claims)

	token, err := provider.Keypair.SignJWT(claims)
	require.NoError(t, err)
	return token
}

func requestWithToken(t *testing.T, path, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, testServerURL+path, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestOIDCAudienceAuthorization(t *testing.T) {
	t.Run("authorize_with_expected_audience", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		resp := requestWithToken(t, "/api/apps", signedAccessToken(t, setup.mockOIDCProvider, []string{audienceID}))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("reject_wrong_audience", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		resp := requestWithToken(t, "/api/apps", signedAccessToken(t, setup.mockOIDCProvider, []string{"other-api"}))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("reject_missing_audience", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		resp := requestWithToken(t, "/api/apps", signedAccessToken(t, setup.mockOIDCProvider, nil))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorize_when_one_of_multiple_audiences_matches", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		resp := requestWithToken(t, "/api/apps", signedAccessToken(t, setup.mockOIDCProvider, []string{"account", audienceID}))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("reject_token_issued_for_the_frontend_client", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		resp := requestWithToken(t, "/api/apps", signedAccessToken(t, setup.mockOIDCProvider, []string{clientID}))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// An operator who maps the API audience onto ID tokens as well as access
	// tokens produces an ID token that passes audience validation. Only the
	// token type check rejects it.
	t.Run("reject_id_token_carrying_the_api_audience", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		token := signedTokenWithClaims(t, setup.mockOIDCProvider, func(c jwt.MapClaims) {
			c["aud"] = []string{clientID, audienceID}
			c["typ"] = "ID"
		})

		resp := requestWithToken(t, "/api/apps", token)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Providers that mark access tokens the Keycloak way must keep working.
	t.Run("authorize_token_with_bearer_typ_claim", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		token := signedTokenWithClaims(t, setup.mockOIDCProvider, func(c jwt.MapClaims) {
			c["typ"] = "Bearer"
		})

		resp := requestWithToken(t, "/api/apps", token)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("skip_audience_check_restores_old_behavior", func(t *testing.T) {
		setup := startWithOIDC(t, func(c *config.Config) {
			c.OidcSkipAudienceCheck = true
		})
		defer setup.shutdown()

		resp := requestWithToken(t, "/api/apps", signedAccessToken(t, setup.mockOIDCProvider, []string{"other-api"}))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("validate_token_with_expected_audience", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		resp := requestWithToken(t, "/login/validate_token", signedAccessToken(t, setup.mockOIDCProvider, []string{audienceID}))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("validate_token_rejects_wrong_audience", func(t *testing.T) {
		setup := startWithOIDC(t)
		defer setup.shutdown()

		resp := requestWithToken(t, "/login/validate_token", signedAccessToken(t, setup.mockOIDCProvider, []string{"other-api"}))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
