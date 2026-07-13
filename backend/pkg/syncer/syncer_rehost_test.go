package syncer

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSyncer_RehostsMissingHostedFile verifies that when Nebraska hosts packages
// locally and an already-recorded package's hosted file has gone missing on disk,
// getOrCreatePackage restores it by re-downloading - without erroring on the existing
// DB row and without touching packages whose file is present.
func TestSyncer_RehostsMissingHostedFile(t *testing.T) {
	payload := []byte("test flatcar hosted payload")
	sha1sum := sha1.Sum(payload)
	sha256sum := sha256.Sum256(payload)
	sha1b64 := base64.StdEncoding.EncodeToString(sha1sum[:])
	sha256b64 := base64.StdEncoding.EncodeToString(sha256sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	syncer := newForTest(t, &Config{HostPackages: true, PackagesPath: tmp})
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel
	require.NoError(t, syncer.initialize())
	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}

	const version = "9999.0.0"
	update := &omaha.UpdateResponse{
		Status: "ok",
		URLs:   []*omaha.URL{{CodeBase: srv.URL}},
		Manifests: []*omaha.Manifest{{
			Version:  version,
			Packages: []*omaha.Package{{Name: "update.gz", SHA1: sha1b64, SHA256: sha256b64, Size: uint64(len(payload))}},
			Actions:  []*omaha.Action{{Event: "postinstall", SHA256: sha256b64}},
			IsTarget: true,
		}},
	}
	manifest := update.Manifests[0]
	hosted := filepath.Join(tmp, fmt.Sprintf("flatcar-%s-%s.gz", getArchString(desc.arch), version))

	// First pass creates the package and hosts the payload locally.
	pkg, err := syncer.getOrCreatePackage(desc, manifest, update)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	require.FileExists(t, hosted)

	// File present: getOrCreatePackage must be a no-op for hosting (no re-download,
	// no error, file untouched).
	_, err = syncer.getOrCreatePackage(desc, manifest, update)
	require.NoError(t, err)
	require.FileExists(t, hosted)

	// Simulate an out-of-band loss of the hosted file.
	require.NoError(t, os.Remove(hosted))

	// getOrCreatePackage must notice the missing file and re-download it, without
	// erroring on the existing DB row.
	pkg, err = syncer.getOrCreatePackage(desc, manifest, update)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	require.FileExists(t, hosted)

	got, err := os.ReadFile(hosted)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
}

// TestSyncer_RehostIsNoopWhenNotHosting verifies the restore path never touches the
// filesystem when hosting is disabled.
func TestSyncer_RehostIsNoopWhenNotHosting(t *testing.T) {
	syncer := newForTest(t, &Config{}) // HostPackages defaults to false
	a := syncer.api
	t.Cleanup(func() { a.Close() })

	desc := channelDescriptor{name: "stable", arch: api.ArchAMD64}
	update := &omaha.UpdateResponse{
		URLs:      []*omaha.URL{{CodeBase: "https://example.com"}},
		Manifests: []*omaha.Manifest{{Version: "1.2.3"}},
	}
	pkg := &api.Package{Version: "1.2.3"}

	require.NoError(t, syncer.rehostPackageFilesIfMissing(desc, update.Manifests[0], update, pkg))
}
