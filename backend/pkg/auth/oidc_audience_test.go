package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOIDCAudienceConfig(t *testing.T) {
	tests := []struct {
		name     string
		audience string
		skip     bool
		wantErr  bool
	}{
		{name: "distinct API audience", audience: "https://nebraska-api"},
		{name: "missing audience is rejected", wantErr: true},
		{name: "skip permits legacy configuration", skip: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOIDCAudienceConfig(tt.audience, tt.skip)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// The audience is rejected before the provider is contacted, so this needs no
// network. The message is asserted because an unreachable issuer would
// otherwise satisfy a bare error check.
func TestNewOIDCAuthenticatorRequiresAudience(t *testing.T) {
	_, err := NewOIDCAuthenticator(&OIDCAuthConfig{IssuerURL: "https://issuer.example"})
	assert.ErrorContains(t, err, "no access token audience configured",
		"OIDC mode must refuse to start without an access token audience")
}

func makeJWT(t *testing.T, header, payload map[string]any) string {
	t.Helper()
	enc := func(v map[string]any) string {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return enc(header) + "." + enc(payload) + ".signature"
}

func TestVerifyAccessTokenType(t *testing.T) {
	const apiAudience = "nebraska-api"

	tests := []struct {
		name    string
		header  map[string]any
		payload map[string]any
		wantErr bool
	}{
		{
			name:    "RFC 9068 access token",
			header:  map[string]any{"alg": "RS256", "typ": "at+jwt"},
			payload: map[string]any{"aud": apiAudience},
		},
		{
			name:    "Keycloak access token",
			header:  map[string]any{"alg": "RS256", "typ": "JWT"},
			payload: map[string]any{"aud": apiAudience, "typ": "Bearer"},
		},
		{
			// The case audience validation alone cannot catch: an operator maps
			// the API audience onto ID tokens as well as access tokens.
			name:    "Keycloak ID token carrying the API audience is rejected",
			header:  map[string]any{"alg": "RS256", "typ": "JWT"},
			payload: map[string]any{"aud": apiAudience, "typ": "ID"},
			wantErr: true,
		},
		{
			// Azure AD, Okta and Auth0 emit no token kind marker at all.
			name:    "provider without any token type marker is accepted",
			header:  map[string]any{"alg": "RS256", "typ": "JWT"},
			payload: map[string]any{"aud": apiAudience},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyAccessTokenType(makeJWT(t, tt.header, tt.payload))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyAccessTokenTypeMalformed(t *testing.T) {
	for _, raw := range []string{"", "not-a-jwt", "only.two", "a.b.c.d", "a.b.c", "!!!.eyJ9.sig"} {
		assert.Error(t, verifyAccessTokenType(raw), "expected %q to be rejected", raw)
	}
}
