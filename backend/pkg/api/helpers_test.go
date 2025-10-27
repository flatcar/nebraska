package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// Test helper functions to reduce boilerplate in floor tests

// quickPkg creates a Flatcar package with minimal setup
func quickPkg(t *testing.T, a *API, appID, version string) *Package {
	t.Helper()
	pkg, err := a.AddPackage(&Package{
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

	// Use existing Flatcar app if available, otherwise create test app
	tApp, err := a.GetApp("e96281a6-d1af-4bde-9a0a-97b76e56dc57") // flatcarAppID
	if err != nil {
		tTeam, _ := a.AddTeam(&Team{Name: "test_team_" + name})
		tApp, _ = a.AddApp(&Application{Name: "test_app_" + name, TeamID: tTeam.ID})
	}

	// Create all packages
	allVersions := append(floorVersions, targetVersion)
	pkgs := quickPkgs(t, a, tApp.ID, allVersions...)

	// Split into floors and target
	floors := pkgs[:len(floorVersions)]
	target := pkgs[len(pkgs)-1]

	// Create channel with target
	channel, err := a.AddChannel(&Channel{
		Name:          name,
		ApplicationID: tApp.ID,
		PackageID:     null.StringFrom(target.ID),
		Arch:          ArchAMD64,
	})
	require.NoError(t, err)

	// Add floors with reasons
	for i, floor := range floors {
		err := a.AddChannelPackageFloor(channel.ID, floor.ID,
			null.StringFrom(fmt.Sprintf("Floor %d", i+1)))
		require.NoError(t, err)
	}

	// Create group
	group, err := a.AddGroup(&Group{
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
