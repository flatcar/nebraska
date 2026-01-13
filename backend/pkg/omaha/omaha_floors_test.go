package omaha

import (
	"bytes"
	"encoding/xml"
	"os"
	"testing"

	omahaSpec "github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestFloorUpdateScenarios(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	h := NewHandler(a)

	// RegularClient: Tests that regular Flatcar clients receive ONE package at a time
	// Setup: floors [2000, 2500], target 3000
	// Regular clients progress through floors sequentially with reboots between each step
	t.Run("RegularClient", func(t *testing.T) {
		group, _ := setupOmahaFloorTest(t, a, "regular", []string{"2000.0.0", "2500.0.0"}, "3000.0.0")

		tests := []struct {
			instance, expected string
		}{
			{"1500.0.0", "2000.0.0"}, // below all floors -> gets first floor
			{"2000.0.0", "2500.0.0"}, // at first floor -> gets second floor
			{"2200.0.0", "2500.0.0"}, // between floors -> gets next floor above
			{"2500.0.0", "3000.0.0"}, // at last floor -> gets target
			{"2700.0.0", "3000.0.0"}, // above floors but below target -> gets target
		}
		for _, tc := range tests {
			resp := doOmahaRequest(t, h, flatcarAppID, tc.instance, "client-"+tc.instance, group.ID, "10.0.0.1", false, true, nil)
			checkOmahaUpdateResponse(t, resp, tc.expected, "flatcar_"+tc.expected+".gz", "http://sample.url/"+tc.expected, omahaSpec.UpdateOK)
			assert.Len(t, resp.Apps[0].UpdateCheck.Manifests, 1)
		}
	})

	// RegularClientNoUpdate: Client already at target version gets NoUpdate
	t.Run("RegularClientNoUpdate", func(t *testing.T) {
		group, _ := setupOmahaFloorTest(t, a, "noupdate", []string{}, "5000.0.0")
		resp := doOmahaRequest(t, h, flatcarAppID, "5000.0.0", "client", group.ID, "10.0.0.1", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "", "", "", omahaSpec.NoUpdate)
	})

	// SyncerMultiManifest: Modern syncers (MultiManifestOK=true) receive ALL packages in one response
	// Setup: floors [9000, 9500], target 10000
	// Unlike regular clients, syncers get floors + target together to sync them all at once
	t.Run("SyncerMultiManifest", func(t *testing.T) {
		group, _ := setupOmahaFloorTest(t, a, "syncer", []string{"9000.0.0", "9500.0.0"}, "10000.0.0")

		tests := []struct {
			version           string
			expectedManifests int
		}{
			{"8500.0.0", 3}, // below all -> gets 2 floors + target
			{"9200.0.0", 2}, // between floors -> gets 1 floor + target
			{"9700.0.0", 1}, // above floors -> gets only target
		}
		for _, tc := range tests {
			resp := doSyncerRequest(t, h, tc.version, group.ID, true)
			require.Len(t, resp.Apps[0].UpdateCheck.Manifests, tc.expectedManifests)
			assert.True(t, resp.Apps[0].UpdateCheck.Manifests[tc.expectedManifests-1].IsTarget)
			for i := 0; i < tc.expectedManifests-1; i++ {
				assert.True(t, resp.Apps[0].UpdateCheck.Manifests[i].IsFloor)
			}
		}
	})

	// LegacySyncerBlocked: Old syncers without MultiManifestOK get NoUpdate when floors exist
	// This prevents legacy syncers from skipping mandatory floor versions
	t.Run("LegacySyncerBlocked", func(t *testing.T) {
		group, _ := setupOmahaFloorTest(t, a, "oldsyncer", []string{"6000.0.0"}, "7000.0.0")
		resp := doSyncerRequest(t, h, "1500.0.0", group.ID, false) // multiManifestOK=false
		checkOmahaUpdateResponse(t, resp, "", "", "", omahaSpec.NoUpdate)
	})

	// ModernSyncerNoUpdate: Modern syncer already at target version gets NoUpdate
	t.Run("ModernSyncerNoUpdate", func(t *testing.T) {
		group, _ := setupOmahaFloorTest(t, a, "syncernoup", []string{}, "8000.0.0")
		resp := doSyncerRequest(t, h, "8000.0.0", group.ID, true) // at target version
		checkOmahaUpdateResponse(t, resp, "", "", "", omahaSpec.NoUpdate)
	})

	// NoFloors: When no floors are configured, both clients and syncers get direct update to target
	t.Run("NoFloors", func(t *testing.T) {
		group, _ := setupOmahaFloorTest(t, a, "nofloor", []string{}, "4000.0.0")

		// Regular client gets direct update
		resp := doOmahaRequest(t, h, flatcarAppID, "1000.0.0", "client", group.ID, "10.0.0.1", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "4000.0.0", "flatcar_4000.0.0.gz", "http://sample.url/4000.0.0", omahaSpec.UpdateOK)

		// Syncer gets single manifest with IsTarget=true
		resp = doSyncerRequest(t, h, "1000.0.0", group.ID, true)
		require.Len(t, resp.Apps[0].UpdateCheck.Manifests, 1)
		assert.True(t, resp.Apps[0].UpdateCheck.Manifests[0].IsTarget)
	})

	// TargetAsFloor: Target package can also be marked as a floor
	// Setup: floors [11000, 12000], target 13000, then mark 13000 as floor too
	// Regular clients progress through floors including the target-floor
	// Syncers receive all manifests with the last one having BOTH IsFloor=true AND IsTarget=true
	t.Run("TargetAsFloor", func(t *testing.T) {
		group, pkgs := setupOmahaFloorTest(t, a, "targetfloor", []string{"11000.0.0", "12000.0.0"}, "13000.0.0")
		// Mark target as also being a floor (critical version that must be installed)
		require.NoError(t, a.AddChannelPackageFloor(group.ChannelID.String, pkgs[2].ID, null.StringFrom("Critical mandatory version")))

		// Regular client below all versions -> gets first floor
		resp := doOmahaRequest(t, h, flatcarAppID, "10000.0.0", "client-low", group.ID, "10.0.0.1", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "11000.0.0", "flatcar_11000.0.0.gz", "http://sample.url/11000.0.0", omahaSpec.UpdateOK)

		// Regular client between floors -> gets second floor
		resp = doOmahaRequest(t, h, flatcarAppID, "11500.0.0", "client-mid", group.ID, "10.0.0.2", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "12000.0.0", "flatcar_12000.0.0.gz", "http://sample.url/12000.0.0", omahaSpec.UpdateOK)

		// Regular client above regular floors but below target-floor -> gets target (which is also a floor)
		resp = doOmahaRequest(t, h, flatcarAppID, "12500.0.0", "client-high", group.ID, "10.0.0.3", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "13000.0.0", "flatcar_13000.0.0.gz", "http://sample.url/13000.0.0", omahaSpec.UpdateOK)

		// Syncer should get all manifests with correct flags
		resp = doSyncerRequest(t, h, "10000.0.0", group.ID, true)
		require.Len(t, resp.Apps[0].UpdateCheck.Manifests, 3)

		// First two are floors only
		assert.True(t, resp.Apps[0].UpdateCheck.Manifests[0].IsFloor)
		assert.False(t, resp.Apps[0].UpdateCheck.Manifests[0].IsTarget)
		assert.True(t, resp.Apps[0].UpdateCheck.Manifests[1].IsFloor)
		assert.False(t, resp.Apps[0].UpdateCheck.Manifests[1].IsTarget)
		// Last one is BOTH floor AND target
		assert.True(t, resp.Apps[0].UpdateCheck.Manifests[2].IsFloor)
		assert.True(t, resp.Apps[0].UpdateCheck.Manifests[2].IsTarget)
		assert.Equal(t, "Critical mandatory version", resp.Apps[0].UpdateCheck.Manifests[2].FloorReason)
	})
}

