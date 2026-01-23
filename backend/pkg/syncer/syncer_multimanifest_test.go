package syncer

import (
	"os"
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
)

func createMultiManifestUpdateWithReasons() *omaha.UpdateResponse {
	update := createMultiManifestUpdate("1000.0.0", "2000.0.0", "3000.0.0")
	// Set specific floor reasons for testing
	update.Manifests[0].FloorReason = "Filesystem support for usr dir"
	update.Manifests[1].FloorReason = "Critical update"
	return update
}

// TestSyncer_MultiManifestWithExistingPackage tests the main bug fix:
// when a package exists, it should be verified and marked as floor
func TestSyncer_MultiManifestWithExistingPackage(t *testing.T) {
	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel
	require.NoError(t, syncer.initialize())

	// Pre-create first package
	tPkg, err := a.AddPackage(&api.Package{
		Type: api.PkgTypeFlatcar, URL: "https://example.com/1000.0.0",
		Version: "1000.0.0", Filename: null.StringFrom("flatcar-1000.0.0.gz"),
		Size: null.StringFrom("1000"), Hash: null.StringFrom("hash1000"),
		ApplicationID: flatcarAppID, Arch: tChannel.Arch,
	})
	require.NoError(t, err)
	_, err = a.AddFlatcarAction(&api.FlatcarAction{
		Event: "postinstall", Sha256: "dGVzdHNoYTI1Ng==", PackageID: tPkg.ID,
	})
	require.NoError(t, err)

	// Process multi-manifest update
	update := createMultiManifestUpdateWithReasons()
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err = syncer.processMultiManifestUpdate(desc, update)
	require.NoError(t, err)

	// Verify existing package was marked as floor
	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 2)
	assert.Equal(t, "1000.0.0", floors[0].Version)
	assert.Equal(t, "Filesystem support for usr dir", floors[0].FloorReason.String)
}

// TestSyncer_PackageVerificationErrors tests hash/size mismatch detection
func TestSyncer_PackageVerificationErrors(t *testing.T) {
	tests := []struct {
		name   string
		hash   string
		size   string
		errMsg string
	}{
		{"hash mismatch", "WRONGHASH", "1000", "hash mismatch"},
		{"size mismatch", "hash1000", "9999", "size mismatch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer := newForTest(t, &Config{})
			a := syncer.api
			t.Cleanup(func() { a.Close() })

			tGroup := setupFlatcarAppStableGroup(t, a)
			tChannel := tGroup.Channel
			require.NoError(t, syncer.initialize())

			// Create package with wrong hash or size
			tPkg, err := a.AddPackage(&api.Package{
				Type: api.PkgTypeFlatcar, URL: "https://example.com/1000.0.0",
				Version: "1000.0.0", Filename: null.StringFrom("flatcar-1000.0.0.gz"),
				Size: null.StringFrom(tt.size), Hash: null.StringFrom(tt.hash),
				ApplicationID: flatcarAppID, Arch: tChannel.Arch,
			})
			require.NoError(t, err)
			_, err = a.AddFlatcarAction(&api.FlatcarAction{
				Event: "postinstall", Sha256: "dGVzdHNoYTI1Ng==", PackageID: tPkg.ID,
			})
			require.NoError(t, err)

			// Process should fail
			update := createMultiManifestUpdateWithReasons()
			desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
			err = syncer.processMultiManifestUpdate(desc, update)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
			assert.Contains(t, err.Error(), "1000.0.0")
		})
	}
}

// TestSyncer_MissingFlatcarAction tests error when FlatcarAction is missing
func TestSyncer_MissingFlatcarAction(t *testing.T) {
	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel
	require.NoError(t, syncer.initialize())

	// Create package WITHOUT FlatcarAction
	_, err := a.AddPackage(&api.Package{
		Type: api.PkgTypeFlatcar, URL: "https://example.com/1000.0.0",
		Version: "1000.0.0", Filename: null.StringFrom("flatcar-1000.0.0.gz"),
		Size: null.StringFrom("1000"), Hash: null.StringFrom("hash1000"),
		ApplicationID: flatcarAppID, Arch: tChannel.Arch,
	})
	require.NoError(t, err)

	// Process should fail
	update := createMultiManifestUpdateWithReasons()
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err = syncer.processMultiManifestUpdate(desc, update)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing FlatcarAction")
}

