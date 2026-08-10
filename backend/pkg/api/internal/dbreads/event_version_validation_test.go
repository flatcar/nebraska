package dbreads

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsValidSemverUsedInRegisterInstanceButNotRegisterEvent documents the
// inconsistency between RegisterInstance() and RegisterEvent() in version
// validation.
//
// RegisterInstance() in backend/pkg/api/instances.go validates the version:
//
//	if !dbreads.IsValidSemver(instApp.Version) {
//	    return nil, ErrInvalidSemver
//	}
//
// But RegisterEvent() in backend/pkg/api/events.go stores previousVersion
// directly into the database without calling IsValidSemver():
//
//	Vals(goqu.Vals{eventTypeID, instanceID, appID, previousVersion, errorCode})
//
// This means a malformed previousVersion like "not-a-version" gets stored in
// the event table. Later, when triggerEventConsequences() reads it back and
// calls semver.Make(lastUpdateVersion), the error is silently discarded and
// the version becomes 0.0.0.
func TestIsValidSemverUsedInRegisterInstanceButNotRegisterEvent(t *testing.T) {
	validVersions := []string{
		"1.0.0",
		"3.2.1",
		"0.0.1",
		"766.0.0",
	}

	invalidVersions := []string{
		"not-a-version",
		"v1.2.3",
		"1.2",
		"",
		"0.0.0.0", // Flatcar-specific, not valid semver
	}

	// RegisterInstance() correctly rejects these
	for _, v := range validVersions {
		assert.True(t, IsValidSemver(v),
			"valid version %q should pass IsValidSemver", v)
	}

	// RegisterEvent() does NOT call IsValidSemver on previousVersion
	// These would be silently stored in the DB if passed to RegisterEvent()
	for _, v := range invalidVersions {
		assert.False(t, IsValidSemver(v),
			"invalid version %q should fail IsValidSemver - but RegisterEvent() never checks this", v)
	}
}

// TestPreviousVersionValidationGap shows the specific versions that
// RegisterEvent() accepts but should reject.
//
// RegisterEvent() only has two special cases:
//
//	if previousVersion == "" || previousVersion == "0.0.0.0" {
//	    // handle Flatcar specific case
//	}
//
// Everything else goes straight to the DB without validation.
func TestPreviousVersionValidationGap(t *testing.T) {
	// These pass RegisterEvent()'s current check but are invalid semver
	// They would be stored in the DB and later cause silent 0.0.0 comparisons
	problematicVersions := []string{
		"v1.2.3",         // v-prefixed - passes "" and "0.0.0.0" check but invalid semver
		"not-a-version",  // garbage - passes the check but invalid semver
		"1.2",            // missing patch - passes the check but invalid semver
		"1.0.0.0",        // 4-component - passes the check but invalid semver
	}

	for _, v := range problematicVersions {
		t.Run(v, func(t *testing.T) {
			// RegisterEvent() current guard - only catches "" and "0.0.0.0"
			passesCurrentGuard := v != "" && v != "0.0.0.0"
			assert.True(t, passesCurrentGuard,
				"%q passes RegisterEvent() current guard and goes to DB unvalidated", v)

			// But IsValidSemver correctly rejects it
			assert.False(t, IsValidSemver(v),
				"%q should be rejected by IsValidSemver - RegisterEvent() should call this", v)
		})
	}
}

// TestRegisterEventProposedFix demonstrates what RegisterEvent() should do:
// call IsValidSemver on previousVersion before storing it,
// consistent with how RegisterInstance() validates instApp.Version.
func TestRegisterEventProposedFix(t *testing.T) {
	validatePreviousVersion := func(version string) bool {
		// Empty and "0.0.0.0" are special Flatcar cases - allow them
		if version == "" || version == "0.0.0.0" {
			return true
		}
		// All other versions must be valid semver
		return IsValidSemver(version)
	}

	// Valid semver - should pass
	assert.True(t, validatePreviousVersion("1.0.0"))
	assert.True(t, validatePreviousVersion("3.2.1"))

	// Special Flatcar cases - should pass
	assert.True(t, validatePreviousVersion(""))
	assert.True(t, validatePreviousVersion("0.0.0.0"))

	// Invalid semver - should be rejected
	assert.False(t, validatePreviousVersion("not-a-version"))
	assert.False(t, validatePreviousVersion("v1.2.3"))
	assert.False(t, validatePreviousVersion("1.2"))
}
