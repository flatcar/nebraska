package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// TestFloorOperations tests basic floor CRUD operations
func TestFloorOperations(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Setup using helpers
	setup := setupFloors(t, a, "test", []string{"1000.0.0", "2000.0.0"}, "3000.0.0")

	// Update floor reason for first floor
	assert.NoError(t, a.RemoveChannelPackageFloor(setup.Channel.ID, setup.Floors[0].ID))
	assert.NoError(t, a.AddChannelPackageFloor(setup.Channel.ID, setup.Floors[0].ID, null.StringFrom("Filesystem upgrade")))

	// Test wrong arch
	tTeam, err := a.AddTeam(&Team{Name: "test_team_arch"})
	assert.NoError(t, err)
	tApp, err := a.AddApp(&Application{Name: "test_app_arch", TeamID: tTeam.ID})
	assert.NoError(t, err)
	pkgWrongArch, err := a.AddPackage(&Package{
		Type:          PkgTypeFlatcar,
		URL:           "http://sample.url/1500.0.0",
		Version:       "1500.0.0",
		ApplicationID: tApp.ID,
		Arch:          ArchAArch64,
	})
	assert.NoError(t, err)
	assert.Equal(t, ErrArchMismatch, a.AddChannelPackageFloor(setup.Channel.ID, pkgWrongArch.ID, null.String{}))

	// Test getting floors
	floors, err := a.GetChannelFloorPackages(setup.Channel.ID)
	assert.NoError(t, err)
	assert.Len(t, floors, 2)
	assert.Equal(t, "1000.0.0", floors[0].Version)
	assert.Equal(t, "Filesystem upgrade", floors[0].FloorReason.String)
	assert.True(t, floors[0].IsFloor)

	// Test required floors between versions
	testCases := map[string]int{
		"500.0.0":  2, // below all floors
		"1500.0.0": 1, // between floors
		"2500.0.0": 0, // above all floors
		"3000.0.0": 0, // at target
	}

	for instance, expected := range testCases {
		ch, err := a.GetChannel(setup.Channel.ID)
		assert.NoError(t, err)
		floors, err := a.GetRequiredChannelFloors(ch, instance)
		assert.NoError(t, err)
		assert.Len(t, floors, expected, "instance %s", instance)
	}

	// Test removing floor
	assert.NoError(t, a.RemoveChannelPackageFloor(setup.Channel.ID, setup.Floors[0].ID))
	floors, err = a.GetChannelFloorPackages(setup.Channel.ID)
	assert.NoError(t, err)
	assert.Len(t, floors, 1)
	assert.Equal(t, ErrNoRowsAffected, a.RemoveChannelPackageFloor(setup.Channel.ID, setup.Floors[0].ID))
}

// TestFloorsUncapped verifies all in-range floors are returned (the per-response cap was removed)
func TestFloorsUncapped(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// All in-range floors are returned; the per-response cap was removed (syncers
	// walk floors one at a time). Use more than the old default cap (5) so a
	// re-introduced limit would fail this test.
	floorVersions := []string{"1000.0.0", "2000.0.0", "3000.0.0", "4000.0.0", "5000.0.0", "6000.0.0", "7000.0.0"}
	setup := setupFloors(t, a, "maxtest", floorVersions, "8000.0.0")

	ch, err := a.GetChannel(setup.Channel.ID)
	assert.NoError(t, err)
	floors, err := a.GetRequiredChannelFloors(ch, "0.0.0")
	assert.NoError(t, err)
	assert.Len(t, floors, 7)
}

// TestFloorPagination tests paginated floor retrieval
func TestFloorPagination(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Setup floors - no target needed for this test
	floorVersions := []string{"100.0.0", "200.0.0", "300.0.0", "400.0.0", "500.0.0"}
	setup := setupFloors(t, a, "pagination", floorVersions, "600.0.0")

	// Test count
	count, err := a.GetChannelFloorPackagesCount(setup.Channel.ID)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)

	// Test pagination
	floors, err := a.GetChannelFloorPackagesPaginated(setup.Channel.ID, 1, 3)
	assert.NoError(t, err)
	assert.Len(t, floors, 3)
	assert.Equal(t, "100.0.0", floors[0].Version)

	floors, err = a.GetChannelFloorPackagesPaginated(setup.Channel.ID, 2, 3)
	assert.NoError(t, err)
	assert.Len(t, floors, 2)
	assert.Equal(t, "400.0.0", floors[0].Version)
}

