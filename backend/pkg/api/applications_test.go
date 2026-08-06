package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestAddApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})

	newApp, err := as.AddApp(&Application{Name: "app1", TeamID: tTeam.ID})
	assert.NoError(t, err)

	newAppX, err := a.GetApp(newApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "app1", newAppX.Name)

	_, err = as.AddApp(&Application{Name: "app1", TeamID: tTeam.ID})
	assert.Error(t, err, "App name must be unique per team.")

	_, err = as.AddApp(&Application{TeamID: tTeam.ID})
	assert.Error(t, err, "App name is required.")

	_, err = as.AddApp(&Application{Name: "app2"})
	assert.Error(t, err, "Team id is required.")

	_, err = as.AddApp(&Application{Name: "app2", TeamID: uuid.New().String()})
	assert.Error(t, err, "Team id used must exist.")

	_, err = as.AddApp(&Application{Name: "app2", TeamID: "invalidTeamID"})
	assert.Error(t, err, "Team id must be a valid uuid.")
}

func TestAddAppCloning(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	_, _ = as.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "beta"})
	_, _ = as.AddGroup(&Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	clonedApp, err := as.AddAppCloning(&Application{Name: "app1", TeamID: tTeam.ID}, tApp.ID)
	assert.NoError(t, err)

	sourceApp, _ := a.GetApp(tApp.ID)
	clonedAppX, _ := a.GetApp(clonedApp.ID)
	assert.Equal(t, len(sourceApp.Groups), len(clonedAppX.Groups))
	assert.Equal(t, len(sourceApp.Channels), len(clonedAppX.Channels))

	// Verify cloned channel attributes
	assert.Equal(t, "test_channel", clonedAppX.Channels[0].Name)
	assert.Equal(t, "blue", clonedAppX.Channels[0].Color)
	assert.Equal(t, clonedApp.ID, clonedAppX.Channels[0].ApplicationID)
	assert.False(t, clonedAppX.Channels[0].PackageID.Valid, "Cloned channel should not have package attached")

	// Verify cloned group attributes, custom track, and mapped channel ID
	var group1, group2 *Group
	for _, g := range clonedAppX.Groups {
		if g.Name == "group1" {
			group1 = g
		} else if g.Name == "group2" {
			group2 = g
		}
	}
	assert.NotNil(t, group1)
	assert.Equal(t, clonedApp.ID, group1.ApplicationID)
	assert.True(t, group1.ChannelID.Valid)
	assert.Equal(t, clonedAppX.Channels[0].ID, group1.ChannelID.String)
	assert.NotEqual(t, tChannel.ID, group1.ChannelID.String)
	assert.Equal(t, "beta", group1.Track, "Custom track should be preserved")

	assert.NotNil(t, group2)
	assert.Equal(t, clonedApp.ID, group2.ApplicationID)
	assert.False(t, group2.ChannelID.Valid, "Group without channel should remain without channel")
	assert.Equal(t, group2.ID, group2.Track, "Default track should match new group ID")

	_, err = as.AddAppCloning(&Application{Name: "app2", TeamID: tTeam.ID}, "")
	assert.NoError(t, err, "Using an empty source app id when cloning has the same effect as not cloning.")

	// Test cloning with invalid/non-existent source app id returns error and does not create app
	_, err = as.AddAppCloning(&Application{Name: "app_invalid_source", TeamID: tTeam.ID}, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)

	apps, listErr := a.GetApps(tTeam.ID, 0, 0)
	assert.NoError(t, listErr)
	found := false
	for _, existing := range apps {
		if existing.Name == "app_invalid_source" {
			found = true
			break
		}
	}
	assert.False(t, found, "Application should not be created when source app is invalid")

	// Test rollback when cloning fails midway: create source app with a broken group
	brokenSourceApp, err := as.AddApp(&Application{Name: "broken_source_app", TeamID: tTeam.ID})
	assert.NoError(t, err)
	_, err = as.AddChannel(&Channel{Name: "broken_channel", Color: "green", ApplicationID: brokenSourceApp.ID})
	assert.NoError(t, err)
	// Directly insert a group with invalid timezone and policy_office_hours=true to trigger error during AddGroup cloning
	_, err = a.db.Exec("insert into groups (id, name, description, application_id, policy_office_hours, policy_timezone, policy_period_interval, policy_max_updates_per_period, policy_update_timeout) values ($1, $2, 'test_desc', $3, true, 'InvalidTimezone', '15 minutes', 2, '60 minutes')",
		uuid.New().String(), "invalid_tz_group", brokenSourceApp.ID)
	assert.NoError(t, err)

	_, err = as.AddAppCloning(&Application{Name: "app_should_rollback", TeamID: tTeam.ID}, brokenSourceApp.ID)
	assert.Error(t, err, "AddAppCloning should fail when group cloning fails")

	apps, listErr = a.GetApps(tTeam.ID, 0, 0)
	assert.NoError(t, listErr)
	foundRollback := false
	for _, existing := range apps {
		if existing.Name == "app_should_rollback" {
			foundRollback = true
			break
		}
	}
	assert.False(t, foundRollback, "Partially cloned application should be deleted/rolled back on failure")

	_, err = as.AddApp(&Application{Name: "productIDApp1", TeamID: tTeam.ID, ProductID: null.StringFrom("io.invalid. Name")})
	assert.Error(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp1", TeamID: tTeam.ID, ProductID: null.StringFrom("1io.invalid.Name")})
	assert.Error(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp1", TeamID: tTeam.ID, ProductID: null.StringFrom("")})
	assert.Error(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp1", TeamID: tTeam.ID, ProductID: null.StringFrom("io.invalid-.Name")})
	assert.Error(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp1", TeamID: tTeam.ID, ProductID: null.StringFrom("io.invalid_.Name")})
	assert.Error(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp1", TeamID: tTeam.ID, ProductID: null.StringFrom("io.valid.Name")})
	assert.NoError(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp2", TeamID: tTeam.ID, ProductID: null.StringFrom("io.valid.New-Name")})
	assert.NoError(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp3", TeamID: tTeam.ID, ProductID: null.StringFrom("io2.valid12.New-Name")})
	assert.NoError(t, err)

	_, err = as.AddApp(&Application{Name: "productIDApp4", TeamID: tTeam.ID, ProductID: null.StringFrom("io.invalid.New_Name")})
	assert.Error(t, err)

	tooLongName := `io.` +
		`loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong` +
		`loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong` +
		`loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong` +
		`loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong.example.com`

	_, err = as.AddApp(&Application{Name: "productIDApp5", TeamID: tTeam.ID, ProductID: null.StringFrom(tooLongName)})
	assert.Error(t, err)

	_, err = as.AddApp(
		&Application{
			Name:      "productIDApp4",
			TeamID:    tTeam.ID,
			ProductID: null.StringFrom("io.VALID.Name"),
		},
	)
	assert.Error(t, err, "duplicate name because it is case insensitive")
}

func TestUpdateApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", Description: "description", TeamID: tTeam.ID})

	err := as.UpdateApp(&Application{ID: tApp.ID, Name: "test_app_updated"})
	assert.NoError(t, err)

	app, _ := a.GetApp(tApp.ID)
	assert.Equal(t, "test_app_updated", app.Name)
	assert.Equal(t, "", app.Description, "Description set to empty string in last update as it wasn't provided")

	err = as.UpdateApp(&Application{ID: tApp.ID, Name: "test_app", Description: "description_updated"})
	assert.NoError(t, err)

	app, _ = a.GetApp(tApp.ID)
	assert.Equal(t, "test_app", app.Name)
	assert.Equal(t, "description_updated", app.Description)

	err = as.UpdateApp(&Application{Name: "test_app_updated_again"})
	assert.Error(t, err, "App id is required.")

	err = as.UpdateApp(&Application{ID: "invalidAppID", Name: "test_app_updated_again"})
	assert.Error(t, err, "App id must be a valid uuid.")
}

func TestDeleteApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})

	err := as.DeleteApp(tApp.ID)
	assert.NoError(t, err)

	_, err = a.GetApp(tApp.ID)
	assert.Error(t, err, "Trying to get deleted app.")
}

func TestGetApp(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, err := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	assert.NoError(t, err)
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID})
	tGroup, _ := as.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	_, _ = a.RegisterInstance(Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	app, err := a.GetApp(tApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, tApp.Name, app.Name)
	assert.False(t, tApp.ProductID.Valid)
	assert.Equal(t, tChannel.Name, app.Channels[0].Name)
	assert.Equal(t, 1, app.Instances.Count)

	_, err = a.GetApp(uuid.New().String())
	assert.Error(t, err, "Trying to get non existent app.")

	tApp1, err := as.AddApp(&Application{Name: "test_app1", ProductID: null.StringFrom("io.flatcar.MyNewApp"), TeamID: tTeam.ID})
	assert.NoError(t, err)
	assert.NotEqual(t, null.StringFrom(""), tApp1.ProductID)

	app, err = a.GetApp(tApp1.ID)
	assert.NoError(t, err)
	assert.Equal(t, tApp1.Name, app.Name)

	appID, err := a.GetAppID(*tApp1.ProductID.Ptr())
	assert.NoError(t, err)

	app, err = a.GetApp(appID)
	assert.NoError(t, err)
	assert.Equal(t, tApp1.ProductID, app.ProductID)

	// App with same product_id
	_, err = as.AddApp(&Application{Name: "test_app2", ProductID: null.StringFrom("io.flatcar.MyNewApp"), TeamID: tTeam.ID})
	assert.Error(t, err)

	// App with a default product_id, to test the constraint is not limiting too much
	_, err = as.AddApp(&Application{Name: "test_app3", TeamID: tTeam.ID})
	assert.NoError(t, err)
}

func TestGetApps(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp1, _ := as.AddApp(&Application{Name: "test_app1", TeamID: tTeam.ID})
	tApp2, _ := as.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp1.ID})

	apps, err := a.GetApps(tTeam.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(apps))
	assert.Equal(t, tApp1.Name, apps[1].Name)
	assert.Equal(t, tApp2.Name, apps[0].Name)
	assert.Equal(t, tChannel.Name, apps[1].Channels[0].Name)

	_, err = a.GetApps(uuid.New().String(), 0, 0)
	assert.NoError(t, err, "should not have any error for non existing appID")
}

