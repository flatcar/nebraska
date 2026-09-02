package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestRegisterInstance(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := as.AddApp(&types.Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := as.AddGroup(&types.Group{Name: "group2", ApplicationID: tApp2.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup3, _ := as.AddGroup(&types.Group{Name: "group3", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.New().String()

	_, err := rs.RegisterInstance(types.Instance{ID: "", IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	assert.Error(t, err, "Using empty string as instance id.")

	_, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "invalidIP"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	assert.Error(t, err, "Using an invalid instance ip.")

	_, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.1"}, runtime.NewInstanceApplication("invalidAppID", tGroup.ID, "1.0.0"))
	assert.Error(t, err, "Using an invalid application id.")

	_, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, "invalidGroupID", "1.0.0"))
	assert.Error(t, err, "Using an invalid group id.")

	_, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, "invalidGroupID", ""))
	assert.Error(t, err, "Using an empty instance version.")

	_, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, "invalidGroupID", "aaa1.0.0"))
	assert.Equal(t, types.ErrInvalidSemver, err, "Using an invalid instance version.")

	_, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup2.ID, "1.0.0"))
	assert.Equal(t, types.ErrInvalidApplicationOrGroup, err, "The group provided doesn't belong to the application provided.")

	instance, err := rs.RegisterInstance(types.Instance{ID: instanceID, Alias: "myalias", IP: "10.0.0.1", OEM: "azure", AlephVersion: "2.9.1.1-r1"}, runtime.NewInstanceApplication("{"+tApp.ID+"}", "{"+tGroup.ID+"}", "1.0.0"))
	assert.NoError(t, err)
	assert.Equal(t, instanceID, instance.ID)
	assert.Equal(t, "myalias", instance.Alias)
	assert.Equal(t, "10.0.0.1", instance.IP)
	assert.Equal(t, "azure", instance.OEM)
	assert.Equal(t, "2.9.1.1-r1", instance.AlephVersion)

	instance, err = rs.RegisterInstance(types.Instance{ID: instanceID, Alias: "mynewalias", IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.2"))
	assert.NoError(t, err, "Registering an already registered instance with some updates, that's fine.")
	assert.Equal(t, "mynewalias", instance.Alias)
	assert.Equal(t, "10.0.0.2", instance.IP)
	assert.Equal(t, "1.0.2", instance.Application.Version)
	assert.Equal(t, "azure", instance.OEM, "OEM should be preserved when not provided")
	assert.Equal(t, "2.9.1.1-r1", instance.AlephVersion, "AlephVersion should be preserved when not provided")

	_, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp2.ID, tGroup.ID, "1.0.2"))
	assert.Error(t, err, "Application id cannot be updated.")

	instance, err = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.3", OEM: "gcp", AlephVersion: "3.0.0"}, runtime.NewInstanceApplication(tApp.ID, tGroup3.ID, "1.0.3"))
	assert.NoError(t, err, "Registering an already registered instance using a different group, that's fine.")
	assert.Equal(t, "10.0.0.3", instance.IP)
	assert.Equal(t, "1.0.3", instance.Application.Version)
	assert.Equal(t, null.StringFrom(tGroup3.ID), instance.Application.GroupID)
	assert.Equal(t, "gcp", instance.OEM, "OEM should be updated when provided")
	assert.Equal(t, "3.0.0", instance.AlephVersion, "AlephVersion should be updated when provided")
}

func TestGetInstance(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

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
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := as.AddGroup(&types.Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.1"))
	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.3"}, runtime.NewInstanceApplication(tApp.ID, tGroup2.ID, "1.0.2"))

	result, err := a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 1, (int)(result.TotalInstances))

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Instances))
	assert.Equal(t, 2, (int)(result.TotalInstances))

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 1}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 2, (int)(result.TotalInstances))

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup2.ID, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 1, (int)(result.TotalInstances))

	// Search for a non-existant Version should give no results.
	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "non-existant", Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, "1.0.0", result.Instances[0].Application.Version)

	// Search for a non-existant GroupID should give no results.
	nonExistentGroupID := "123e4567-e89b-12d3-a456-426614174000"
	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: nonExistentGroupID, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	// sortFilter 2 == last_check_for_updates
	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup2.ID, Page: 1, PerPage: 10, SortFilter: "2"}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, 1, (int)(result.TotalInstances))

	// Search with sortFilter and non-existant GroupID should give no results.
	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: nonExistentGroupID, Page: 1, PerPage: 10, SortFilter: "2"}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	// Search with sortFilter for a non-existant Version should give no results.
	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "non-existant", Page: 1, PerPage: 10, SortFilter: "2"}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Instances))

	_, _ = rs.GetUpdatePackage(types.Instance{ID: tInstance.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	_ = rs.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, types.EventUpdateComplete, types.ResultSuccessReboot, "", "")

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Status: types.InstanceStatusComplete, Page: 1, PerPage: 10}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))

	_, err = a.GetInstances(types.InstancesQueryParams{GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Application id must be provided.")

	_, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Group id must be provided.")

	_, err = a.GetInstances(types.InstancesQueryParams{Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Application id and group id are required and must be valid uuids.")

	_, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: "invalidApplicationID", GroupID: "invalidGroupID", Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
	assert.Error(t, err, "Application id and group id are required and must be valid uuids.")
}

func TestGetInstancesSearch(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.1"))
	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.3"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.2"))

	instanceAlias := "instance_alias"
	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), Alias: instanceAlias, IP: "10.0.0.4"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.4"))

	result, err := a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 10, SearchFilter: "All", SearchValue: tInstance.ID}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, result.Instances[0].ID, tInstance.ID)

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 10, SearchFilter: "id", SearchValue: tInstance.ID}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
	assert.Equal(t, result.Instances[0].ID, tInstance.ID)

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 10, SearchFilter: "ip", SearchValue: "10.0"}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(result.Instances))

	result, err = a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 10, SearchFilter: "alias", SearchValue: instanceAlias}, testDuration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Instances))
}