// TestNonStandardVersions tests floors with non-standard Flatcar versions
func TestNonStandardVersions(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Setup with non-standard versions
	floorVersions := []string{"3510.2.0+test", "3602.2.1-lts", "3760.2.0"}
	setup := setupFloors(t, a, "nonstandard", floorVersions, "3815.2.0-beta")

	testCases := map[string]int{
		"3400.0.0":        3, // below all
		"3550.0.0+custom": 2, // between floors
		"3800.0.0-test":   0, // above all
	}

	for instance, expected := range testCases {
		ch, err := a.GetChannel(setup.Channel.ID)
		assert.NoError(t, err)
		floors, err := a.GetRequiredChannelFloors(ch, instance)
		assert.NoError(t, err)
		assert.Len(t, floors, expected, "instance %s", instance)
	}
}

// TestFloorReason tests floor reason persistence
func TestFloorReason(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Setup with one floor
	setup := setupFloors(t, a, "reason", []string{"1000.0.0"}, "2000.0.0")
	pkg := setup.Floors[0]
	channel := setup.Channel

	// Remove existing floor to test fresh add
	err := a.RemoveChannelPackageFloor(channel.ID, pkg.ID)
	require.NoError(t, err)

	// Add with reason
	reason := "Critical boot partition restructuring"
	err = a.AddChannelPackageFloor(channel.ID, pkg.ID, null.StringFrom(reason))
	assert.NoError(t, err)

	floors, err := a.GetChannelFloorPackages(channel.ID)
	assert.NoError(t, err)
	assert.Equal(t, reason, floors[0].FloorReason.String)

	// Update reason (UPSERT)
	newReason := "Updated: Filesystem upgrade"
	err = a.AddChannelPackageFloor(channel.ID, pkg.ID, null.StringFrom(newReason))
	assert.NoError(t, err)

	floors, err = a.GetChannelFloorPackages(channel.ID)
	assert.NoError(t, err)
	assert.Equal(t, newReason, floors[0].FloorReason.String)
}

// TestFloorRolloutPolicy tests floors respect rollout policies
func TestFloorRolloutPolicy(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Setup with one floor
	setup := setupFloors(t, a, "policy", []string{"2000.0.0"}, "3000.0.0")

	// Update group to have restricted policy
	group := setup.Group
	group.PolicyMaxUpdatesPerPeriod = 1 // Only 1 update allowed
	err := a.UpdateGroup(group)
	assert.NoError(t, err)

	// First client gets floor
	pkg1, err := a.GetUpdatePackage(Instance{ID: "i1", IP: "10.0.0.1"}, NewInstanceApplication(setup.AppID, group.ID, "1000.0.0"))
	assert.NoError(t, err)
	assert.Equal(t, "2000.0.0", pkg1.Version)

	// Second client blocked by policy
	_, err = a.GetUpdatePackage(Instance{ID: "i2", IP: "10.0.0.2"}, NewInstanceApplication(setup.AppID, group.ID, "1000.0.0"))
	assert.Equal(t, ErrMaxUpdatesPerPeriodLimitReached, err)
}

// TestTargetAsFloor tests when target package is also marked as a floor
func TestTargetAsFloor(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	// Setup with target also being a floor
	// This represents a critical version that MUST be installed
	// and also becomes the new channel target
	setup := setupFloors(t, a, "targetfloor", []string{"1000.0.0", "2000.0.0"}, "3000.0.0")

	// Mark the target as ALSO being a floor (critical version)
	err := a.AddChannelPackageFloor(setup.Channel.ID, setup.Target.ID,
		null.StringFrom("Filesystem support for usr dir - mandatory"))
	assert.NoError(t, err)

	// Verify target is marked as floor
	floors, err := a.GetChannelFloorPackages(setup.Channel.ID)
	assert.NoError(t, err)
	assert.Len(t, floors, 3) // All three are floors now
	assert.Equal(t, "3000.0.0", floors[2].Version)
	assert.Equal(t, "Filesystem support for usr dir - mandatory", floors[2].FloorReason.String)

	// Test required floors for different client versions
	testCases := map[string]struct {
		expectedCount int
		expectedLast  string
	}{
		"500.0.0":  {3, "3000.0.0"}, // Gets all 3 floors including target
		"1500.0.0": {2, "3000.0.0"}, // Gets 2000 and 3000 (both floors)
		"2500.0.0": {1, "3000.0.0"}, // Gets only 3000 (target-floor)
		"3000.0.0": {0, ""},         // At target, no update needed
	}

	for instance, expected := range testCases {
		ch, err := a.GetChannel(setup.Channel.ID)
		assert.NoError(t, err)
		floors, err := a.GetRequiredChannelFloors(ch, instance)
		assert.NoError(t, err)
		assert.Len(t, floors, expected.expectedCount, "instance %s", instance)
		if expected.expectedCount > 0 {
			// Verify last floor is always the target (3000.0.0)
			lastFloor := floors[expected.expectedCount-1]
			assert.Equal(t, expected.expectedLast, lastFloor.Version,
				"instance %s should have target-floor as last", instance)
			assert.True(t, lastFloor.IsFloor)
		}
	}

	// Test that regular client gets the appropriate update
	pkg, err := a.GetUpdatePackage(Instance{ID: "i1", IP: "10.0.0.1"}, NewInstanceApplication(setup.AppID, setup.Group.ID, "500.0.0"))
	assert.NoError(t, err)
	assert.Equal(t, "1000.0.0", pkg.Version) // Gets first floor

	pkg, err = a.GetUpdatePackage(Instance{ID: "i2", IP: "10.0.0.2"}, NewInstanceApplication(setup.AppID, setup.Group.ID, "2500.0.0"))
	assert.NoError(t, err)
	assert.Equal(t, "3000.0.0", pkg.Version) // Gets target-floor directly
}

