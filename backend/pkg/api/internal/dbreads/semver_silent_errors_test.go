package dbreads

import (
	"fmt"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSemverMakeSilentlyDiscardedInUpdatesGo documents the silent semver.Make()
// error discards in backend/pkg/api/updates.go.
//
// There are 6 occurrences in updates.go where semver.Make() errors are
// discarded with _:
//
//	instanceSemver, _ := semver.Make(instanceVersion)  // lines 111, 144, 230
//	grantedSemver,  _ := semver.Make(grantedVersion)   // line 112
//	packageSemver,  _ := semver.Make(...)              // lines 145, 231
//
// When semver.Make() fails it returns semver.Version{} (0.0.0).
// Since 0.0.0 < any real package version, a malformed instanceVersion
// always passes the version check and gets an update granted incorrectly.
func TestSemverMakeSilentlyDiscardedInUpdatesGo(t *testing.T) {
	// Valid version parses fine
	v, err := semver.Make("3.0.0")
	require.NoError(t, err)
	assert.Equal(t, "3.0.0", v.String())

	// Invalid version - error is silently discarded in updates.go with _
	malformed, err := semver.Make("not-a-version")
	assert.Error(t, err,
		"semver.Make returns error for malformed version - updates.go discards it with _")
	assert.Equal(t, semver.Version{}, malformed,
		"malformed version silently becomes 0.0.0")

	// The consequence: 0.0.0 < any real package version
	pkg, _ := semver.Make("3.0.0")
	assert.True(t, malformed.LT(pkg),
		"BUG: 0.0.0 (from malformed) < 3.0.0 → update incorrectly granted to bad instance")
}

// TestMalformedVersionsAllBecome0_0_0 shows common malformed version strings
// that silently become 0.0.0 in updates.go.
func TestMalformedVersionsAllBecome0_0_0(t *testing.T) {
	tests := []struct {
		version string
		desc    string
	}{
		{"not-a-version", "garbage string"},
		{"v1.2.3", "v-prefixed (common mistake)"},
		{"1.2", "missing patch component"},
		{"", "empty string"},
		{"abc.def.ghi", "non-numeric components"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			parsed, err := semver.Make(tc.version)

			assert.Error(t, err,
				"%s (%q) should fail to parse", tc.desc, tc.version)
			assert.Equal(t, semver.Version{}, parsed,
				"%s becomes 0.0.0 - silently in updates.go", tc.desc)

			realPkg, _ := semver.Make("1.0.0")
			assert.True(t, parsed.LT(realPkg),
				"0.0.0 from %q < 1.0.0 → update granted to instance with bad version",
				tc.version)
		})
	}
}

// TestProposedFix shows what the fix should look like in updates.go.
// Instead of discarding the error, return it so callers handle it gracefully.
func TestProposedFix(t *testing.T) {
	parseVersion := func(version string) (semver.Version, error) {
		v, err := semver.Make(version)
		if err != nil {
			return semver.Version{}, fmt.Errorf("invalid semver %q: %w", version, err)
		}
		return v, nil
	}

	// Valid version - works fine
	_, err := parseVersion("3.0.0")
	assert.NoError(t, err)

	// Malformed version - returns error instead of silently using 0.0.0
	_, err = parseVersion("not-a-version")
	assert.Error(t, err,
		"proposed fix: malformed version returns error instead of silently using 0.0.0")
	assert.Contains(t, err.Error(), "invalid semver")
}
