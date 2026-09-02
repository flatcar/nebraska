package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateAuthMode verifies that Validate() rejects unknown auth-mode values.
// Before this fix, an unknown mode like "typo" silently passed validation and
// Nebraska would start with no authentication enforced.
func TestValidateAuthMode(t *testing.T) {
	tests := []struct {
		name      string
		authMode  string
		wantError bool
		errMsg    string
	}{
		{
			name:      "noop is valid",
			authMode:  "noop",
			wantError: false,
		},
		{
			name:      "unknown mode returns error",
			authMode:  "typo",
			wantError: true,
			errMsg:    "invalid auth-mode",
		},
		{
			name:      "empty string returns error",
			authMode:  "",
			wantError: true,
			errMsg:    "invalid auth-mode",
		},
		{
			name:      "uppercase NOOP is rejected",
			authMode:  "NOOP",
			wantError: true,
			errMsg:    "invalid auth-mode",
		},
		{
			name:      "github with missing fields returns error",
			authMode:  "github",
			wantError: true,
			errMsg:    "invalid github configuration",
		},
		{
			name:      "oidc with missing fields returns error",
			authMode:  "oidc",
			wantError: true,
			errMsg:    "invalid OIDC configuration",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{AuthMode: tc.authMode}
			err := c.Validate()
			if tc.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateGithubMode verifies all required fields for GitHub auth mode.
func TestValidateGithubMode(t *testing.T) {
	validConfig := &Config{
		AuthMode:         "github",
		GhClientID:       "client-id",
		GhClientSecret:   "client-secret",
		GhReadOnlyTeams:  "org/ro-team",
		GhReadWriteTeams: "org/rw-team",
	}

	// All fields present - should pass
	require.NoError(t, validConfig.Validate())

	// Missing one field at a time - should all fail
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"missing GhClientID", func(c *Config) { c.GhClientID = "" }},
		{"missing GhClientSecret", func(c *Config) { c.GhClientSecret = "" }},
		{"missing GhReadOnlyTeams", func(c *Config) { c.GhReadOnlyTeams = "" }},
		{"missing GhReadWriteTeams", func(c *Config) { c.GhReadWriteTeams = "" }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := *validConfig
			tc.mutate(&c)
			err := c.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid github configuration")
		})
	}
}

// TestValidateOIDCMode verifies all required fields for OIDC auth mode.
func TestValidateOIDCMode(t *testing.T) {
	validConfig := &Config{
		AuthMode:       "oidc",
		OidcClientID:   "client-id",
		OidcIssuerURL:  "https://issuer.example.com",
		OidcAdminRoles: "admin",
		OidcViewerRoles: "viewer",
	}

	// All fields present - should pass
	require.NoError(t, validConfig.Validate())

	// Missing one field at a time - should all fail
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"missing OidcClientID", func(c *Config) { c.OidcClientID = "" }},
		{"missing OidcIssuerURL", func(c *Config) { c.OidcIssuerURL = "" }},
		{"missing OidcAdminRoles", func(c *Config) { c.OidcAdminRoles = "" }},
		{"missing OidcViewerRoles", func(c *Config) { c.OidcViewerRoles = "" }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := *validConfig
			tc.mutate(&c)
			err := c.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid OIDC configuration")
		})
	}
}

// TestValidateCAFile verifies the CA file validation.
func TestValidateCAFile(t *testing.T) {
	// Empty CA file - should pass (optional field)
	c := &Config{AuthMode: "noop", CAFile: ""}
	require.NoError(t, c.Validate())

	// Non-existent CA file - should fail
	c = &Config{AuthMode: "noop", CAFile: "/nonexistent/ca.pem"}
	err := c.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ca-file")

	// Valid CA file - should pass
	tmpFile, err := os.CreateTemp("", "ca-*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	c = &Config{AuthMode: "noop", CAFile: tmpFile.Name()}
	require.NoError(t, c.Validate())
}

// TestValidateNoopMode verifies noop auth mode passes without any extra fields.
func TestValidateNoopMode(t *testing.T) {
	c := &Config{AuthMode: "noop"}
	require.NoError(t, c.Validate())
}
