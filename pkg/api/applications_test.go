package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgutz/dat.v1"
)

func TestAddApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam := mustAddTeam(t, a, &Team{Name: "test_team"})

	newApp, err := a.AddApp(&Application{Name: "app1", TeamID: tTeam.ID})
	assert.NoError(t, err)
	if assert.NotNil(t, newApp) {
		newAppX, err := a.GetApp(newApp.ID)
		assert.NoError(t, err)
		if assert.NotNil(t, newAppX) {
			assert.Equal(t, "app1", newAppX.Name)
		}
	}

	var app *Application
	app, err = a.AddApp(&Application{Name: "app1", TeamID: tTeam.ID})
	assert.Nil(t, app)
	assert.Error(t, err, "App name must be unique per team.")

	app, err = a.AddApp(&Application{TeamID: tTeam.ID})
	assert.Nil(t, app)
	assert.Error(t, err, "App name is required.")

	app, err = a.AddApp(&Application{Name: "app2"})
	assert.Nil(t, app)
	assert.Error(t, err, "Team id is required.")

	app, err = a.AddApp(&Application{Name: "app2", TeamID: uuid.New().String()})
	assert.Nil(t, app)
	assert.Error(t, err, "Team id used must exist.")

	app, err = a.AddApp(&Application{Name: "app2", TeamID: "invalidTeamID"})
	assert.Nil(t, app)
	assert.Error(t, err, "Team id must be a valid uuid.")
}