// TestSyncer_TargetAsFloor tests when target package is both floor and target
func TestSyncer_TargetAsFloor(t *testing.T) {
	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel
	require.NoError(t, syncer.initialize())

	// Create update where last manifest has BOTH IsFloor and IsTarget
	update := &omaha.UpdateResponse{
		Status: "ok",
		URLs: []*omaha.URL{
			{CodeBase: "https://example.com"},
		},
		Manifests: []*omaha.Manifest{
			{
				Version:     "1000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-1000.0.0.gz", SHA1: "hash1000", Size: 1000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Critical bootloader update",
			},
			{
				Version:     "2000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-2000.0.0.gz", SHA1: "hash2000", Size: 2000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				IsTarget:    true, // BOTH floor AND target
				FloorReason: "Critical mandatory version",
			},
		},
	}

	// Process multi-manifest update
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err := syncer.processMultiManifestUpdate(desc, update)
	require.NoError(t, err)

	// Verify both packages were marked as floors
	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 2)
	assert.Equal(t, "1000.0.0", floors[0].Version)
	assert.Equal(t, "Critical bootloader update", floors[0].FloorReason.String)
	assert.Equal(t, "2000.0.0", floors[1].Version)
	assert.Equal(t, "Critical mandatory version", floors[1].FloorReason.String)

	// Verify channel points to the target (2000.0.0)
	updatedChannel, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "2000.0.0", updatedChannel.Package.Version)

	// Verify the target package is marked as floor
	assert.True(t, floors[1].IsFloor)
	assert.Equal(t, updatedChannel.Package.ID, floors[1].ID)
}

// TestSyncer_AllFloorsNoTarget tests all-floors response (no explicit target)
func TestSyncer_AllFloorsNoTarget(t *testing.T) {
	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel
	require.NoError(t, syncer.initialize())

	// Pre-set channel to existing version
	existingPkg, err := a.AddPackage(&api.Package{
		Type: api.PkgTypeFlatcar, URL: "https://example.com/500.0.0",
		Version: "500.0.0", Filename: null.StringFrom("flatcar-500.0.0.gz"),
		ApplicationID: flatcarAppID, Arch: tChannel.Arch,
	})
	require.NoError(t, err)
	tChannel.PackageID = null.StringFrom(existingPkg.ID)
	err = a.UpdateChannel(tChannel)
	require.NoError(t, err)

	// Create update with ONLY floors, no target
	update := &omaha.UpdateResponse{
		Status: "ok",
		URLs: []*omaha.URL{
			{CodeBase: "https://example.com"},
		},
		Manifests: []*omaha.Manifest{
			{
				Version:     "1000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-1000.0.0.gz", SHA1: "hash1000", Size: 1000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Floor 1",
			},
			{
				Version:     "2000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-2000.0.0.gz", SHA1: "hash2000", Size: 2000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Floor 2",
			},
		},
	}

	// Process should succeed - floors are added but channel stays at 500.0.0
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err = syncer.processMultiManifestUpdate(desc, update)
	require.NoError(t, err)

	// Verify floors were added
	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 2)

	// Verify channel stayed at existing version (not updated)
	updatedChannel, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "500.0.0", updatedChannel.Package.Version)
}

// TestSyncer_BackwardCompatibilityNoFlags tests backward compatibility:
// When no explicit flags are set, last manifest becomes implicit target
func TestSyncer_BackwardCompatibilityNoFlags(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)

	// Create update with NO explicit flags (legacy behavior)
	update := &omaha.UpdateResponse{
		Status: "ok",
		URLs: []*omaha.URL{
			{CodeBase: "https://example.com"},
		},
		Manifests: []*omaha.Manifest{
			{
				Version:  "1000.0.0",
				Packages: []*omaha.Package{{Name: "flatcar-1000.0.0.gz", SHA1: "hash1000", Size: 1000}},
				Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				// NO IsFloor or IsTarget flags
			},
			{
				Version:  "2000.0.0",
				Packages: []*omaha.Package{{Name: "flatcar-2000.0.0.gz", SHA1: "hash2000", Size: 2000}},
				Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				// NO IsFloor or IsTarget flags
			},
		},
	}

	// Process multi-manifest update
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err := syncer.processMultiManifestUpdate(desc, update)
	require.NoError(t, err)

	// Verify NO floors were marked (backward compatibility)
	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 0, "No floors should be marked in backward compatibility mode")

	// Verify channel points to LAST manifest (2000.0.0) - implicit target
	updatedChannel, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "2000.0.0", updatedChannel.Package.Version, "Last manifest should be implicit target")
}

