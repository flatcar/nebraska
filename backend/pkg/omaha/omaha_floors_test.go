package omaha

import (
	"bytes"
	"encoding/xml"
	"testing"

	omahaSpec "github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// TestFloorUpdateScenarios tests all floor-based update scenarios
func TestFloorUpdateScenarios(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	h := NewHandler(a)

	// Helper for syncer requests
	syncerRequest := func(h *Handler, version, group string, multiManifestOK bool) *omahaSpec.Response {
		req := omahaSpec.NewRequest()
		req.OS.Version = "3"
		req.OS.Platform = "CoreOS"
		req.OS.ServicePack = "linux"
		req.OS.Arch = "x64"
		req.Version = "CoreOSUpdateEngine-0.1.0.0"
		req.InstallSource = "scheduler"
		app := req.AddApp(flatcarAppID, version)
		app.MachineID = "syncer-" + version
		app.Track = group
		app.AddUpdateCheck()
		app.MultiManifestOK = multiManifestOK

		buf := bytes.NewBuffer(nil)
		err := xml.NewEncoder(buf).Encode(req)
		if err != nil {
			t.Fatalf("Failed to encode request: %v", err)
		}
		respBuf := bytes.NewBuffer(nil)
		err = h.Handle(buf, respBuf, "10.0.0.1")
		if err != nil {
			t.Fatalf("Failed to handle request: %v", err)
		}
		var resp omahaSpec.Response
		err = xml.NewDecoder(respBuf).Decode(&resp)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		return &resp
	}

	t.Run("RegularClientWithUpdate", func(t *testing.T) {
		group, pkgs := setupOmahaFloorTest(t, a, "regular", []string{"2000.0.0", "2500.0.0"}, "3000.0.0")
		require.NotNil(t, group)
		require.Len(t, pkgs, 3)

		tests := []struct {
			instance, expected string
		}{
			{"1500.0.0", "2000.0.0"}, // below floors -> floor1
			{"2000.0.0", "2500.0.0"}, // at floor1 -> floor2
			{"2200.0.0", "2500.0.0"}, // between floors -> floor2
			{"2500.0.0", "3000.0.0"}, // at floor2 -> target
			{"2700.0.0", "3000.0.0"}, // above floors -> target
		}

		for _, tc := range tests {
			resp := doOmahaRequest(t, h, flatcarAppID, tc.instance, "client-"+tc.instance, group.ID, "10.0.0.1", false, true, nil)
			require.NotNil(t, resp)
			filename := "flatcar_" + tc.expected + ".gz"
			url := "http://sample.url/" + tc.expected
			checkOmahaUpdateResponse(t, resp, tc.expected, filename, url, omahaSpec.UpdateOK)
			assert.Len(t, resp.Apps[0].UpdateCheck.Manifests, 1)
		}
	})

	t.Run("RegularClientNoUpdate", func(t *testing.T) {
		group, pkgs := setupOmahaFloorTest(t, a, "noupdate", []string{}, "5000.0.0")
		require.NotNil(t, group)
		require.Len(t, pkgs, 1)
		resp := doOmahaRequest(t, h, flatcarAppID, "5000.0.0", "client", group.ID, "10.0.0.1", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "", "", "", omahaSpec.NoUpdate)
	})

	t.Run("SyncerMultiManifest", func(t *testing.T) {
		group, pkgs := setupOmahaFloorTest(t, a, "syncer", []string{"9000.0.0", "9500.0.0"}, "10000.0.0")
		require.NotNil(t, group)
		require.Len(t, pkgs, 3)

		testCases := map[string]int{
			"8500.0.0": 3, // below floors: all manifests
			"9200.0.0": 2, // between: floor2 + target
			"9700.0.0": 1, // above: target only
		}

		for version, expectedManifests := range testCases {
			resp := syncerRequest(h, version, group.ID, true)
			assert.Len(t, resp.Apps[0].UpdateCheck.Manifests, expectedManifests)
			// Verify last is target
			last := resp.Apps[0].UpdateCheck.Manifests[expectedManifests-1]
			assert.True(t, last.IsTarget)
			// Verify others are floors
			for i := 0; i < expectedManifests-1; i++ {
				assert.True(t, resp.Apps[0].UpdateCheck.Manifests[i].IsFloor)
			}
		}
	})

	t.Run("OldSyncerBlocked", func(t *testing.T) {
		group, pkgs := setupOmahaFloorTest(t, a, "oldsyncer", []string{"6000.0.0"}, "7000.0.0")
		require.NotNil(t, group)
		require.Len(t, pkgs, 2)
		resp := syncerRequest(h, "1500.0.0", group.ID, false) // multiPkgOK=false
		checkOmahaUpdateResponse(t, resp, "", "", "", omahaSpec.NoUpdate)
	})

	t.Run("ModernSyncerNoUpdate", func(t *testing.T) {
		group, pkgs := setupOmahaFloorTest(t, a, "syncernoup", []string{}, "8000.0.0")
		require.NotNil(t, group)
		require.Len(t, pkgs, 1)
		resp := syncerRequest(h, "8000.0.0", group.ID, true) // At target version
		checkOmahaUpdateResponse(t, resp, "", "", "", omahaSpec.NoUpdate)
	})

	t.Run("NoFloors", func(t *testing.T) {
		// Setup without floors
		group, pkgs := setupOmahaFloorTest(t, a, "nofloor", []string{}, "4000.0.0")
		require.NotNil(t, group)
		require.Len(t, pkgs, 1)

		// Regular client gets direct update
		resp := doOmahaRequest(t, h, flatcarAppID, "1000.0.0", "client", group.ID, "10.0.0.1", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "4000.0.0", "flatcar_4000.0.0.gz", "http://sample.url/4000.0.0", omahaSpec.UpdateOK)

		// Syncer gets single manifest
		resp = syncerRequest(h, "1000.0.0", group.ID, true)
		assert.Len(t, resp.Apps[0].UpdateCheck.Manifests, 1)
		assert.True(t, resp.Apps[0].UpdateCheck.Manifests[0].IsTarget)
	})

	t.Run("TargetAsFloor", func(t *testing.T) {
		// Setup where target is also a floor (critical mandatory version)
		group, pkgs := setupOmahaFloorTest(t, a, "targetfloor", []string{"11000.0.0", "12000.0.0"}, "13000.0.0")
		require.NotNil(t, group)
		require.Len(t, pkgs, 3)

		// Mark the target as ALSO being a floor
		err := a.AddChannelPackageFloor(group.ChannelID.String, pkgs[2].ID,
			null.StringFrom("Critical mandatory version"))
		require.NoError(t, err)

		// Regular client below all versions
		resp := doOmahaRequest(t, h, flatcarAppID, "10000.0.0", "client-low", group.ID, "10.0.0.1", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "11000.0.0", "flatcar_11000.0.0.gz", "http://sample.url/11000.0.0", omahaSpec.UpdateOK)

		// Regular client between floors
		resp = doOmahaRequest(t, h, flatcarAppID, "11500.0.0", "client-mid", group.ID, "10.0.0.2", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "12000.0.0", "flatcar_12000.0.0.gz", "http://sample.url/12000.0.0", omahaSpec.UpdateOK)

		// Regular client above regular floors but below target-floor
		resp = doOmahaRequest(t, h, flatcarAppID, "12500.0.0", "client-high", group.ID, "10.0.0.3", false, true, nil)
		checkOmahaUpdateResponse(t, resp, "13000.0.0", "flatcar_13000.0.0.gz", "http://sample.url/13000.0.0", omahaSpec.UpdateOK)

		// Syncer should get all manifests with correct flags
		resp = syncerRequest(h, "10000.0.0", group.ID, true)
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
