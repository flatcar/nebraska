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

// createSyncerUpdate builds a single flagged-manifest response as the upstream now
// serves it to a syncer: one package per response, tagged is_floor/is_target.
func createSyncerUpdate(version string, isFloor, isTarget bool) *omaha.UpdateResponse {
	m := &omaha.Manifest{
		Version:  version,
		Packages: []*omaha.Package{{Name: "flatcar-" + version + ".gz", SHA1: "hash-" + version, Size: 1000}},
		Actions:  []*omaha.Action{{Event: "postinstall", SHA256: "dGVzdHNoYTI1Ng=="}},
		IsFloor:  isFloor,
		IsTarget: isTarget,
	}
	if isFloor {
		m.FloorReason = "Floor " + version
	}
	return &omaha.UpdateResponse{
		Status:    "ok",
		URLs:      []*omaha.URL{{CodeBase: "https://example.com"}},
		Manifests: []*omaha.Manifest{m},
	}
}
