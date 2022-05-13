package api_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
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
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		method := "GET"

		var appsResp codegen.AppsPage
		httpDo(t, url, method, nil, http.StatusOK, "json", &appsResp)

		assert.NotEqual(t, len(appsDB), 0)
		assert.Equal(t, len(appsDB), len(appsResp.Applications))
		for i := range appsDB {
			assert.Equal(t, appsDB[i].ID, appsResp.Applications[i].Id)
			assert.Equal(t, appsDB[i].Name, appsResp.Applications[i].Name)
		}
	})
}

func TestCreateApp(t *testing.T) {
	t.Run("success_do_not_copy", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// Create App request
		url := fmt.Sprintf("%s%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), "/api/apps")
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
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		app := getRandomApp(t, db)

		// Create App request
		url := fmt.Sprintf("%s/api/apps?clone_from=%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "POST"

		appName := "test_with_clone"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		// response
		var application api.Application

		httpDo(t, url, method, payload, http.StatusOK, "json", &application)

		// close and create new db session to clear and update cached appIds
		db.Close()
		db, err := api.New()
		require.NoError(t, err)
		require.NotNil(t, db)

		// check if app exists in DB
		app, err = db.GetApp(application.ID)
		require.NoError(t, err)

		assert.Equal(t, application.ID, app.ID)
	})

	t.Run("success_with_copy_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		app := getAppWithProductID(t, db)

		// Create App request
		url := fmt.Sprintf("%s/api/apps?clone_from=%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "POST"

		appName := "test_with_clone"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		// response
		var application api.Application

		httpDo(t, url, method, payload, http.StatusOK, "json", &application)

		// close and create new db session to clear and update cached appIds
		db.Close()
		db, err := api.New()
		require.NoError(t, err)
		require.NotNil(t, db)

		// check if app exists in DB
		app, err = db.GetApp(application.ID)
		require.NoError(t, err)

		assert.Equal(t, application.ID, app.ID)
	})
}

func TestGetApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		app := getRandomApp(t, db)

		// fetch app by id request
		url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "GET"

		// check response
		var application api.Application

		httpDo(t, url, method, nil, http.StatusOK, "json", &application)

		assert.Equal(t, app.Name, application.Name)
		assert.Equal(t, app.Instances, application.Instances)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from db
		app := getAppWithProductID(t, db)

		// fetch app by product_id request
		url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "GET"

		// check response
		var application codegen.Application

		httpDo(t, url, method, nil, http.StatusOK, "json", &application)

		assert.Equal(t, app.Name, application.Name)
		assert.Equal(t, app.ProductID.String, application.ProductId)
		assert.Equal(t, app.ID, application.Id)
	})
}

func TestUpdateApp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB to update
		app := getRandomApp(t, db)

		// Update App Request
		url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
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

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product_id from DB to update
		app := getAppWithProductID(t, db)

		// Update App Request
		url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "PUT"

		name := "updated_name_product_id"
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
		defer db.Close()

		// get random app from db
		app := getRandomApp(t, db)

		// Update App Request
		url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		// check if app exists in db
		app, err := db.GetApp(app.ID)
		assert.Error(t, err)
		assert.Nil(t, app)
	})

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from db
		app := getAppWithProductID(t, db)

		// Update App Request
		url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		// check if app exists in db
		app, err := db.GetApp(app.ID)
		assert.Error(t, err)
		assert.Nil(t, app)
	})
}
