package api_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

func TestListInstances(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app which has instance
		appWithInstance := getAppWithInstance(t, db)

		// fetch instances from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances?status=0&sort=2&sortOrder=0&page=1&perpage=10&duration=30d", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appWithInstance.ID, appWithInstance.Groups[0].ID)
		method := "GET"

		// response
		var instances api.InstancesWithTotal

		httpDo(t, url, method, nil, http.StatusOK, "json", &instances)

		count, err := db.GetInstancesCount(api.InstancesQueryParams{
			ApplicationID: appWithInstance.ID,
			GroupID:       appWithInstance.Groups[0].ID,
		}, "30d")

		require.NoError(t, err)
		assert.Equal(t, len(instances.Instances), int(count))

		instancesDB, err := db.GetInstances(api.InstancesQueryParams{
			ApplicationID: appWithInstance.ID,
			GroupID:       appWithInstance.Groups[0].ID,
			Status:        0,
			SortOrder:     "0",
			Page:          0,
			PerPage:       10,
		}, "30d")
		require.NoError(t, err)
		require.Equal(t, instancesDB.Instances[0].ID, instances.Instances[0].ID)
	})

	t.Run("returns oem and aleph_version", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app which has instance
		appWithInstance := getAppWithInstance(t, db)
		group := appWithInstance.Groups[0]

		// an instance that reports OEM attributes returns them in the list
		instanceID := uuid.New().String()
		_, err := db.RegisterInstance(api.Instance{ID: instanceID, IP: "10.99.88.77", OEM: "azure", AlephVersion: "2.9.1.1-r1"}, api.NewInstanceApplication(appWithInstance.ID, group.ID, "1.0.0"))
		require.NoError(t, err)

		instance := fetchInstanceByID(t, appWithInstance.ID, group.ID, instanceID)
		require.NotNil(t, instance.Oem)
		require.NotNil(t, instance.AlephVersion)
		assert.Equal(t, "azure", *instance.Oem)
		assert.Equal(t, "2.9.1.1-r1", *instance.AlephVersion)

		// an instance that reports no OEM attributes omits both fields
		emptyOEMInstanceID := uuid.New().String()
		_, err = db.RegisterInstance(api.Instance{ID: emptyOEMInstanceID, IP: "10.99.88.78"}, api.NewInstanceApplication(appWithInstance.ID, group.ID, "1.0.0"))
		require.NoError(t, err)

		emptyOEMInstance := fetchInstanceByID(t, appWithInstance.ID, group.ID, emptyOEMInstanceID)
		assert.Nil(t, emptyOEMInstance.Oem)
		assert.Nil(t, emptyOEMInstance.AlephVersion)
	})
}

// fetchInstanceByID fetches a single instance from the group instances list
// API by searching for its id, decoding the response through the
// spec-generated client type.
func fetchInstanceByID(t *testing.T, appID, groupID, instanceID string) codegen.Instance {
	t.Helper()

	url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances?status=0&page=1&perpage=10&duration=1d&searchFilter=id&searchValue=%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appID, groupID, instanceID)

	var instances codegen.InstancePage

	httpDo(t, url, "GET", nil, http.StatusOK, "json", &instances)

	require.Len(t, instances.Instances, 1)
	require.Equal(t, instanceID, instances.Instances[0].Id)
	return instances.Instances[0]
}

func TestGetInstanceCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app which has instance
		appWithInstance := getAppWithInstance(t, db)

		// fetch instanceCount from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instancescount?duration=30d", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appWithInstance.ID, appWithInstance.Groups[0].ID)
		method := "GET"

		var instancesCountResp codegen.InstanceCount

		httpDo(t, url, method, nil, http.StatusOK, "json", &instancesCountResp)

		count, err := db.GetInstancesCount(api.InstancesQueryParams{
			ApplicationID: appWithInstance.ID,
			GroupID:       appWithInstance.Groups[0].ID,
		}, "30d")

		require.NoError(t, err)
		assert.Equal(t, uint64(count), instancesCountResp.Count)
	})
}

func TestGetInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(api.Instance{ID: instanceID.String(), Alias: "alias", IP: "0.0.0.0"}, api.NewInstanceApplication(app.ID, app.Groups[0].ID, "0.0.1"))
		require.NoError(t, err)

		// fetch instance from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, app.Groups[0].ID, instanceDB.ID)
		method := "GET"

		var instance api.Instance

		httpDo(t, url, method, nil, http.StatusOK, "json", &instance)

		assert.Equal(t, instanceDB.ID, instance.ID)
		assert.Equal(t, instanceDB.IP, instance.IP)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(api.Instance{ID: instanceID.String(), Alias: "alias", IP: "0.0.0.0"}, api.NewInstanceApplication(app.ID, app.Groups[0].ID, "0.0.1"))
		require.NoError(t, err)

		// fetch instance from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, app.Groups[0].ID, instanceDB.ID)
		method := "GET"

		var instance api.Instance

		httpDo(t, url, method, nil, http.StatusOK, "json", &instance)

		assert.Equal(t, instanceDB.ID, instance.ID)
		assert.Equal(t, instanceDB.IP, instance.IP)
	})
}

func TestGetInstanceStatusHistory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(api.Instance{ID: instanceID.String(), Alias: "alias", IP: "0.0.0.0"}, api.NewInstanceApplication(app.ID, app.Groups[0].ID, "0.0.1"))
		require.NoError(t, err)

		// GetUpdatePackage
		_, err = db.GetUpdatePackage(api.Instance{ID: instanceDB.ID, Alias: instanceDB.Alias, IP: instanceDB.IP}, api.NewInstanceApplication(app.ID, app.Groups[0].ID, instanceDB.Application.Version))
		require.NoError(t, err)

		// create event for instance
		err = db.RegisterEvent(instanceDB.ID, app.ID, app.Groups[0].ID, api.EventUpdateComplete, api.ResultSuccessReboot, "0.0.0", "0")
		require.NoError(t, err)

		// fetch instance status_history
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances/%s/status_history", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, app.Groups[0].ID, instanceDB.ID)
		method := "GET"

		var instanceEvents []api.InstanceStatusHistoryEntry

		httpDo(t, url, method, nil, http.StatusOK, "json", &instanceEvents)

		require.Equal(t, 2, len(instanceEvents))

		assert.Equal(t, api.InstanceStatusComplete, instanceEvents[0].Status)
		assert.Equal(t, api.InstanceStatusUpdateGranted, instanceEvents[1].Status)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(api.Instance{ID: instanceID.String(), Alias: "alias", IP: "0.0.0.0"}, api.NewInstanceApplication(app.ID, app.Groups[0].ID, "0.0.1"))
		require.NoError(t, err)

		// GetUpdatePackage
		_, err = db.GetUpdatePackage(api.Instance{ID: instanceDB.ID, Alias: instanceDB.Alias, IP: instanceDB.IP}, api.NewInstanceApplication(app.ID, app.Groups[0].ID, instanceDB.Application.Version))
		require.NoError(t, err)

		// create event for instance
		err = db.RegisterEvent(instanceDB.ID, app.ID, app.Groups[0].ID, api.EventUpdateComplete, api.ResultSuccessReboot, "0.0.0", "0")
		require.NoError(t, err)

		// fetch instance status_history
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances/%s/status_history", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, app.Groups[0].ID, instanceDB.ID)
		method := "GET"

		var instanceEvents []api.InstanceStatusHistoryEntry

		httpDo(t, url, method, nil, http.StatusOK, "json", &instanceEvents)

		require.Equal(t, 2, len(instanceEvents))

		assert.Equal(t, api.InstanceStatusComplete, instanceEvents[0].Status)
		assert.Equal(t, api.InstanceStatusUpdateGranted, instanceEvents[1].Status)
	})
}

func TestUpdateInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(api.Instance{ID: instanceID.String(), Alias: "alias", IP: "0.0.0.0"}, api.NewInstanceApplication(app.ID, app.Groups[0].ID, "0.0.1"))
		require.NoError(t, err)

		// fetch instance from API
		url := fmt.Sprintf("%s/api/instances/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), instanceDB.ID)
		method := "PUT"

		newAlias := "new_alias"
		payload := strings.NewReader(fmt.Sprintf(`{"alias":"%s"}`, newAlias))

		// response
		var instance api.Instance

		httpDo(t, url, method, payload, http.StatusOK, "json", &instance)

		assert.Equal(t, newAlias, instance.Alias)

		// check alias in DB
		updatedInstanceDB, err := db.GetInstance(instanceDB.ID, app.ID)
		require.NoError(t, err)
		require.NotNil(t, updatedInstanceDB)

		assert.Equal(t, newAlias, updatedInstanceDB.Alias)
	})
}
