package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}
	os.Exit(m.Run())
}

// TestRolesPathExtraction pins the path extraction contract of rolesFromToken
// and rolesFromUserInfo. Both must answer identically, so every case runs
// against both.
func TestRolesPathExtraction(t *testing.T) {
	tests := []struct {
		name          string
		claims        map[string]any
		rolesPath     string
		expectExists  bool
		expectedRoles []string
	}{
		{
			name: "path exists with roles",
			claims: map[string]any{
				"groups": []any{"admin", "viewer"},
			},
			rolesPath:     "groups",
			expectExists:  true,
			expectedRoles: []string{"admin", "viewer"},
		},
		{
			name: "nested path exists with roles",
			claims: map[string]any{
				"realm_access": map[string]any{
					"roles": []any{"nebraska-admin"},
				},
			},
			rolesPath:     "realm_access.roles",
			expectExists:  true,
			expectedRoles: []string{"nebraska-admin"},
		},
		{
			name: "path exists but empty array",
			claims: map[string]any{
				"groups": []any{},
			},
			rolesPath:     "groups",
			expectExists:  true,
			expectedRoles: []string{},
		},
		{
			// A scalar must not be rejected; the pre-gjson code type-asserted
			// this to []any and panicked.
			name: "scalar value at path",
			claims: map[string]any{
				"groups": "admin",
			},
			rolesPath:     "groups",
			expectExists:  true,
			expectedRoles: []string{"admin"},
		},
		{
			name: "path does not exist",
			claims: map[string]any{
				"other": "stuff",
			},
			rolesPath:    "groups",
			expectExists: false,
		},
		{
			name: "partial nested path exists",
			claims: map[string]any{
				"realm_access": map[string]any{},
			},
			rolesPath:    "realm_access.roles",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRoles, tokenErr := rolesFromToken(idTokenWithClaims(t, tt.claims), tt.rolesPath)

			oa := &oidcAuth{provider: userInfoProvider(t, tt.claims)}
			userInfoRoles, userInfoErr := oa.rolesFromUserInfo(context.Background(), userInfoToken, tt.rolesPath)

			if !tt.expectExists {
				assert.Error(t, tokenErr, "a missing roles path must be an error")
				assert.Error(t, userInfoErr, "a missing roles path must be an error")
				return
			}

			assert.NoError(t, tokenErr)
			assert.Equal(t, tt.expectedRoles, tokenRoles)
			assert.NoError(t, userInfoErr)
			assert.Equal(t, tt.expectedRoles, userInfoRoles)
		})
	}
}

func TestDetermineAccessLevel(t *testing.T) {
	oa := &oidcAuth{
		adminRoles:  []string{"nebraska-admin", "super-admin"},
		viewerRoles: []string{"nebraska-member", "readonly"},
	}

	tests := []struct {
		name        string
		roles       []string
		expectLevel string
	}{
		{
			name:        "admin role",
			roles:       []string{"nebraska-admin"},
			expectLevel: "admin",
		},
		{
			name:        "viewer role",
			roles:       []string{"nebraska-member"},
			expectLevel: "viewer",
		},
		{
			name:        "admin takes precedence over viewer",
			roles:       []string{"nebraska-member", "nebraska-admin"},
			expectLevel: "admin",
		},
		{
			name:        "no matching roles",
			roles:       []string{"other-role"},
			expectLevel: "",
		},
		{
			name:        "empty roles",
			roles:       []string{},
			expectLevel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := oa.determineAccessLevel(tt.roles)
			assert.Equal(t, tt.expectLevel, level)
		})
	}
}

const userInfoToken = "test-access-token"

// signingKey is generated once; RSA key generation dominates the test runtime.
var signingKey = func() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}()

// idTokenWithClaims signs claims and runs them back through go-oidc's verifier,
// the only way to build a populated *oidc.IDToken from outside that package.
func idTokenWithClaims(t *testing.T, claims map[string]any) *oidc.IDToken {
	t.Helper()

	const issuer = "https://issuer.example"

	mapClaims := jwt.MapClaims{
		"iss": issuer,
		"sub": "roles-test-user",
		"aud": []string{"nebraska-api"},
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	}
	for k, v := range claims {
		mapClaims[k] = v
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims).SignedString(signingKey)
	require.NoError(t, err)

	token, err := oidc.NewVerifier(
		issuer,
		&oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{signingKey.Public()}},
		&oidc.Config{SkipClientIDCheck: true},
	).Verify(context.Background(), raw)
	require.NoError(t, err)

	return token
}

// userInfoProvider serves the discovery document go-oidc needs plus a userinfo
// endpoint returning the given claims.
func userInfoProvider(t *testing.T, userInfoClaims map[string]any) *oidc.Provider {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/keys",
			"userinfo_endpoint":      server.URL + "/userinfo",
		}))
	})

	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+userInfoToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(userInfoClaims))
	})

	provider, err := oidc.NewProvider(context.Background(), server.URL)
	require.NoError(t, err)

	return provider
}