func TestGetInstancesFiltered(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	instanceID1 := "{8d180b2a-0734-4406-af02-9a4f86bd1ee0}"
	instanceID2 := "{8d180b2a07344406af029a4f86bd1ee1}"
	instanceID3 := "8d180b2a-0734-4406-af02-9a4f86bd1ee2"
	instanceID4 := "8d180b2a07344406af029a4f86bd1ee3"
	for idx, id := range []string{instanceID1, instanceID2, instanceID3, instanceID4} {
		ip := fmt.Sprintf("10.0.0.%d", idx+1)
		_, _ = rs.RegisterInstance(types.Instance{ID: id, IP: ip}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	}

	result, err := a.GetInstances(types.InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10}, testDuration)
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

func TestUpdateInstanceStats(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	instances, err := a.GetInstanceStats()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(instances))

	// First test case: Create tInstance1, tInstance2, and tInstance3; check tInstance1 twice; switch tInstance2 version
	start := time.Now().UTC()

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: types.ArchAMD64})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID), Arch: types.ArchAMD64})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance1, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	tInstance2, _ := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.3"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.1"))

	_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance1.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	assert.NoError(t, err)

	_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance2.ID, IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.1"))
	assert.NoError(t, err)

	ts := time.Now().UTC()
	elapsed := ts.Sub(start)

	err = rs.UpdateInstanceStats(&ts, &elapsed)
	assert.NoError(t, err)

	instances, err = a.GetInstanceStats()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(instances))

	instanceStats, err := a.GetInstanceStatsByTimestamp(ts)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(instanceStats))
	assert.Equal(t, "1.0.0", instanceStats[0].Version)
	assert.Equal(t, 1, instanceStats[0].Instances)
	assert.Equal(t, "1.0.1", instanceStats[1].Version)
	assert.Equal(t, 2, instanceStats[1].Instances)

	// Next test case: Switch tInstance1 and tInstance2 versions to workaround the 5-minutes-rate-limiting of the check-in time and add new instance
	ts2 := time.Now().UTC()

	_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance1.ID, IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.3"))
	assert.NoError(t, err)

	_, err = rs.GetUpdatePackage(types.Instance{ID: tInstance2.ID, IP: "10.0.0.2"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.4"))
	assert.NoError(t, err)

	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.4"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.5"))

	ts3 := time.Now().UTC()
	elapsed = ts3.Sub(ts2)

	err = rs.UpdateInstanceStats(&ts3, &elapsed)
	assert.NoError(t, err)

	instances, err = a.GetInstanceStats()
	assert.NoError(t, err)
	assert.Equal(t, 5, len(instances))

	instanceStats, err = a.GetInstanceStatsByTimestamp(ts3)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(instanceStats))
	assert.Equal(t, "1.0.3", instanceStats[0].Version)
	assert.Equal(t, 1, instanceStats[0].Instances)
	assert.Equal(t, "1.0.4", instanceStats[1].Version)
	assert.Equal(t, 1, instanceStats[1].Instances)
	assert.Equal(t, "1.0.5", instanceStats[2].Version)
	assert.Equal(t, 1, instanceStats[2].Instances)
}

func TestUpdateInstanceStatsNoArch(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	_, _ = rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, runtime.NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))

	ts := time.Now().UTC()
	// Use large duration to have some test coverage for durationToInterval
	elapsed := 3*time.Hour + 45*time.Minute + 30*time.Second + 1000*time.Microsecond

	err := rs.UpdateInstanceStats(&ts, &elapsed)
	assert.NoError(t, err)

	instanceStats, err := a.GetInstanceStatsByTimestamp(ts)
	assert.NoError(t, err)
	assert.Equal(t, "", instanceStats[0].Arch)
	assert.Equal(t, "1.0.0", instanceStats[0].Version)
	assert.Equal(t, 1, instanceStats[0].Instances)
}