func TestGetAppIDs(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp1, _ := as.AddApp(&Application{Name: "test_app1", TeamID: tTeam.ID})
	tApp2, _ := as.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID, ProductID: null.StringFrom("io.flatcar.MyApp2")})

	apps, err := a.GetApps(tTeam.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(apps))
	assert.Equal(t, tApp1.Name, apps[1].Name)
	assert.Equal(t, tApp2.Name, apps[0].Name)

	app1ID, err := a.GetAppID(tApp1.ID)
	assert.NoError(t, err)
	assert.Equal(t, tApp1.ID, app1ID)

	app2ID, err := a.GetAppID(tApp2.ID)
	assert.NoError(t, err)
	assert.Equal(t, tApp2.ID, app2ID)

	_, err = a.GetAppID("io.flatcar.InvalidApp")
	assert.Error(t, err)

	_, err = a.GetAppID("")
	assert.Error(t, err)

	_, err = a.GetAppID("{")
	assert.Error(t, err)

	app2ID, err = a.GetAppID("io.flatcar.MyApp2")
	assert.NoError(t, err)
	assert.Equal(t, tApp2.ID, app2ID)

	tApp3, err := as.AddApp(&Application{Name: "test_app3", TeamID: tTeam.ID, ProductID: null.StringFrom("io.flatcar.MyApp3")})
	assert.NoError(t, err)

	app3ID, err := a.GetAppID("io.flatcar.MyApp3")
	assert.NoError(t, err)
	assert.Equal(t, tApp3.ID, app3ID)

	tApp2.ProductID = null.StringFrom("io.flatcar.App")
	err = as.UpdateApp(tApp2)
	assert.NoError(t, err)

	_, err = a.GetAppID("io.flatcar.MyApp2")
	assert.Error(t, err)
	assert.Equal(t, tApp2.ID, app2ID)

	app2ID, err = a.GetAppID("io.flatcar.App")
	assert.NoError(t, err)
	assert.Equal(t, tApp2.ID, app2ID)

	wrappedInBrackets := "{io.flatcar.App}"
	app2ID, err = a.GetAppID(wrappedInBrackets)
	assert.NoError(t, err)
	assert.Equal(t, tApp2.ID, app2ID)

	caseInsensitive := "io.Flatcar.app"
	app2ID, err = a.GetAppID(caseInsensitive)
	assert.NoError(t, err)
	assert.Equal(t, tApp2.ID, app2ID)
}

func TestGetAppsFiltered(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)

	tTeam, _ := as.AddTeam(&Team{Name: "test_team"})
	tApp, _ := as.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	realInstanceID := uuid.New().String()
	fakeInstanceID := "{" + uuid.New().String() + "}"
	_, _ = a.RegisterInstance(Instance{ID: realInstanceID, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	_, _ = a.RegisterInstance(Instance{ID: fakeInstanceID, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	// should ignore fake instance in Instances count
	apps, err := a.GetApps(tTeam.ID, 1, 10)
	assert.NoError(t, err)
	if assert.Len(t, apps, 1) {
		assert.Equal(t, 1, apps[0].Instances.Count)
	}
}
