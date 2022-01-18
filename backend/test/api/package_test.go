package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
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
		var packagesResp codegen.PackagePage

		httpDo(t, url, method, nil, http.StatusOK, "json", &packagesResp)

		assert.NotEqual(t, 0, len(packagesResp.Packages))
		assert.Equal(t, len(packagesDB), len(packagesResp.Packages))
		for i := range packagesDB {
			assert.Equal(t, packagesDB[i].ApplicationID, packagesResp.Packages[i].ApplicationID)
			assert.Equal(t, packagesDB[i].ID, packagesResp.Packages[i].Id)
		}
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

		if packageDB.ChannelsBlacklist == nil {
			packageDB.ChannelsBlacklist = []string{}
		}
		if packageDB.Description.IsZero() {
			packageDB.Description = null.StringFrom("some desc")
		}
		if packageDB.Size.IsZero() {
			packageDB.Size = null.StringFrom("20")
		}
		if packageDB.Hash.IsZero() {
			packageDB.Hash = null.StringFrom(uuid.New().String())
		}

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