func TestFloorLimitPagination(t *testing.T) {
	oldMax := os.Getenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE")
	defer os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", oldMax)
	os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", "2")

	a := newForTest(t)
	defer a.Close()
	h := NewHandler(a)

	group, _ := setupOmahaFloorTest(t, a, "floor-limit",
		[]string{"1000.0.0", "2000.0.0", "3000.0.0", "4000.0.0", "5000.0.0"}, "6000.0.0")

	// Test scenarios for floor limit pagination with limit=2 and floors [1000, 2000, 3000, 4000, 5000] + target 6000
	//
	// Scenario 1 (round1): Syncer at 0.0.0
	//   - 5 floors remain (1000-5000), exceeds limit of 2
	//   - Returns floors [1000, 2000], NO target (hasTarget=false)
	//   - Syncer should request again with version 2000.0.0
	//
	// Scenario 2 (round2): Syncer at 2000.0.0 (after processing round1)
	//   - 3 floors remain (3000-5000), exceeds limit of 2
	//   - Returns floors [3000, 4000], NO target (hasTarget=false)
	//   - Syncer should request again with version 4000.0.0
	//
	// Scenario 3 (round3): Syncer at 4000.0.0 (after processing round2)
	//   - 1 floor remains (5000), under limit
	//   - Returns floor [5000] + target [6000] (hasTarget=true)
	//   - All floors sent, syncer can update channel to target
	//
	// Scenario 4 (at_limit): Syncer at 3000.0.0
	//   - 2 floors remain (4000, 5000), exactly at limit
	//   - Returns floors [4000, 5000] + target [6000] (hasTarget=true)
	//   - All floors sent, syncer can update channel to target
	tests := []struct {
		name          string
		version       string // syncer's current version
		expectedCount int    // number of manifests in response
		expectedFirst string // first manifest version
		expectedLast  string // last manifest version
		hasTarget     bool   // whether target is included (all floors sent)
	}{
		{"round1", "0.0.0", 2, "1000.0.0", "2000.0.0", false},
		{"round2", "2000.0.0", 2, "3000.0.0", "4000.0.0", false},
		{"round3", "4000.0.0", 2, "5000.0.0", "6000.0.0", true},
		{"at_limit", "3000.0.0", 3, "4000.0.0", "6000.0.0", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := doSyncerRequest(t, h, tc.version, group.ID, true)

			require.Len(t, resp.Apps[0].UpdateCheck.Manifests, tc.expectedCount)
			assert.Equal(t, tc.expectedFirst, resp.Apps[0].UpdateCheck.Manifests[0].Version)
			assert.Equal(t, tc.expectedLast, resp.Apps[0].UpdateCheck.Manifests[tc.expectedCount-1].Version)

			for i, m := range resp.Apps[0].UpdateCheck.Manifests {
				isLast := i == tc.expectedCount-1
				if tc.hasTarget && isLast {
					assert.True(t, m.IsTarget, "last manifest should be target")
				} else {
					assert.True(t, m.IsFloor, "manifest %d should be floor", i)
					assert.False(t, m.IsTarget, "manifest %d should not be target when more floors remain", i)
				}
			}
		})
	}
}

