package api

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestRegisterInstance(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "group2", ApplicationID: tApp2.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup3, _ := a.AddGroup(&Group{Name: "group3", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.New().String()

	_, err := a.RegisterInstance("", "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	assert.Error(t, err, "Using empty string as instance id.")

	_, err = a.RegisterInstance(instanceID, "", "invalidIP", "1.0.0", tApp.ID, tGroup.ID)
	assert.Error(t, err, "Using an invalid instance ip.")

	_, err = a.RegisterInstance(instanceID, "", "10.0.0.1", "1.0.0", "invalidAppID", tGroup.ID)
	assert.Error(t, err, "Using an invalid application id.")

	_, err = a.RegisterInstance(instanceID, "", "10.0.0.1", "1.0.0", tApp.ID, "invalidGroupID")
	assert.Error(t, err, "Using an invalid group id.")

	_, err = a.RegisterInstance(instanceID, "", "10.0.0.1", "", tApp.ID, "invalidGroupID")
	assert.Error(t, err, "Using an empty instance version.")

	_, err = a.RegisterInstance(instanceID, "", "10.0.0.1", "aaa1.0.0", tApp.ID, "invalidGroupID")
	assert.Equal(t, ErrInvalidSemver, err, "Using an invalid instance version.")

	_, err = a.RegisterInstance(instanceID, "", "10.0.0.1", "1.0.0", tApp.ID, tGroup2.ID)
	assert.Equal(t, ErrInvalidApplicationOrGroup, err, "The group provided doesn't belong to the application provided.")

	instance, err := a.RegisterInstance(instanceID, "myalias", "10.0.0.1", "1.0.0", "{"+tApp.ID+"}", "{"+tGroup.ID+"}")
	assert.NoError(t, err)
	assert.Equal(t, instanceID, instance.ID)
	assert.Equal(t, "myalias", instance.Alias)
	assert.Equal(t, "10.0.0.1", instance.IP)

	instance, err = a.RegisterInstance(instanceID, "mynewalias", "10.0.0.2", "1.0.2", tApp.ID, tGroup.ID)
	assert.NoError(t, err, "Registering an already registered instance with some updates, that's fine.")
	assert.Equal(t, "mynewalias", instance.Alias)
	assert.Equal(t, "10.0.0.2", instance.IP)
	assert.Equal(t, "1.0.2", instance.Application.Version)

	_, err = a.RegisterInstance(instanceID, "", "10.0.0.2", "1.0.2", tApp2.ID, tGroup.ID)
	assert.Error(t, err, "Application id cannot be updated.")

	instance, err = a.RegisterInstance(instanceID, "", "10.0.0.3", "1.0.3", tApp.ID, tGroup3.ID)
	assert.NoError(t, err, "Registering an already registered instance using a different group, that's fine.")
	assert.Equal(t, "10.0.0.3", instance.IP)
	assert.Equal(t, "1.0.3", instance.Application.Version)
	assert.Equal(t, null.StringFrom(tGroup3.ID), instance.Application.GroupID)
}

func TestGetInstance(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, err := a.GetInstance(uuid.New().String(), tApp.ID)
	assert.Error(t, err, "Using non existent instance id.")

	_, err = a.GetInstance("invalidInstanceID", tApp.ID)
	assert.Error(t, err, "Using invalid instance id.")

	_, err = a.GetInstance(tInstance.ID, "invalidApplicationID")
	assert.Error(t, err, "Using invalid application id.")

	instance, err := a.GetInstance(tInstance.ID, tApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", instance.IP)
	assert.Equal(t, tApp.ID, instance.Application.ApplicationID)
	assert.Equal(t, null.StringFrom(tGroup.ID), instance.Application.GroupID)
	assert.Equal(t, "1.0.0", instance.Application.Version)
}

func TestGetInstances(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.New().String(), "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	_, _ = a.RegisterInstance(uuid.New().String(), "", "10.0.0.2", "1.0.1", tApp.ID, tGroup.ID)
	_, _ = a.RegisterInstance(uuid.New().String(), "", "10.0.0.3", "1.0.2", tApp.ID, tGroup2.ID)

	result, err := a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 1, (int)(result.TotalInstances))

	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Instances))
	assert.Equal(t, 2, (int)(result.TotalInstances))

	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 1}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 2, (int)(result.TotalInstances))

	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup2.ID, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 1, (int)(result.TotalInstances))

	// Search for a non-existant Version should give no results.
	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "non-existant", Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, "1.0.0", result.Instances[0].Application.Version)

	// Search for a non-existant GroupID should give no results.
	nonExistentGroupID := "123e4567-e89b-12d3-a456-426614174000"
	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: nonExistentGroupID, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	// sortFilter 2 == last_check_for_updates
	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup2.ID, Page: 1, PerPage: 10, SortFilter: "2"}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 1, (int)(result.TotalInstances))

	// Search with sortFilter and non-existant GroupID should give no results.
	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: nonExistentGroupID, Page: 1, PerPage: 10, SortFilter: "2"}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	// Search with sortFilter for a non-existant Version should give no results.
	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "non-existant", Page: 1, PerPage: 10, SortFilter: "2"}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	_, _ = a.GetUpdatePackage(tInstance.ID, "", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	_ = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")

	result, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Status: InstanceStatusComplete, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))

	_, err = a.GetInstances(InstancesQueryParams{GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Application id must be provided.")

	_, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Group id must be provided.")

	_, err = a.GetInstances(InstancesQueryParams{Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Application id and group id are required and must be valid uuids.")

	_, err = a.GetInstances(InstancesQueryParams{ApplicationID: "invalidApplicationID", GroupID: "invalidGroupID", Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Application id and group id are required and must be valid uuids.")
}

func TestGetInstancesFiltered(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	instanceID1 := "{8d180b2a-0734-4406-af02-9a4f86bd1ee0}"
	instanceID2 := "{8d180b2a07344406af029a4f86bd1ee1}"
	instanceID3 := "8d180b2a-0734-4406-af02-9a4f86bd1ee2"
	instanceID4 := "8d180b2a07344406af029a4f86bd1ee3"
	for idx, id := range []string{instanceID1, instanceID2, instanceID3, instanceID4} {
		ip := fmt.Sprintf("10.0.0.%d", idx+1)
		_, _ = a.RegisterInstance(id, "", ip, "1.0.0", tApp.ID, tGroup.ID)
	}

	result, err := a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	expectedIDs := map[string]struct{}{
		instanceID2: {},
		instanceID3: {},
		instanceID4: {},
	}
	if assert.Equal(t, 3, len(result.Instances)) {
		for _, instance := range result.Instances {
			assert.Contains(t, expectedIDs, instance.ID)
			delete(expectedIDs, instance.ID)
		}
		assert.Empty(t, expectedIDs)
	}
}
