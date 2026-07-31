package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
)

// flatcarAppID is the well-known Flatcar application UUID seeded into every
// database.
const flatcarAppID = types.FlatcarAppID

// adminSvc returns an admin.Service that reuses a's shared read queries so
// tests can exercise admin write operations (which moved to package admin).
func adminSvc(a *API) *admin.Service {
	return admin.NewService(a.Reads())
}

// runtimeSvc returns a runtime.Service that reuses a's shared read queries so
// tests can exercise the runtime operations (which moved to package
// runtime).
func runtimeSvc(a *API) *runtime.Service {
	return runtime.NewService(a.Reads(), runtime.Config{DisableUpdatesOnFailedRollout: true})
}

// Test helper functions to reduce boilerplate in floor tests

// quickPkg creates a Flatcar package with minimal setup
func quickPkg(t *testing.T, a *API, appID, version string) *Package {
	t.Helper()
	as := adminSvc(a)
	pkg, err := as.AddPackage(&Package{
		Type:          PkgTypeFlatcar,
		URL:           "http://sample.url/" + version,
		Version:       version,
		Filename:      null.StringFrom("flatcar_" + version + ".gz"),
		ApplicationID: appID,
		Arch:          ArchAMD64,
		FlatcarAction: &FlatcarAction{
			Event:                 "postinstall",
			Sha256:                "sha256-" + version,
			DisablePayloadBackoff: true,
		},
	})
	require.NoError(t, err)
	return pkg
}

// quickPkgs creates multiple packages at once
func quickPkgs(t *testing.T, a *API, appID string, versions ...string) []*Package {
	t.Helper()
	pkgs := make([]*Package, len(versions))
	for i, v := range versions {
		pkgs[i] = quickPkg(t, a, appID, v)
	}
	return pkgs
}

// floorTestSetup creates a complete floor test scenario
type floorTestSetup struct {
	Channel *Channel
	Group   *Group
	Floors  []*Package
	Target  *Package
	AppID   string
}

// setupFloors creates a standard floor test configuration
func setupFloors(t *testing.T, a *API, name string, floorVersions []string, targetVersion string) *floorTestSetup {
	t.Helper()
	as := adminSvc(a)

	// Use existing Flatcar app if available, otherwise create test app
	tApp, err := a.GetApp(flatcarAppID)
	if err != nil {
		tTeam, _ := as.AddTeam(&Team{Name: "test_team_" + name})
		tApp, _ = as.AddApp(&Application{Name: "test_app_" + name, TeamID: tTeam.ID})
	}

	// Create all packages
	allVersions := append(floorVersions, targetVersion)
	pkgs := quickPkgs(t, a, tApp.ID, allVersions...)

	// Split into floors and target
	floors := pkgs[:len(floorVersions)]
	target := pkgs[len(pkgs)-1]

	// Create channel with target
	channel, err := as.AddChannel(&Channel{
		Name:          name,
		ApplicationID: tApp.ID,
		PackageID:     null.StringFrom(target.ID),
		Arch:          ArchAMD64,
	})
	require.NoError(t, err)

	// Add floors with reasons
	for i, floor := range floors {
		err := as.AddChannelPackageFloor(channel.ID, floor.ID,
			null.StringFrom(fmt.Sprintf("Floor %d", i+1)))
		require.NoError(t, err)
	}

	// Create group
	group, err := as.AddGroup(&Group{
		Name:                      name,
		ApplicationID:             tApp.ID,
		ChannelID:                 null.StringFrom(channel.ID),
		PolicyUpdatesEnabled:      true,
		PolicyPeriodInterval:      "15 minutes",
		PolicyMaxUpdatesPerPeriod: 100,
		PolicyUpdateTimeout:       "60 minutes",
	})
	require.NoError(t, err)

	return &floorTestSetup{
		Channel: channel,
		Group:   group,
		Floors:  floors,
		Target:  target,
		AppID:   tApp.ID,
	}
}