// TestLegacySyncerBlockedWithFloors tests that legacy syncers without MultiManifestOK are blocked when floors exist
func TestLegacySyncerBlockedWithFloors(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	h := NewHandler(a)

	// Setup floor configuration
	group, _ := setupOmahaFloorTest(t, a, "legacy-syncer", []string{"1000.0.0", "2000.0.0"}, "3000.0.0")

	// Helper for syncer request without MultiManifestOK
	legacySyncerRequest := func(version string) *omahaSpec.Response {
		req := omahaSpec.NewRequest()
		req.OS.Version = "3"
		req.OS.Platform = "CoreOS"
		req.OS.ServicePack = "linux"
		req.OS.Arch = "x64"
		req.Version = "CoreOSUpdateEngine-0.1.0.0"
		req.InstallSource = "scheduler"
		app := req.AddApp(flatcarAppID, version)
		app.MachineID = "legacy-syncer-" + version
		app.Track = group.ID
		app.AddUpdateCheck()
		app.MultiManifestOK = false // Legacy syncer without multi-manifest support

		buf := bytes.NewBuffer(nil)
		err := xml.NewEncoder(buf).Encode(req)
		require.NoError(t, err)

		respBuf := bytes.NewBuffer(nil)
		err = h.Handle(buf, respBuf, "10.0.0.1")
		if err != nil {
			t.Fatalf("Failed to handle request: %v", err)
		}
		var resp omahaSpec.Response
		err = xml.NewDecoder(respBuf).Decode(&resp)
		require.NoError(t, err)
		return &resp
	}

	t.Run("legacy_syncer_blocked_when_floors_exist", func(t *testing.T) {
		// Legacy syncer at version 500.0.0 trying to update
		resp := legacySyncerRequest("500.0.0")

		// Should get NoUpdate response because floors exist and syncer can't handle them
		require.Len(t, resp.Apps, 1)
		require.NotNil(t, resp.Apps[0].UpdateCheck)
		assert.Equal(t, omahaSpec.NoUpdate, resp.Apps[0].UpdateCheck.Status)
		assert.Empty(t, resp.Apps[0].UpdateCheck.Manifests)
	})

	t.Run("legacy_syncer_at_intermediate_version", func(t *testing.T) {
		// Legacy syncer at version between floors
		resp := legacySyncerRequest("1500.0.0")

		// Should still get NoUpdate
		require.Len(t, resp.Apps, 1)
		require.NotNil(t, resp.Apps[0].UpdateCheck)
		assert.Equal(t, omahaSpec.NoUpdate, resp.Apps[0].UpdateCheck.Status)
		assert.Empty(t, resp.Apps[0].UpdateCheck.Manifests)
	})

	t.Run("modern_syncer_with_multimanifest_gets_update", func(t *testing.T) {
		// Modern syncer with MultiManifestOK=true at same version
		req := omahaSpec.NewRequest()
		req.OS.Version = "3"
		req.OS.Platform = "CoreOS"
		req.OS.ServicePack = "linux"
		req.OS.Arch = "x64"
		req.Version = "CoreOSUpdateEngine-0.1.0.0"
		req.InstallSource = "scheduler"
		app := req.AddApp(flatcarAppID, "500.0.0")
		app.MachineID = "modern-syncer"
		app.Track = group.ID
		app.AddUpdateCheck()
		app.MultiManifestOK = true // Modern syncer with multi-manifest support

		buf := bytes.NewBuffer(nil)
		err := xml.NewEncoder(buf).Encode(req)
		require.NoError(t, err)

		respBuf := bytes.NewBuffer(nil)
		err = h.Handle(buf, respBuf, "10.0.0.1")
		if err != nil {
			t.Fatalf("Failed to handle request: %v", err)
		}
		var resp omahaSpec.Response
		err = xml.NewDecoder(respBuf).Decode(&resp)
		require.NoError(t, err)

		// Should get update with multiple manifests
		require.Len(t, resp.Apps, 1)
		require.NotNil(t, resp.Apps[0].UpdateCheck)
		assert.Equal(t, omahaSpec.UpdateOK, resp.Apps[0].UpdateCheck.Status)
		assert.Len(t, resp.Apps[0].UpdateCheck.Manifests, 3) // 2 floors + 1 target
	})
}
