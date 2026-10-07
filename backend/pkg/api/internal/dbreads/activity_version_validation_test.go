package dbreads

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestActivityVersionValidationGap documents the inconsistency between
// newGroupActivityEntry()/newInstanceActivityEntry() in activity.go and
// RegisterInstance() in instances.go.
//
// RegisterInstance() validates the version string before storing it:
//
//	if !dbreads.IsValidSemver(instApp.Version) {
//	    return nil, ErrInvalidSemver
//	}
//
// But newGroupActivityEntry() and newInstanceActivityEntry() in activity.go
// insert the version string directly into the activity table with NO validation:
//
//	// newGroupActivityEntry() - no version validation
//	Vals(goqu.Vals{class, severity, version, appID, groupID})
//
//	// newInstanceActivityEntry() - no version validation
//	Vals(goqu.Vals{class, severity, version, appID, groupID, instanceID})
//
// This means malformed version strings can be stored in the activity table,
// making activity logs unreliable.
func TestActivityVersionValidationGap(t *testing.T) {
	// These versions are valid semver - should be accepted everywhere
	validVersions := []string{
		"1.0.0",
		"3.2.1",
		"766.0.0",
		"0.0.1",
	}

	// These versions are invalid semver - RegisterInstance() rejects them
	// but newGroupActivityEntry() and newInstanceActivityEntry() accept them
	invalidVersions := []string{
		"not-a-version",
		"v1.2.3",   // v-prefixed
		"1.2",      // missing patch
		"",         // empty string
		"1.0.0.0",  // 4-component version
	}

	for _, v := range validVersions {
		assert.True(t, IsValidSemver(v),
			"valid version %q should pass IsValidSemver", v)
	}

	for _, v := range invalidVersions {
		assert.False(t, IsValidSemver(v),
			"invalid version %q should fail IsValidSemver - "+
				"but newGroupActivityEntry() and newInstanceActivityEntry() never check this", v)
	}
}

// TestActivityVersionValidationInconsistency shows exactly which functions
// validate version and which do not, across the Nebraska API layer.
//
// This documents the inconsistency that leads to malformed versions in the DB:
//
//	RegisterInstance()          → validates with IsValidSemver() ✅
//	RegisterEvent()             → NO validation ❌ (issue #1553)
//	newGroupActivityEntry()     → NO validation ❌ (this issue)
//	newInstanceActivityEntry()  → NO validation ❌ (this issue)
func TestActivityVersionValidationInconsistency(t *testing.T) {
	malformedVersion := "v1.2.3" // v-prefixed, common mistake

	// IsValidSemver correctly rejects it
	assert.False(t, IsValidSemver(malformedVersion),
		"v-prefixed version should be rejected by IsValidSemver")

	// But activity.go functions would store it - they only call:
	// goqu.Insert("activity").Vals(goqu.Vals{class, severity, version, ...})
	// with no IsValidSemver check before insertion.

	// The proposed fix is consistent validation:
	validateActivityVersion := func(version string) bool {
		if version == "" {
			return true // empty version is allowed in activity entries
		}
		return IsValidSemver(version)
	}

	assert.True(t, validateActivityVersion("1.0.0"),
		"valid semver should pass proposed validation")
	assert.True(t, validateActivityVersion(""),
		"empty version should be allowed in activity entries")
	assert.False(t, validateActivityVersion("v1.2.3"),
		"v-prefixed version should be rejected by proposed validation")
	assert.False(t, validateActivityVersion("not-a-version"),
		"garbage version should be rejected by proposed validation")
}

// TestActivityFunctionsAffected lists all activity functions that insert
// version into the DB without validation, for traceability.
func TestActivityFunctionsAffected(t *testing.T) {
	// This test documents which functions are affected.
	// Both are in backend/pkg/api/activity.go
	affectedFunctions := []string{
		"newGroupActivityEntry(class, severity, version, appID, groupID)",
		"newInstanceActivityEntry(class, severity, version, appID, groupID, instanceID)",
	}

	// And the callers that pass version strings to these functions:
	// - triggerEventConsequences() in events.go
	// - processUpdate() in syncer/syncer.go
	// All pass version strings that could be malformed

	for _, fn := range affectedFunctions {
		t.Logf("MISSING validation in: %s", fn)
	}

	// Verify IsValidSemver exists and works correctly as the proposed fix
	assert.True(t, IsValidSemver("1.0.0"))
	assert.False(t, IsValidSemver("not-valid"))
}
