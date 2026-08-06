package api

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// TestDeleteFloorPackageBlocked verifies that a package marked as a floor for
// any channel cannot be deleted until it's unmarked first. Without this guard,
// deleting a floor package would silently remove the channel_package_floors
// row (ON DELETE CASCADE) and erase the mandatory-upgrade-path guarantee for
// any instance still below that version.
func TestDeleteFloorPackageBlocked(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	setup := setupFloors(t, a, "delete-floor-blocked", []string{"2000.0.0"}, "3000.0.0")
	floor := setup.Floors[0]

	err := a.DeletePackage(floor.ID)
	assert.Equal(t, ErrPackageIsFloor, err, "deleting a package marked as floor must be rejected")

	// The package must still exist and still be a floor after the rejected delete.
	pkgs, err := a.GetChannelFloorPackages(setup.Channel.ID)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	assert.Equal(t, floor.ID, pkgs[0].ID)

	// Once unmarked as a floor, deletion must succeed.
	require.NoError(t, a.RemoveChannelPackageFloor(setup.Channel.ID, floor.ID))
	assert.NoError(t, a.DeletePackage(floor.ID))

	pkgs, err = a.GetChannelFloorPackages(setup.Channel.ID)
	require.NoError(t, err)
	assert.Empty(t, pkgs)
}

// TestOrphanedFloorAfterTargetPackageDeleted covers the "missing package" edge
// case: the channel's target package gets removed (channel.package_id is
// FK'd ON DELETE SET NULL, unlike floors which block deletion), leaving
// channel_package_floors rows pointing at a channel with no target. Update
// requests against that channel must fail gracefully instead of panicking or
// returning a bogus package.
func TestOrphanedFloorAfterTargetPackageDeleted(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	setup := setupFloors(t, a, "orphaned-target", []string{"2000.0.0"}, "3000.0.0")

	// The target itself isn't a floor, so deleting it is allowed and the
	// channel's package_id is set to NULL by the FK.
	require.NoError(t, a.DeletePackage(setup.Target.ID))

	channel, err := a.GetChannel(setup.Channel.ID)
	require.NoError(t, err)
	assert.Nil(t, channel.Package, "channel target should be nil after target package deletion")

	// The floor row is untouched (only the target FK was affected).
	floors, err := a.GetChannelFloorPackages(setup.Channel.ID)
	require.NoError(t, err)
	require.Len(t, floors, 1)

	_, err = a.GetUpdatePackage(Instance{ID: "orphan-instance", IP: "10.0.0.1"},
		NewInstanceApplication(setup.AppID, setup.Group.ID, "1000.0.0"))
	assert.Equal(t, ErrNoPackageFound, err, "should fail gracefully when channel has no target package")

	_, err = a.GetUpdatePackagesForSyncer(Instance{ID: "orphan-syncer", IP: "10.0.0.1"},
		NewInstanceApplication(setup.AppID, setup.Group.ID, "1000.0.0"))
	assert.Equal(t, ErrNoPackageFound, err, "syncer path should also fail gracefully with no target package")
}

// TestFloorAboveTargetIgnored covers a misconfiguration edge case: nothing
// stops AddChannelPackageFloor from marking a package with a version *higher*
// than the channel's current target as a floor (e.g. stale data from a
// channel target rollback). If the read path didn't bound floors to
// (instanceVersion, targetVersion], an instance could get wedged forever
// waiting for a "floor" beyond the target it's supposed to reach - the
// closest failure mode to a circular/impossible-to-satisfy prerequisite this
// schema can produce, since floors have no explicit dependency graph.
func TestFloorAboveTargetIgnored(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	setup := setupFloors(t, a, "floor-above-target", []string{"2000.0.0"}, "3000.0.0")

	bogusFloor := quickPkg(t, a, setup.AppID, "4000.0.0")
	require.NoError(t, a.AddChannelPackageFloor(setup.Channel.ID, bogusFloor.ID, null.StringFrom("above target")))

	instanceID := "floor-above-target-instance"

	// Below the legitimate floor: must get 2000.0.0, never the bogus 4000.0.0.
	pkg, err := a.GetUpdatePackage(Instance{ID: instanceID, IP: "10.0.0.1"},
		NewInstanceApplication(setup.AppID, setup.Group.ID, "1000.0.0"))
	require.NoError(t, err)
	assert.Equal(t, "2000.0.0", pkg.Version)

	// Reached the legitimate floor: must jump straight to target (3000.0.0),
	// completely skipping the out-of-range bogus floor.
	_, err = a.GetUpdatePackage(Instance{ID: instanceID, IP: "10.0.0.1"},
		NewInstanceApplication(setup.AppID, setup.Group.ID, "2000.0.0"))
	assert.Equal(t, ErrNoUpdatePackageAvailable, err)

	pkg, err = a.GetUpdatePackage(Instance{ID: instanceID, IP: "10.0.0.1"},
		NewInstanceApplication(setup.AppID, setup.Group.ID, "2000.0.0"))
	require.NoError(t, err)
	assert.Equal(t, "3000.0.0", pkg.Version, "must reach target directly, bogus above-target floor must never be offered")

	// Even a syncer, which asks for the full floor+target set, must never see it.
	packages, err := a.GetUpdatePackagesForSyncer(Instance{ID: "floor-above-target-syncer", IP: "10.0.0.1"},
		NewInstanceApplication(setup.AppID, setup.Group.ID, "0.0.0"))
	require.NoError(t, err)
	for _, p := range packages {
		assert.NotEqual(t, "4000.0.0", p.Version, "bogus above-target floor must be excluded from syncer manifests too")
	}
}