// TestFloorBlacklistConflict tests that packages cannot be both floor and blacklisted for the same channel
func TestFloorBlacklistConflict(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	setup := setupFloors(t, a, "test-blacklist", []string{"1.0.0"}, "2.0.0")

	t.Run("cannot_blacklist_floor", func(t *testing.T) {
		// Floor is already set up, try to blacklist it by updating the package
		floorPkg, err := a.GetPackage(setup.Floors[0].ID)
		assert.NoError(t, err)
		floorPkg.ChannelsBlacklist = append(floorPkg.ChannelsBlacklist, setup.Channel.ID)
		err = a.UpdatePackage(floorPkg)
		assert.Equal(t, ErrBlacklistingFloor, err)
	})

	t.Run("cannot_blacklist_channel_target", func(t *testing.T) {
		// Try to blacklist channel's current package
		targetPkg, err := a.GetPackage(setup.Target.ID)
		assert.NoError(t, err)
		targetPkg.ChannelsBlacklist = append(targetPkg.ChannelsBlacklist, setup.Channel.ID)
		err = a.UpdatePackage(targetPkg)
		assert.Equal(t, ErrBlacklistingChannel, err)
	})

	t.Run("cannot_mark_blacklisted_as_floor", func(t *testing.T) {
		// Create new package with blacklist
		pkg := quickPkg(t, a, setup.AppID, "3.0.0")
		pkg.ChannelsBlacklist = StringArray{setup.Channel.ID}
		err := a.UpdatePackage(pkg)
		assert.NoError(t, err)

		// Try to mark as floor - should fail because it's blacklisted
		err = a.AddChannelPackageFloor(setup.Channel.ID, pkg.ID, null.StringFrom("Should fail"))
		assert.Error(t, err)
	})
}

// TestSelectFloorsInRange covers the floor version range, ordering, and limit logic,
// including pre-release and build metadata precedence.
func TestSelectFloorsInRange(t *testing.T) {
	mk := func(vs ...string) []*Package {
		ps := make([]*Package, len(vs))
		for i, v := range vs {
			ps[i] = &Package{Version: v}
		}
		return ps
	}
	vers := func(ps []*Package) []string {
		out := make([]string, 0, len(ps))
		for _, p := range ps {
			out = append(out, p.Version)
		}
		return out
	}

	tests := []struct {
		name             string
		floors           []*Package
		instance, target string
		want             []string
	}{
		{"release floor not skipped over pre-release instance", mk("2000.0.0"), "2000.0.0-rc1", "3000.0.0", []string{"2000.0.0"}},
		{"same-core floors ordered, none dropped", mk("2000.0.0", "2000.0.0-beta", "2000.0.0-alpha"), "1000.0.0", "3000.0.0", []string{"2000.0.0-alpha", "2000.0.0-beta", "2000.0.0"}},
		{"lower pre-release below release instance excluded", mk("2000.0.0-rc1"), "2000.0.0", "3000.0.0", []string{}},
		{"build metadata treated as the release", mk("2000.0.0+build"), "1999.0.0", "3000.0.0", []string{"2000.0.0+build"}},
		{"shuffled input sorted, all in range", mk("3000.0.0", "1000.0.0", "2000.0.0"), "500.0.0", "4000.0.0", []string{"1000.0.0", "2000.0.0", "3000.0.0"}},
		{"target inclusive, above-target excluded", mk("2000.0.0", "4000.0.0"), "1000.0.0", "3000.0.0", []string{"2000.0.0"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := selectFloorsInRange(tt.floors, tt.instance, tt.target)
			require.NoError(t, err)
			assert.Equal(t, tt.want, vers(res))
		})
	}
}

// TestFloorPreReleaseNotSkipped checks that a release floor is still returned when the
// instance reports a pre-release of the same version.
func TestFloorPreReleaseNotSkipped(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	setup := setupFloors(t, a, "prerelease", []string{"2000.0.0"}, "3000.0.0")

	ch, err := a.GetChannel(setup.Channel.ID)
	require.NoError(t, err)
	floors, err := a.GetRequiredChannelFloors(ch, "2000.0.0-rc1")
	require.NoError(t, err)
	require.Len(t, floors, 1)
	assert.Equal(t, "2000.0.0", floors[0].Version)
}
