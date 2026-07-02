package syncer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
)

// pinChannel pre-creates a package at version and points the channel (and the walk
// cursor) at it, giving the walk a known starting point.
func pinChannel(t *testing.T, s *Syncer, a *api.API, tChannel *api.Channel, version string) channelDescriptor {
	t.Helper()
	pkg, err := a.AddPackage(&api.Package{
		Type: api.PkgTypeFlatcar, URL: "https://example.com/" + version,
		Version: version, Filename: null.StringFrom("flatcar-" + version + ".gz"),
		ApplicationID: flatcarAppID, Arch: tChannel.Arch,
	})
	require.NoError(t, err)
	tChannel.PackageID = null.StringFrom(pkg.ID)
	require.NoError(t, a.UpdateChannel(tChannel))

	desc := channelDescriptor{name: tChannel.Name, arch: tChannel.Arch}
	s.versions[desc] = version
	return desc
}

// TestSyncer_FloorAdvancesCursorNotChannel verifies that an intermediate floor is
// recorded and advances only the walk cursor - the channel must stay put (it may
// never point at an intermediate floor).
func TestSyncer_FloorAdvancesCursorNotChannel(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)
	desc := pinChannel(t, syncer, a, tChannel, "500.0.0")

	require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate("1000.0.0", true, false)))

	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	require.Len(t, floors, 1)
	assert.Equal(t, "1000.0.0", floors[0].Version)

	assert.Equal(t, "1000.0.0", syncer.versions[desc], "cursor advances to the floor")

	ch, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "500.0.0", ch.Package.Version, "channel must NOT advance to an intermediate floor")
}

// TestSyncer_TargetAdvancesChannel verifies the target advances the channel directly.
func TestSyncer_TargetAdvancesChannel(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)
	desc := pinChannel(t, syncer, a, tChannel, "500.0.0")

	require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate("2000.0.0", false, true)))

	ch, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "2000.0.0", ch.Package.Version, "target advances the channel")
	assert.Equal(t, "2000.0.0", syncer.versions[desc])
}

// TestSyncer_TargetThatIsAlsoFloor records the floor and advances the channel.
func TestSyncer_TargetThatIsAlsoFloor(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)
	desc := pinChannel(t, syncer, a, tChannel, "500.0.0")

	require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate("2000.0.0", true, true)))

	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	require.Len(t, floors, 1)
	assert.Equal(t, "2000.0.0", floors[0].Version)

	ch, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "2000.0.0", ch.Package.Version)
}

// TestSyncer_FloorWalkInOrder walks floors then the target across cycles: the channel
// stays put while walking floors, floors are recorded in ascending order, and the
// channel advances directly to the target at the end.
func TestSyncer_FloorWalkInOrder(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)
	desc := pinChannel(t, syncer, a, tChannel, "500.0.0")

	require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate("1000.0.0", true, false)))
	require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate("2000.0.0", true, false)))

	ch, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "500.0.0", ch.Package.Version, "channel stays put across floor steps")

	require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate("3000.0.0", false, true)))

	floors, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	require.Len(t, floors, 2)
	assert.Equal(t, "1000.0.0", floors[0].Version)
	assert.Equal(t, "2000.0.0", floors[1].Version)

	ch, err = a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "3000.0.0", ch.Package.Version, "channel advances directly to the target")
}

// TestSyncer_LongFloorWalk walks more floors than the old per-response cap (5) to
// guard against a re-introduced limit: every floor must be recorded, in ascending
// order, with the channel staying at the start until the target arrives.
func TestSyncer_LongFloorWalk(t *testing.T) {
	syncer, a, _, tChannel := setupSyncerTest(t)
	desc := pinChannel(t, syncer, a, tChannel, "500.0.0")

	floors := []string{"1000.0.0", "2000.0.0", "3000.0.0", "4000.0.0", "5000.0.0", "6000.0.0", "7000.0.0"}
	for _, v := range floors {
		require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate(v, true, false)))
		ch, err := a.GetChannel(tChannel.ID)
		require.NoError(t, err)
		assert.Equal(t, "500.0.0", ch.Package.Version, "channel must stay at start while walking floor %s", v)
	}

	// Finally the target advances the channel.
	require.NoError(t, syncer.processUpdate(desc, createSyncerUpdate("8000.0.0", false, true)))

	recorded, err := a.GetChannelFloorPackages(tChannel.ID)
	require.NoError(t, err)
	got := make([]string, len(recorded))
	for i, f := range recorded {
		got[i] = f.Version
	}
	assert.Equal(t, floors, got, "every floor recorded in ascending order, none dropped")

	ch, err := a.GetChannel(tChannel.ID)
	require.NoError(t, err)
	assert.Equal(t, "8000.0.0", ch.Package.Version, "channel advances to the target only at the end")
}