// TestSyncer_ExplicitTargetPriority tests that explicit is_target takes priority
// over implicit last-non-floor detection
func TestSyncer_ExplicitTargetPriority(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)

	// Create update where middle manifest has explicit is_target
	// Last manifest is non-floor but should NOT be target
	update := &omaha.UpdateResponse{
		Status: "ok",
		URLs: []*omaha.URL{
			{CodeBase: "https://example.com"},
		},
		Manifests: []*omaha.Manifest{
			{
				Version:     "1000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-1000.0.0.gz", SHA1: "hash1000", Size: 1000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Filesystem support for usr dir",
			},
			{
				Version:  "2000.0.0",
				Packages: []*omaha.Package{{Name: "flatcar-2000.0.0.gz", SHA1: "hash2000", Size: 2000}},
				Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsTarget: true, // Explicit target in middle
			},
			{
				Version:  "3000.0.0",
				Packages: []*omaha.Package{{Name: "flatcar-3000.0.0.gz", SHA1: "hash3000", Size: 3000}},
				Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				// Last non-floor, but NOT marked as target
			},
		},
	}

	// Process multi-manifest update
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err := syncer.processMultiManifestUpdate(desc, update)
	require.NoError(t, err)

	// Verify channel points to EXPLICIT target (2000.0.0), not last (3000.0.0)
	updatedChannel, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "2000.0.0", updatedChannel.Package.Version, "Explicit target should take priority")
}

// TestSyncer_MixedFloorsAndRegular tests mixed floor and regular packages
func TestSyncer_MixedFloorsAndRegular(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)

	// Create [floor, regular, floor, regular] pattern
	update := &omaha.UpdateResponse{
		Status: "ok",
		URLs: []*omaha.URL{
			{CodeBase: "https://example.com"},
		},
		Manifests: []*omaha.Manifest{
			{
				Version:     "1000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-1000.0.0.gz", SHA1: "hash1000", Size: 1000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "First floor",
			},
			{
				Version:  "2000.0.0",
				Packages: []*omaha.Package{{Name: "flatcar-2000.0.0.gz", SHA1: "hash2000", Size: 2000}},
				Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				// Regular package
			},
			{
				Version:     "3000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-3000.0.0.gz", SHA1: "hash3000", Size: 3000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Second floor",
			},
			{
				Version:  "4000.0.0",
				Packages: []*omaha.Package{{Name: "flatcar-4000.0.0.gz", SHA1: "hash4000", Size: 4000}},
				Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				// Last regular package (implicit target)
			},
		},
	}

	// Process multi-manifest update
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err := syncer.processMultiManifestUpdate(desc, update)
	require.NoError(t, err)

	// Verify only marked floors are in floor list
	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 2, "Only explicitly marked floors should be in floor list")
	assert.Equal(t, "1000.0.0", floors[0].Version)
	assert.Equal(t, "3000.0.0", floors[1].Version)

	// Verify channel points to last non-floor (4000.0.0)
	updatedChannel, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "4000.0.0", updatedChannel.Package.Version)
}

// TestSyncer_EmptyManifestError tests error handling for empty manifests
func TestSyncer_EmptyManifestError(t *testing.T) {
	syncer, _, _, tChannel := setupSyncerTest(t)

	// Create update with empty manifests array
	update := &omaha.UpdateResponse{
		Status:    "ok",
		URLs:      []*omaha.URL{{CodeBase: "https://example.com"}},
		Manifests: []*omaha.Manifest{}, // Empty!
	}

	// Process should fail
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	err := syncer.processMultiManifestUpdate(desc, update)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manifests")
}

