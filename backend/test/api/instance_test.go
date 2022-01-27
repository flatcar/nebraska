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

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

func TestListInstances(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get app which has instance
		appWithInstance := getAppWithInstance(t, db)

		// fetch instances from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances?status=0&version=&sort=2&sortOrder=0&page=1&perpage=10&duration=30d", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appWithInstance.ID, appWithInstance.Groups[0].ID)
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
}

func TestGetInstanceCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get app which has instance
		appWithInstance := getAppWithInstance(t, db)

		// fetch instanceCount from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instancescount?duration=30d", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appWithInstance.ID, appWithInstance.Groups[0].ID)
		method := "GET"

		// TODO: will require change as response struct is changed in POC2 branch
		var instancesCount int

		httpDo(t, url, method, nil, http.StatusOK, "json", &instancesCount)

		count, err := db.GetInstancesCount(api.InstancesQueryParams{
			ApplicationID: appWithInstance.ID,
			GroupID:       appWithInstance.Groups[0].ID,
		}, "30d")

		require.NoError(t, err)
		assert.Equal(t, int(count), instancesCount)
	})
}

func TestGetInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(instanceID.String(), "alias", "0.0.0.0", "0.0.1", app.ID, app.Groups[0].ID)
		require.NoError(t, err)

		// fetch instance from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, app.Groups[0].ID, instanceDB.ID)
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

		// get random app
		app := getRandomApp(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(instanceID.String(), "alias", "0.0.0.0", "0.0.1", app.ID, app.Groups[0].ID)
		require.NoError(t, err)

		// GetUpdatePackage
		_, err = db.GetUpdatePackage(instanceDB.ID, instanceDB.Alias, instanceDB.IP, instanceDB.Application.Version, app.ID, app.Groups[0].ID)
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
}

func TestUpdateInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// create instance for app
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(instanceID.String(), "alias", "0.0.0.0", "0.0.1", app.ID, app.Groups[0].ID)
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
