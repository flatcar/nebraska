package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

func TestListApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// fetch teamID
		teamID := getTeamID(t, db)

		// get apps from DB
		appsDB, err := db.GetApps(teamID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, appsDB)

		// fetch apps from the API
		url := fmt.Sprintf("%s/api/apps", testServerURL)
		method := "GET"

		// TODO: will require change as response struct is changed in POC2 branch
		var appsResp []*api.Application

		httpDo(t, url, method, nil, http.StatusOK, "json", &appsResp)

		assert.NotEqual(t, len(appsDB), 0)
		assert.Equal(t, len(appsDB), len(appsResp))

		for i := range appsDB {
			assert.Equal(t, appsResp[i].ID, appsDB[i].ID)
			assert.Equal(t, appsResp[i].Name, appsDB[i].Name)
		}
	})
}

func TestCreateApp(t *testing.T) {
	// establish DB connection
	db := newDBForTest(t)

	t.Run("success_do_not_copy", func(t *testing.T) {
		// Create App request
		url := fmt.Sprintf("%s%s", testServerURL, "/api/apps")
		method := "POST"

		appName := "test"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		// response struct
		var application api.Application

		httpDo(t, url, method, payload, http.StatusOK, "json", &application)

		assert.Equal(t, appName, application.Name)

		// check if app exists in DB
		app, err := db.GetApp(application.ID)
		require.NoError(t, err)

		assert.Equal(t, application.ID, app.ID)
	})

	t.Run("success_with_copy", func(t *testing.T) {
		app := getRandomApp(t, db)

		// Create App request
		url := fmt.Sprintf("%s/api/apps?clone_from=%s", testServerURL, app.ID)
		method := "POST"

		appName := "test_with_clone"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		// response
		var application api.Application

		httpDo(t, url, method, payload, http.StatusOK, "json", &application)

		// check if app exists in DB
		app, err := db.GetApp(application.ID)
		require.NoError(t, err)

		assert.Equal(t, application.ID, app.ID)
	})
}

func TestGetApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		app := getRandomApp(t, db)

		// fetch app by id request
		url := fmt.Sprintf("%s/api/apps/%s", testServerURL, app.ID)
		method := "GET"

		// check response
		var application api.Application

		httpDo(t, url, method, nil, http.StatusOK, "json", &application)

		assert.Equal(t, app.Name, application.Name)
		assert.Equal(t, app.Instances, application.Instances)
	})
}

func TestUpdateApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app from DB to update
		app := getRandomApp(t, db)

		// Update App Request
		url := fmt.Sprintf("%s/api/apps/%s", testServerURL, app.ID)
		method := "PUT"

		name := "updated_name"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","description":"%s","id":"%s"}`, name, app.Description, app.ID))

		// response struct
		var application api.Application

		httpDo(t, url, method, payload, http.StatusOK, "json", &application)

		assert.Equal(t, name, application.Name)

		// check name in DB

		app, err := db.GetApp(app.ID)
		require.NoError(t, err)

		assert.Equal(t, name, app.Name)
	})
}

func TestDeleteApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app from db
		app := getRandomApp(t, db)

		// Update App Request
		url := fmt.Sprintf("%s/api/apps/%s", testServerURL, app.ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		// check if app exists in db
		app, err := db.GetApp(app.ID)
		assert.Error(t, err)
		assert.Nil(t, app)
	})
}
