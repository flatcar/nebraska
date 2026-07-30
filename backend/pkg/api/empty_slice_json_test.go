package api

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func assertMarshalsToEmptyArray(t *testing.T, name string, v interface{}) {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Equal(t, "[]", string(b), "%s must serialize to [] and not null", name)
}

func TestEmptyResultsSerializeToEmptyArray(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, err := as.AddTeam(&Team{Name: "test_team_empty_slice"})
	require.NoError(t, err)
	tApp, err := as.AddApp(&Application{Name: "test_app_empty_slice", TeamID: tTeam.ID})
	require.NoError(t, err)
	tChannel, err := as.AddChannel(&Channel{Name: "test_channel_empty_slice", Color: "blue", ApplicationID: tApp.ID, Arch: ArchAMD64})
	require.NoError(t, err)
	tGroup, err := as.AddGroup(&Group{
		Name:                      "test_group_empty_slice",
		ApplicationID:             tApp.ID,
		ChannelID:                 null.StringFrom(tChannel.ID),
		PolicyUpdatesEnabled:      true,
		PolicySafeMode:            true,
		PolicyPeriodInterval:      "15 minutes",
		PolicyMaxUpdatesPerPeriod: 2,
		PolicyUpdateTimeout:       "60 minutes",
	})
	require.NoError(t, err)
	tPkg, err := as.AddPackage(&Package{
		Type:          PkgTypeOther,
		URL:           "http://sample.url/pkg",
		Version:       "1.0.0",
		ApplicationID: tApp.ID,
		Arch:          ArchAMD64,
	})
	require.NoError(t, err)

	t.Run("group version breakdown", func(t *testing.T) {
		got, err := a.GetGroupVersionBreakdown(tGroup.ID)
		require.NoError(t, err)
		assertMarshalsToEmptyArray(t, "version breakdown", got)
	})

	t.Run("instance status history", func(t *testing.T) {
		got, err := a.GetInstanceStatusHistory(uuid.New().String(), tApp.ID, tGroup.ID, 20)
		require.NoError(t, err)
		assertMarshalsToEmptyArray(t, "instance status history", got)
	})

	t.Run("activity", func(t *testing.T) {
		got, err := a.GetActivity(uuid.New().String(), ActivityQueryParams{})
		require.NoError(t, err)
		assertMarshalsToEmptyArray(t, "activity", got)
	})

	t.Run("package floor channels", func(t *testing.T) {
		got, err := a.GetPackageFloorChannels(tPkg.ID)
		require.NoError(t, err)
		assertMarshalsToEmptyArray(t, "package floor channels", got)
	})

	t.Run("package extra files", func(t *testing.T) {
		got, err := a.GetPackage(tPkg.ID)
		require.NoError(t, err)
		assertMarshalsToEmptyArray(t, "package extra files", got.ExtraFiles)
	})
}
