package syncer

import (
	"log"
	"os"
	"testing"

	"github.com/kinvolk/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

const (
	defaultTestDbURL string = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"
)

func newAPI(t *testing.T) *api.API {
	t.Helper()
	a, err := api.NewForTest(api.OptionInitDB, api.OptionDisableUpdatesOnFailedRollout)

	t.Logf("Failed to init DB: %v\n", err)
	t.Log("These tests require PostgreSQL running and a tests database created, please adjust NEBRASKA_DB_URL as needed.")
	require.NoError(t, err)

	return a
}

func newForTest(t *testing.T, conf *Config) *Syncer {
	t.Helper()
	a := newAPI(t)

	if conf.API == nil {
		conf.API = a
	}
	s, err := New(conf)
	require.NoError(t, err)

	return s
}

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	if _, ok := os.LookupEnv("NEBRASKA_DB_URL"); !ok {
		log.Printf("NEBRASKA_DB_URL not set, setting to default %q\n", defaultTestDbURL)
		_ = os.Setenv("NEBRASKA_DB_URL", defaultTestDbURL)
	}

	os.Exit(m.Run())
}

func TestSyncer_NoAPI(t *testing.T) {
	_, err := New(&Config{})
	assert.ErrorIs(t, err, ErrInvalidAPIInstance)
}

func TestSyncer_InvalidPkgsURL(t *testing.T) {
	a := newAPI(t)
	t.Cleanup(func() {
		a.Close()
	})

	tests := []struct {
		url   string
		isErr bool
	}{
		{
			url:   "",
			isErr: false,
		},
		{
			url:   ":file",
			isErr: true,
		},
		{
			url:   "https://myphony.url",
			isErr: false,
		},
		{
			url:   "file:///my/file",
			isErr: false,
		},
	}

	for _, tc := range tests {
		testCase := tc
		t.Run(testCase.url, func(t *testing.T) {
			t.Parallel()

			_, err := New(&Config{
				API:         a,
				PackagesURL: testCase.url,
			})
			if testCase.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncer_Init(t *testing.T) {
	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() {
		syncer.api.Close()
	})

	tApp, err := a.GetApp(flatcarAppID)
	require.NoError(t, err)
	tPkg, err := a.AddPackage(&api.Package{Type: api.PkgTypeFlatcar, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: api.ArchAMD64})
	require.NoError(t, err)
	groupID, err := a.GetGroupID(flatcarAppID, "stable", tPkg.Arch)
	require.NoError(t, err)

	tGroup, err := a.GetGroup(groupID)
	require.NoError(t, err)

	tChannel := tGroup.Channel

	tChannel.PackageID = null.StringFrom(tPkg.ID)

	err = a.UpdateChannel(tChannel)
	require.NoError(t, err)

	err = syncer.initialize()
	require.NoError(t, err)

	desc := channelDescriptor{
		name: tChannel.Name,
		arch: tChannel.Arch,
	}

	version, ok := syncer.versions[desc]
	assert.True(t, ok)
	assert.Equal(t, tPkg.Version, version)
}

func createOmahaUpdate() *omaha.UpdateResponse {
	return &omaha.UpdateResponse{
		URLs: []*omaha.URL{
			{CodeBase: "https://example.com"},
		},
		Manifest: &omaha.Manifest{
			Version: "1.2.3",
			Packages: []*omaha.Package{
				{
					Name: "updatepayload.tgz",
					SHA1: "00000000000000000",
				},
			},
			Actions: []*omaha.Action{
				{},
			},
		},
	}
}

func setupFlatcarAppStableGroup(t *testing.T, a *api.API) *api.Group {
	t.Helper()
	tApp, err := a.GetApp(flatcarAppID)
	require.NoError(t, err)
	tPkg, err := a.AddPackage(&api.Package{Type: api.PkgTypeFlatcar, URL: "http://sample.url/pkg", Version: "0.1.0", ApplicationID: tApp.ID, Arch: api.ArchAMD64})
	require.NoError(t, err)
	groupID, err := a.GetGroupID(flatcarAppID, "stable", tPkg.Arch)
	require.NoError(t, err)

	tGroup, err := a.GetGroup(groupID)
	require.NoError(t, err)

	tChannel := tGroup.Channel

	tChannel.PackageID = null.StringFrom(tPkg.ID)

	return tGroup
}

func TestSyncer_GetPackage(t *testing.T) {
	syncer := newForTest(t, &Config{})
	a := syncer.api
	t.Cleanup(func() {
		a.Close()
	})

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel

	err := syncer.initialize()
	require.NoError(t, err)

	update := createOmahaUpdate()

	desc := channelDescriptor{
		name: tChannel.Name,
		arch: tChannel.Arch,
	}
	err = syncer.processUpdate(desc, update)
	require.NoError(t, err)

	// Get updated group
	tGroup, err = a.GetGroup(tGroup.ID)
	require.NoError(t, err)

	assert.Equal(t, update.Manifest.Version, tGroup.Channel.Package.Version)
	assert.Equal(t, update.URLs[0].CodeBase, tGroup.Channel.Package.URL)
	assert.Equal(t, update.Manifest.Packages[0].Name, tGroup.Channel.Package.Filename.String)
}

func TestSyncer_GetPackageWithDiffURL(t *testing.T) {
	conf := &Config{
		PackagesURL: "https://my.super.different.packagesurl.io/bucket/",
	}
	syncer := newForTest(t, conf)
	a := syncer.api
	t.Cleanup(func() {
		a.Close()
	})

	tGroup := setupFlatcarAppStableGroup(t, a)
	tChannel := tGroup.Channel

	err := syncer.initialize()
	require.NoError(t, err)

	update := createOmahaUpdate()

	desc := channelDescriptor{
		name: tChannel.Name,
		arch: tChannel.Arch,
	}
	err = syncer.processUpdate(desc, update)
	require.NoError(t, err)

	// Get updated group
	tGroup, err = a.GetGroup(tGroup.ID)
	require.NoError(t, err)

	assert.Equal(t, update.Manifest.Version, tGroup.Channel.Package.Version)
	assert.Equal(t, conf.PackagesURL, tGroup.Channel.Package.URL)
	assert.Equal(t, update.Manifest.Packages[0].Name, tGroup.Channel.Package.Filename.String)
}
