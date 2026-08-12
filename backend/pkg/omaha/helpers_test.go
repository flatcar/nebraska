package omaha

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// adminSvc returns an admin.Service that reuses a's shared read queries so
// tests can exercise admin write operations.
func adminSvc(a *api.API) *admin.Service {
	return admin.NewService(a.Conn(), a.Reads())
}

// runtimeSvc returns a runtime.Service that reuses a's shared read queries so
// tests can construct the Omaha handler over the runtime service.
func runtimeSvc(a *api.API) *runtime.Service {
	return runtime.NewService(a.Conn(), a.Reads(), runtime.Config{DisableUpdatesOnFailedRollout: true})
}

// setupOmahaFloorTest creates a complete Omaha floor test environment
func setupOmahaFloorTest(t *testing.T, a *api.API, name string, floorVersions []string, targetVersion string) (*types.Group, []*types.Package) {
	t.Helper()
	as := adminSvc(a)

	// Always use the flatcarAppID for Omaha tests
	tApp, err := a.GetApp(flatcarAppID)
	require.NoError(t, err)

	// Create all packages using shared helper
	allVersions := append(floorVersions, targetVersion)
	pkgs := make([]*types.Package, len(allVersions))
	for i, v := range allVersions {
		pkg, err := as.AddPackage(&types.Package{
			Type:          types.PkgTypeFlatcar,
			URL:           "http://sample.url/" + v,
			Version:       v,
			Filename:      null.StringFrom("flatcar_" + v + ".gz"),
			ApplicationID: tApp.ID,
			Arch:          types.ArchAMD64,
			FlatcarAction: &types.FlatcarAction{
				Event:                 "postinstall",
				Sha256:                "sha256-" + v,
				DisablePayloadBackoff: true,
			},
		})
		require.NoError(t, err)
		pkgs[i] = pkg
	}

	// Create channel with target (last package)
	channel, err := as.AddChannel(&types.Channel{
		Name:          name,
		ApplicationID: tApp.ID,
		PackageID:     null.StringFrom(pkgs[len(pkgs)-1].ID),
		Arch:          types.ArchAMD64,
	})
	require.NoError(t, err)

	// Add floors (all but last)
	for i := 0; i < len(pkgs)-1; i++ {
		err = as.AddChannelPackageFloor(channel.ID, pkgs[i].ID,
			null.StringFrom("Floor "+pkgs[i].Version))
		require.NoError(t, err)
	}

	// Create group with standard policy
	group, err := as.AddGroup(&types.Group{
		Name:                      name,
		ApplicationID:             tApp.ID,
		ChannelID:                 null.StringFrom(channel.ID),
		PolicyUpdatesEnabled:      true,
		PolicyPeriodInterval:      "15 minutes",
		PolicyMaxUpdatesPerPeriod: 100,
		PolicyUpdateTimeout:       "60 minutes",
	})
	require.NoError(t, err)

	return group, pkgs
}
