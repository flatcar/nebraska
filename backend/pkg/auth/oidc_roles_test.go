package auth

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}
	os.Exit(m.Run())
}

// TestRolesPathExtraction tests the gjson path extraction logic used in both
// rolesFromToken and rolesFromUserInfo
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
			expectedRoles: nil,
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
			claimsJSON, err := json.Marshal(tt.claims)
			assert.NoError(t, err)

			result := gjson.GetBytes(claimsJSON, tt.rolesPath)
			assert.Equal(t, tt.expectExists, result.Exists())

			if tt.expectExists {
				var roles []string
				result.ForEach(func(_, value gjson.Result) bool {
					roles = append(roles, value.String())
					return true
				})
				assert.Equal(t, tt.expectedRoles, roles)
			}
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
