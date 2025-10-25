package syncer

import (
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
)

// setupSyncerTest sets up a standard syncer test environment
func setupSyncerTest(t *testing.T) (*Syncer, *api.API, *api.Group, *api.Channel) {
	t.Helper()
	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel
	require.NoError(t, syncer.initialize())

	return syncer, a, tGroup, tChannel
}

// createMultiManifestUpdate creates a standard multi-manifest update for testing
func createMultiManifestUpdate(versions ...string) *omaha.UpdateResponse {
	if len(versions) == 0 {
		versions = []string{"1000.0.0", "2000.0.0", "3000.0.0"}
	}

	manifests := make([]*omaha.Manifest, len(versions))
	for i, v := range versions {
		isFloor := i < len(versions)-1 // All but last are floors
		manifest := &omaha.Manifest{
			Version:  v,
			Packages: []*omaha.Package{{Name: "flatcar-" + v + ".gz", SHA1: "hash" + v[:4], Size: uint64(i+1) * 1000}},
			Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
			IsFloor:  isFloor,
		}
		if isFloor {
			manifest.FloorReason = "Floor " + v
		}
		manifests[i] = manifest
	}

	return &omaha.UpdateResponse{
		Status:    "ok",
		URLs:      []*omaha.URL{{CodeBase: "https://example.com"}},
		Manifests: manifests,
	}
}
