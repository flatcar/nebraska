package syncer

import (
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
	update.Manifests[0].FloorReason = "Security fix"
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
	assert.Equal(t, "Security fix", floors[0].FloorReason.String)
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
				FloorReason: "Security update",
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
	assert.Equal(t, "Security update", floors[0].FloorReason.String)
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
				FloorReason: "Security fix",
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
