package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tmpDir := t.TempDir()

	validCAFile := filepath.Join(tmpDir, "ca.pem")
	require.NoError(t, os.WriteFile(validCAFile, []byte("dummy-cert"), 0o600))

	tests := []struct {
		name      string
		cfg       Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid config with syncer disabled and no auth mode",
			cfg:  Config{HostFlatcarPackages: false},
		},
		{
			name:      "host flatcar packages enabled but no path provided",
			cfg:       Config{HostFlatcarPackages: true, FlatcarPackagesPath: ""},
			wantErr:   true,
			errSubstr: "invalid Flatcar packages path",
		},
		{
			name: "host flatcar packages enabled with invalid path",
			cfg: Config{
				HostFlatcarPackages: true,
				FlatcarPackagesPath: filepath.Join(tmpDir, "does-not-exist"),
				NebraskaURL:         "http://localhost:8000",
			},
			wantErr:   true,
			errSubstr: "invalid Flatcar packages path",
		},
		{
			name: "host flatcar packages enabled with valid path but invalid nebraska URL",
			cfg: Config{
				HostFlatcarPackages: true,
				FlatcarPackagesPath: tmpDir,
				NebraskaURL:         "",
			},
			wantErr:   true,
			errSubstr: "invalid Nebraska URL",
		},
		{
			name: "host flatcar packages enabled with valid path and valid nebraska URL",
			cfg: Config{
				HostFlatcarPackages: true,
				FlatcarPackagesPath: tmpDir,
				NebraskaURL:         "http://localhost:8000",
			},
		},
		{
			name:      "github auth mode missing required fields",
			cfg:       Config{AuthMode: "github"},
			wantErr:   true,
			errSubstr: "invalid github configuration",
		},
		{
			name: "github auth mode with all required fields",
			cfg: Config{
				AuthMode:         "github",
				GhClientID:       "client-id",
				GhClientSecret:   "client-secret",
				GhReadOnlyTeams:  "org/ro-team",
				GhReadWriteTeams: "org/rw-team",
			},
		},
		{
			name: "github auth mode missing only read-write teams",
			cfg: Config{
				AuthMode:        "github",
				GhClientID:      "client-id",
				GhClientSecret:  "client-secret",
				GhReadOnlyTeams: "org/ro-team",
			},
			wantErr:   true,
			errSubstr: "invalid github configuration",
		},
		{
			name:      "oidc auth mode missing required fields",
			cfg:       Config{AuthMode: "oidc"},
			wantErr:   true,
			errSubstr: "invalid OIDC configuration",
		},
		{
			name: "oidc auth mode with all required fields",
			cfg: Config{
				AuthMode:        "oidc",
				OidcClientID:    "client-id",
				OidcIssuerURL:   "https://issuer.example.com",
				OidcAdminRoles:  "admin",
				OidcViewerRoles: "viewer",
			},
		},
		{
			name: "oidc auth mode missing only viewer roles",
			cfg: Config{
				AuthMode:       "oidc",
				OidcClientID:   "client-id",
				OidcIssuerURL:  "https://issuer.example.com",
				OidcAdminRoles: "admin",
			},
			wantErr:   true,
			errSubstr: "invalid OIDC configuration",
		},
		{
			name: "unknown auth mode is not validated and passes",
			cfg:  Config{AuthMode: "noop"},
		},
		{
			name: "valid ca-file path",
			cfg:  Config{CAFile: validCAFile},
		},
		{
			name:      "ca-file path does not exist",
			cfg:       Config{CAFile: filepath.Join(tmpDir, "missing-ca.pem")},
			wantErr:   true,
			errSubstr: "invalid ca-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPotentialOrEnv(t *testing.T) {
	const envName = "NEBRASKA_TEST_CONFIG_ENV_VAR"

	tests := []struct {
		name           string
		potentialValue string
		envValue       string
		envSet         bool
		want           string
	}{
		{
			name:           "potential value takes precedence over env var",
			potentialValue: "flag-value",
			envValue:       "env-value",
			envSet:         true,
			want:           "flag-value",
		},
		{
			name:           "falls back to env var when potential value is empty",
			potentialValue: "",
			envValue:       "env-value",
			envSet:         true,
			want:           "env-value",
		},
		{
			name: "returns empty string when neither is set",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envSet {
				t.Setenv(envName, tt.envValue)
			} else {
				os.Unsetenv(envName)
			}
			assert.Equal(t, tt.want, getPotentialOrEnv(tt.potentialValue, envName))
		})
	}
}