func TestAddAppCloning(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam := mustAddTeam(t, a, &Team{Name: "test_team"})
	tApp := mustAddApp(t, a, &Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg := mustAddPackage(t, a, &Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel := mustAddChannel(t, a, &Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	_ = mustAddGroup(t, a, &Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: false, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	_ = mustAddGroup(t, a, &Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	clonedApp, err := a.AddAppCloning(&Application{Name: "app1", TeamID: tTeam.ID}, tApp.ID)
	assert.NoError(t, err)
	if assert.NotNil(t, clonedApp) {
		sourceApp, err := a.GetApp(tApp.ID)
		require.NoError(t, err)
		require.NotNil(t, sourceApp)
		clonedAppX, err := a.GetApp(clonedApp.ID)
		require.NoError(t, err)
		require.NotNil(t, clonedAppX)
		assert.NotEqual(t, sourceApp.ID, clonedAppX.ID)
		assert.Len(t, clonedAppX.Groups, len(sourceApp.Groups))
		assert.Len(t, clonedAppX.Channels, len(sourceApp.Channels))
		sourceGroups := groupsToMap(t, sourceApp.Groups)
		clonedGroups := groupsToMap(t, clonedAppX.Groups)
		sourceChannels := channelsToMap(t, sourceApp.Channels)
		clonedChannels := channelsToMap(t, clonedAppX.Channels)
		channelIDMapping := make(map[string]string)
		for name, clonedChannel := range clonedChannels {
			if !assert.Contains(t, sourceChannels, name) {
				continue
			}
			sourceChannel := sourceChannels[name]
			assert.NotEqual(t, sourceChannel.ID, clonedChannel.ID)
			channelIDMapping[sourceChannel.ID] = clonedChannel.ID
			assert.Equal(t, sourceChannel.Name, clonedChannel.Name)
			assert.Equal(t, sourceChannel.Color, clonedChannel.Color)
			assert.Equal(t, sourceApp.ID, sourceChannel.ApplicationID)
			assert.Equal(t, clonedAppX.ID, clonedChannel.ApplicationID)
			assert.False(t, clonedChannel.PackageID.Valid)
		}
		for name, clonedGroup := range clonedGroups {
			if !assert.Contains(t, sourceGroups, name) {
				continue
			}
			sourceGroup := sourceGroups[name]
			assert.NotEqual(t, sourceGroup.ID, clonedGroup.ID)
			assert.Equal(t, sourceGroup.Name, clonedGroup.Name)
			assert.Equal(t, sourceGroup.Description, clonedGroup.Description)
			assert.Equal(t, sourceApp.ID, sourceGroup.ApplicationID)
			assert.Equal(t, clonedAppX.ID, clonedGroup.ApplicationID)
			if assert.Equal(t, sourceGroup.ChannelID.Valid, clonedGroup.ChannelID.Valid) && sourceGroup.ChannelID.Valid {
				sourceChannelID := sourceGroup.ChannelID.String
				if assert.Contains(t, channelIDMapping, sourceChannelID) {
					clonedChannelID := channelIDMapping[sourceChannelID]
					assert.Equal(t, clonedChannelID, clonedGroup.ChannelID.String)
				}
			}
			assert.True(t, clonedGroup.PolicyUpdatesEnabled)
			assert.Equal(t, sourceGroup.PolicySafeMode, clonedGroup.PolicySafeMode)
			assert.Equal(t, sourceGroup.PolicyOfficeHours, clonedGroup.PolicyOfficeHours)
			assert.Equal(t, sourceGroup.PolicyTimezone, clonedGroup.PolicyTimezone)
			assert.Equal(t, sourceGroup.PolicyPeriodInterval, clonedGroup.PolicyPeriodInterval)
			assert.Equal(t, sourceGroup.PolicyMaxUpdatesPerPeriod, clonedGroup.PolicyMaxUpdatesPerPeriod)
			assert.Equal(t, sourceGroup.PolicyUpdateTimeout, clonedGroup.PolicyUpdateTimeout)
		}
	}

	clonedApp2, err := a.AddAppCloning(&Application{Name: "app2", TeamID: tTeam.ID}, "")
	assert.NoError(t, err, "Using an empty source app id when cloning has the same effect as not cloning.")
	if assert.NotNil(t, clonedApp2) {
		assert.Empty(t, clonedApp2.Groups)
		assert.Empty(t, clonedApp2.Channels)
	}
}

func groupsToMap(t *testing.T, groups []*Group) map[string]*Group {
	t.Helper()
	m := make(map[string]*Group, len(groups))
	for _, group := range groups {
		if assert.NotContains(t, m, group.Name) {
			m[group.Name] = group
		}
	}
	return m
}

func channelsToMap(t *testing.T, channels []*Channel) map[string]*Channel {
	t.Helper()
	m := make(map[string]*Channel, len(channels))
	for _, channel := range channels {
		if assert.NotContains(t, m, channel.Name) {
			m[channel.Name] = channel
		}
	}
	return m
}

func TestUpdateApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam := mustAddTeam(t, a, &Team{Name: "test_team"})
	tApp := mustAddApp(t, a, &Application{Name: "test_app", Description: "description", TeamID: tTeam.ID})

	err := a.UpdateApp(&Application{ID: tApp.ID, Name: "test_app_updated"})
	assert.NoError(t, err)

	app := mustGetApp(t, a, tApp.ID)
	assert.Equal(t, "test_app_updated", app.Name)
	assert.Equal(t, "", app.Description, "Description set to empty string in last update as it wasn't provided")

	err = a.UpdateApp(&Application{ID: tApp.ID, Name: "test_app", Description: "description_updated"})
	assert.NoError(t, err)

	app = mustGetApp(t, a, tApp.ID)
	assert.Equal(t, "test_app", app.Name)
	assert.Equal(t, "description_updated", app.Description)

	err = a.UpdateApp(&Application{Name: "test_app_updated_again"})
	assert.Error(t, err, "App id is required.")

	err = a.UpdateApp(&Application{ID: "invalidAppID", Name: "test_app_updated_again"})
	assert.Error(t, err, "App id must be a valid uuid.")
}

func TestDeleteApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam := mustAddTeam(t, a, &Team{Name: "test_team"})
	tApp := mustAddApp(t, a, &Application{Name: "test_app", TeamID: tTeam.ID})

	err := a.DeleteApp(tApp.ID)
	assert.NoError(t, err)

	app, err := a.GetApp(tApp.ID)
	assert.Error(t, err, "Trying to get deleted app.")
	assert.Nil(t, app)
}

func TestGetApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam := mustAddTeam(t, a, &Team{Name: "test_team"})
	tApp := mustAddApp(t, a, &Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel := mustAddChannel(t, a, &Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID})

	app, err := a.GetApp(tApp.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, tApp.Name, app.Name)
		assert.Equal(t, tChannel.Name, app.Channels[0].Name)
	}

	app, err = a.GetApp(uuid.New().String())
	assert.Error(t, err, "Trying to get non existent app.")
	assert.Nil(t, app)
}

func TestGetApps(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam := mustAddTeam(t, a, &Team{Name: "test_team"})
	tApp1 := mustAddApp(t, a, &Application{Name: "test_app1", TeamID: tTeam.ID})
	tApp2 := mustAddApp(t, a, &Application{Name: "test_app2", TeamID: tTeam.ID})
	tChannel := mustAddChannel(t, a, &Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp1.ID})

	apps, err := a.GetApps(tTeam.ID, 0, 0)
	assert.NoError(t, err)
	if assert.NotNil(t, apps) {
		assert.Equal(t, 2, len(apps))
		assert.Equal(t, tApp1.Name, apps[1].Name)
		assert.Equal(t, tApp2.Name, apps[0].Name)
		if assert.Len(t, apps[1].Channels, 1) {
			assert.Equal(t, tChannel.Name, apps[1].Channels[0].Name)
		}
	}

	apps, err = a.GetApps(uuid.New().String(), 0, 0)
	assert.Error(t, err, "Trying to get apps of inexisting team.")
	assert.Nil(t, apps)
}