// TestManyFloorsExceedingResponseLimitProgression covers the "many
// prerequisites" edge case for regular (non-syncer) clients: when the number
// of required floors exceeds NEBRASKA_MAX_FLOORS_PER_RESPONSE, a single
// instance progressing one grant at a time must still walk through every
// floor in order and eventually reach the target - the per-response limit
// only bounds how many floors are fetched in one query, not how many an
// instance is allowed to complete.
func TestManyFloorsExceedingResponseLimitProgression(t *testing.T) {
	oldMax := os.Getenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE")
	defer os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", oldMax)
	require.NoError(t, os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", "3"))

	a := newForTest(t)
	defer a.Close()

	floorVersions := []string{"1000.0.0", "2000.0.0", "3000.0.0", "4000.0.0", "5000.0.0"}
	setup := setupFloors(t, a, "many-floors-regular", floorVersions, "6000.0.0")

	instanceID := "many-floors-instance"
	appID := setup.AppID
	groupID := setup.Group.ID
	currentVersion := "0.0.0"

	expectedProgression := append(append([]string{}, floorVersions...), "6000.0.0")
	for _, expected := range expectedProgression {
		pkg, err := a.GetUpdatePackage(Instance{ID: instanceID, IP: "10.0.0.1"},
			NewInstanceApplication(appID, groupID, currentVersion))
		require.NoError(t, err, "should still get a package for step %s even though limit is 3", expected)
		assert.Equal(t, expected, pkg.Version)

		currentVersion = expected
		_, err = a.GetUpdatePackage(Instance{ID: instanceID, IP: "10.0.0.1"},
			NewInstanceApplication(appID, groupID, currentVersion))
		assert.Equal(t, ErrNoUpdatePackageAvailable, err, "should complete after reaching %s", expected)
	}

	// Fully progressed: no more updates available.
	_, err := a.GetUpdatePackage(Instance{ID: instanceID, IP: "10.0.0.1"},
		NewInstanceApplication(appID, groupID, "6000.0.0"))
	assert.Equal(t, ErrNoUpdatePackageAvailable, err)
}

func TestSyncerFloorLimitTruncationKnownLimitation(t *testing.T) {
	oldMax := os.Getenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE")
	defer os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", oldMax)
	require.NoError(t, os.Setenv("NEBRASKA_MAX_FLOORS_PER_RESPONSE", "3"))

	a := newForTest(t)
	defer a.Close()

	floorVersions := []string{"1000.0.0", "2000.0.0", "3000.0.0", "4000.0.0", "5000.0.0"}
	setup := setupFloors(t, a, "many-floors-syncer", floorVersions, "6000.0.0")

	packages, err := a.GetUpdatePackagesForSyncer(Instance{ID: "truncated-syncer", IP: "10.0.0.1"},
		NewInstanceApplication(setup.AppID, setup.Group.ID, "0.0.0"))
	require.NoError(t, err)

	gotVersions := make([]string, len(packages))
	for i, p := range packages {
		gotVersions[i] = p.Version
	}

	assert.Equal(t, []string{"1000.0.0", "2000.0.0", "3000.0.0", "6000.0.0"}, gotVersions)

	instance, err := a.GetInstance("truncated-syncer", setup.AppID)
	require.NoError(t, err)
	assert.Equal(t, "6000.0.0", instance.Application.LastUpdateVersion.String)
	assert.Equal(t, InstanceStatusUpdateGranted, int(instance.Application.Status.Int64))
}