// TestSyncer_FloorLimitVersionTracking tests that when there are more floors than
// NEBRASKA_MAX_FLOORS_PER_RESPONSE, the syncer correctly syncs ALL floors across
// multiple sync rounds.
//
// When there are more floors remaining beyond the limit, the server
// sends ONLY floors (no target). This way:
// - Syncer processes floors and updates channel to highest floor
// - Syncer tracks highest floor version for next request
// - Next request fetches remaining floors
// - Only when all floors are sent does the server include the target
//
// Scenario with limit=2 and 5 floors:
// - Round 1: syncer at 0.0.0 -> server sends floors 1,2 (no target, more floors remain)
// - Round 2: syncer at 2000.0.0 -> server sends floors 3,4 (no target, more floors remain)
// - Round 3: syncer at 4000.0.0 -> server sends floor 5 + target (all floors sent)
func TestSyncer_FloorLimitVersionTracking(t *testing.T) {
	// Set a low limit to test pagination behavior
	oldMax := os.Getenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE")
	defer os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", oldMax)
	os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", "2")

	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel
	require.NoError(t, syncer.initialize())

	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}

	// Round 1: Server sends only floors (no target) because more floors remain
	// This simulates what upstream Nebraska would send when there are 5 floors but limit is 2
	round1 := &omaha.UpdateResponse{
		Status: "ok",
		URLs:   []*omaha.URL{{CodeBase: "https://example.com"}},
		Manifests: []*omaha.Manifest{
			{
				Version:     "1000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-1000.0.0.gz", SHA1: "hash1000", Size: 1000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Floor 1",
			},
			{
				Version:     "2000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-2000.0.0.gz", SHA1: "hash2000", Size: 2000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Floor 2",
			},
			// NO TARGET - more floors remain
		},
	}

	// Process round 1
	err := syncer.processMultiManifestUpdate(desc, round1)
	require.NoError(t, err)

	// After round 1: syncer should track highest floor (2000.0.0)
	// Channel should NOT be updated (no target in response)
	trackedVersion := syncer.versions[desc]
	t.Logf("After round 1, syncer tracked version: %s", trackedVersion)

	// Verify floors 1 and 2 were synced
	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 2, "Should have 2 floors after round 1")

	// Round 2: Server sends next batch of floors (still no target, more remain)
	round2 := &omaha.UpdateResponse{
		Status: "ok",
		URLs:   []*omaha.URL{{CodeBase: "https://example.com"}},
		Manifests: []*omaha.Manifest{
			{
				Version:     "3000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-3000.0.0.gz", SHA1: "hash3000", Size: 3000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Floor 3",
			},
			{
				Version:     "4000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-4000.0.0.gz", SHA1: "hash4000", Size: 4000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Floor 4",
			},
			// NO TARGET - more floors remain
		},
	}

	// Process round 2
	err = syncer.processMultiManifestUpdate(desc, round2)
	require.NoError(t, err)

	// Verify floors 3 and 4 were synced
	floors, err = a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 4, "Should have 4 floors after round 2")

	// Round 3: Server sends last floor + target (all floors now sent)
	round3 := &omaha.UpdateResponse{
		Status: "ok",
		URLs:   []*omaha.URL{{CodeBase: "https://example.com"}},
		Manifests: []*omaha.Manifest{
			{
				Version:     "5000.0.0",
				Packages:    []*omaha.Package{{Name: "flatcar-5000.0.0.gz", SHA1: "hash5000", Size: 5000}},
				Actions:     []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsFloor:     true,
				FloorReason: "Floor 5",
			},
			{
				Version:  "6000.0.0",
				Packages: []*omaha.Package{{Name: "flatcar-6000.0.0.gz", SHA1: "hash6000", Size: 6000}},
				Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
				IsTarget: true,
			},
		},
	}

	// Process round 3
	err = syncer.processMultiManifestUpdate(desc, round3)
	require.NoError(t, err)

	// Final verification: ALL 5 floors should be synced
	floors, err = a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	assert.Len(t, floors, 5, "All 5 floors should be synced after 3 rounds")

	// Verify floor versions
	floorVersions := make([]string, len(floors))
	for i, f := range floors {
		floorVersions[i] = f.Version
	}
	assert.Contains(t, floorVersions, "1000.0.0", "Floor 1 should be synced")
	assert.Contains(t, floorVersions, "2000.0.0", "Floor 2 should be synced")
	assert.Contains(t, floorVersions, "3000.0.0", "Floor 3 should be synced")
	assert.Contains(t, floorVersions, "4000.0.0", "Floor 4 should be synced")
	assert.Contains(t, floorVersions, "5000.0.0", "Floor 5 should be synced")

	// Verify channel now points to target (only after all floors sent)
	updatedChannel, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "6000.0.0", updatedChannel.Package.Version, "Channel should point to target after all floors synced")

	// Verify syncer now tracks target version (ready for next sync cycle)
	finalVersion := syncer.versions[desc]
	assert.Equal(t, "6000.0.0", finalVersion, "Syncer should track target after all floors synced")
}
