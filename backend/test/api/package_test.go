package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

func TestListPackages(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app
		packagesDB, err := db.GetPackages(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// fetch packages from API
		url := fmt.Sprintf("%s/api/apps/%s/packages", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "GET"

		// response
		// TODO: will require change as response struct is changed in POC2 branch
		var packages []*api.Package

		httpDo(t, url, method, nil, http.StatusOK, "json", &packages)

		assert.NotEqual(t, 0, len(packages))
		assert.Equal(t, len(packagesDB), len(packages))
	})
}

func TestCreatePackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// create group using the API
		url := fmt.Sprintf("%s/api/apps/%s/packages", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "POST"

		packageName := "test_package"
		payload := strings.NewReader(fmt.Sprintf(`{"arch":1,"filename":"%s","description":"kinvolk package","url":"http://kinvolk.io","version":"20.2.1","type":4,"size":"199","hash":"some random hash","application_id":"%s","channels_blacklist":[]}`, packageName, app.ID))

		// response
		var packageResp api.Package

		httpDo(t, url, method, payload, http.StatusOK, "json", &packageResp)

		assert.Equal(t, packageName, packageResp.Filename.String)

		// check group exists in DB
		packageDB, err := db.GetPackage(packageResp.ID)
		assert.NoError(t, err)
		assert.NotNil(t, packageDB)

		assert.Equal(t, packageName, packageDB.Filename.String)
	})
}

func TestGetPackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// fetch group by id request
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), packagesDB[0].ApplicationID, packagesDB[0].ID)
		method := "GET"

		var packageResp api.Package

		httpDo(t, url, method, nil, http.StatusOK, "json", &packageResp)

		assert.Equal(t, packagesDB[0].Filename, packageResp.Filename)
		assert.Equal(t, packagesDB[0].ID, packageResp.ID)
	})
}

func TestUpdatePackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// update package request
		var packageDB api.Package
		err = copier.Copy(&packageDB, packagesDB[0])
		require.NoError(t, err)

		packageVersion := "20.2.2"
		packageDB.Version = packageVersion

		payload, err := json.Marshal(packageDB)
		require.NoError(t, err)

		// fetch group by id request
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), packageDB.ApplicationID, packageDB.ID)
		method := "PUT"

		var packageResp api.Package

		httpDo(t, url, method, bytes.NewReader(payload), http.StatusOK, "json", &packageResp)

		assert.Equal(t, packageVersion, packageResp.Version)

		// check package version in DB
		updatedPackageDB, err := db.GetPackage(packageDB.ID)
		require.NoError(t, err)
		assert.Equal(t, packageVersion, updatedPackageDB.Version)
	})
}

func TestDeletePackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// delte package by id request
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), packagesDB[0].ApplicationID, packagesDB[0].ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		packageDB, err := db.GetPackage(packagesDB[0].ID)
		assert.Error(t, err)
		assert.Nil(t, packageDB)
	})
}
