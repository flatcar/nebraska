package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

func TestListPackages(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// fetch packages from API
		url := fmt.Sprintf("%s/api/apps/%s/packages", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "GET"

		// response
		var packagesResp codegen.PackagePage

		httpDo(t, url, method, nil, http.StatusOK, "json", &packagesResp)

		assert.NotEqual(t, 0, len(packagesResp.Packages))
		assert.Equal(t, len(packagesDB), len(packagesResp.Packages))
		for i := range packagesDB {
			assert.Equal(t, packagesDB[i].ApplicationID, packagesResp.Packages[i].ApplicationID)
			assert.Equal(t, packagesDB[i].ID, packagesResp.Packages[i].Id)
		}
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// get packages from DB for app
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// fetch packages from API
		url := fmt.Sprintf("%s/api/apps/%s/packages", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "GET"

		// response
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
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// create group using the API
		url := fmt.Sprintf("%s/api/apps/%s/packages", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "POST"

		packageName := "test_package"
		payload := strings.NewReader(fmt.Sprintf(`{"arch":1,"filename":"%s","description":"flatcar package","url":"http://flatcar.org","version":"20.2.1","type":4,"size":"199","hash":"some random hash","application_id":"%s","channels_blacklist":[]}`, packageName, app.ID))

		// response
		var packageResp types.Package

		httpDo(t, url, method, payload, http.StatusOK, "json", &packageResp)

		assert.Equal(t, packageName, packageResp.Filename.String)

		// check group exists in DB
		packageDB, err := db.GetPackage(packageResp.ID)
		assert.NoError(t, err)
		assert.NotNil(t, packageDB)

		assert.Equal(t, packageName, packageDB.Filename.String)
	})

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// create group using the API
		url := fmt.Sprintf("%s/api/apps/%s/packages", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "POST"

		packageName := "test_package"
		payload := strings.NewReader(fmt.Sprintf(`{"arch":1,"filename":"%s","description":"flatcar package","url":"http://flatcar.org","version":"20.2.4","type":4,"size":"199","hash":"some random hash","application_id":"%s","channels_blacklist":[]}`, packageName, app.ID))

		// response
		var packageResp types.Package

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
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// fetch group by id request
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, packagesDB[0].ID)
		method := "GET"

		var packageResp types.Package

		httpDo(t, url, method, nil, http.StatusOK, "json", &packageResp)

		assert.Equal(t, packagesDB[0].Filename, packageResp.Filename)
		assert.Equal(t, packagesDB[0].ID, packageResp.ID)
	})
	t.Run("success_package_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// fetch group by id request
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, packagesDB[0].ID)
		method := "GET"

		var packageResp types.Package

		httpDo(t, url, method, nil, http.StatusOK, "json", &packageResp)

		assert.Equal(t, packagesDB[0].Filename, packageResp.Filename)
		assert.Equal(t, packagesDB[0].ID, packageResp.ID)
	})
}

func TestUpdatePackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// update package request
		var packageDB types.Package
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
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, packageDB.ID)
		method := "PUT"

		var packageResp types.Package

		httpDo(t, url, method, bytes.NewReader(payload), http.StatusOK, "json", &packageResp)

		assert.Equal(t, packageVersion, packageResp.Version)

		// check package version in DB
		updatedPackageDB, err := db.GetPackage(packageDB.ID)
		require.NoError(t, err)
		assert.Equal(t, packageVersion, updatedPackageDB.Version)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// update package request
		var packageDB types.Package
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
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, packageDB.ID)
		method := "PUT"

		var packageResp types.Package

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
		defer db.Close()

		// get random app
		app := getRandomApp(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// delte package by id request
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, packagesDB[0].ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		packageDB, err := db.GetPackage(packagesDB[0].ID)
		assert.Error(t, err)
		assert.Nil(t, packageDB)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// get packages from DB for app[0]
		packagesDB, err := db.GetPackages(app.ID, 1, 10, nil)
		require.NoError(t, err)
		require.NotNil(t, packagesDB)

		// delte package by id request
		url := fmt.Sprintf("%s/api/apps/%s/packages/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, packagesDB[0].ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		packageDB, err := db.GetPackage(packagesDB[0].ID)
		assert.Error(t, err)
		assert.Nil(t, packageDB)
	})
}

func TestPackageLabels(t *testing.T) {
	db := newDBForTest(t)
	defer db.Close()

	app := getRandomApp(t, db)
	serverURL := os.Getenv("NEBRASKA_TEST_SERVER_URL")
	packagesURL := fmt.Sprintf("%s/api/apps/%s/packages", serverURL, app.ID)
	body := func(labels string) *strings.Reader {
		return strings.NewReader(fmt.Sprintf(`{"arch":1,"filename":"labels.img","description":"labels","url":"http://example.com/","version":"20.3.1","type":4,"size":"10","hash":"","application_id":"%s","channels_blacklist":[]%s}`, app.ID, labels))
	}
	// rejected checks that the request fails with 400 and a JSON message,
	// whether the OpenAPI validator or Nebraska itself refuses it.
	rejected := func(t *testing.T, method, url string, payload *strings.Reader) {
		t.Helper()
		resp := httpMakeRequest(t, method, url, payload, map[string]string{"Content-Type": "application/json"})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var errResp struct {
			Message string `json:"message"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
		assert.NotEmpty(t, errResp.Message)
	}

	labels := types.Labels{"image.metadata.sha384": strings.Repeat("ab", 48), "build": "42"}
	labelsJSON, err := json.Marshal(labels)
	require.NoError(t, err)

	var created types.Package
	httpDo(t, packagesURL, "POST", body(`,"labels":`+string(labelsJSON)), http.StatusOK, "json", &created)
	assert.Equal(t, labels, created.Labels)

	packageURL := fmt.Sprintf("%s/%s", packagesURL, created.ID)

	t.Run("get_returns_labels", func(t *testing.T) {
		var got types.Package
		httpDo(t, packageURL, "GET", nil, http.StatusOK, "json", &got)
		assert.Equal(t, labels, got.Labels)
	})

	t.Run("update_without_labels_keeps_them", func(t *testing.T) {
		var updated types.Package
		httpDo(t, packageURL, "PUT", body(""), http.StatusOK, "json", &updated)
		assert.Equal(t, labels, updated.Labels)
	})

	t.Run("ui_edit_dialog_payload_keeps_them", func(t *testing.T) {
		// The fields frontend/src/components/Packages/EditDialog.tsx sends.
		payload := fmt.Sprintf(`{"arch":1,"filename":"labels.img","description":"edited in the UI","url":"http://example.com/","version":"20.3.1","type":4,"size":"10","hash":"","application_id":"%s","channels_blacklist":[],"extra_files":[],"id":"%s"}`, app.ID, created.ID)
		var updated types.Package
		httpDo(t, packageURL, "PUT", strings.NewReader(payload), http.StatusOK, "json", &updated)
		assert.Equal(t, "edited in the UI", updated.Description.String)
		assert.Equal(t, labels, updated.Labels)
	})

	t.Run("invalid_labels_are_rejected", func(t *testing.T) {
		tooMany := map[string]string{}
		for i := 0; i <= types.MaxLabels; i++ {
			tooMany[fmt.Sprintf("k%d", i)] = "v"
		}
		tooManyJSON, err := json.Marshal(tooMany)
		require.NoError(t, err)

		for name, labelsField := range map[string]string{
			"invalid_key":       `,"labels":{"a b":"v"}`,
			"key_with_slash":    `,"labels":{"a/b":"v"}`,
			"control_character": `,"labels":{"k":"a\u0001b"}`,
			"value_too_long":    `,"labels":{"k":"` + strings.Repeat("v", types.MaxLabelValueLength+1) + `"}`,
			"too_many_labels":   `,"labels":` + string(tooManyJSON),
			"null":              `,"labels":null`,
			"number_value":      `,"labels":{"k":1}`,
			"array":             `,"labels":["k"]`,
		} {
			t.Run(name, func(t *testing.T) {
				rejected(t, "POST", packagesURL, body(labelsField))
				rejected(t, "PUT", packageURL, body(labelsField))
			})
		}

		var got types.Package
		httpDo(t, packageURL, "GET", nil, http.StatusOK, "json", &got)
		assert.Equal(t, labels, got.Labels, "a rejected update must not change the labels")
	})

	t.Run("update_check_returns_labels", func(t *testing.T) {
		var channel types.Channel
		httpDo(t, fmt.Sprintf("%s/api/apps/%s/channels", serverURL, app.ID), "POST", strings.NewReader(fmt.Sprintf(`{"name":"labels","arch":1,"color":"","application_id":"%s","package_id":"%s"}`, app.ID, created.ID)), http.StatusOK, "json", &channel)
		var group types.Group
		httpDo(t, fmt.Sprintf("%s/api/apps/%s/groups", serverURL, app.ID), "POST", strings.NewReader(fmt.Sprintf(`{"name":"labels","track":"labels","channel_id":"%s","policy_updates_enabled":true,"policy_max_updates_per_period":100,"policy_period_interval":"1 hours","policy_update_timeout":"1 days","policy_timezone":"UTC","application_id":"%s"}`, channel.ID, app.ID)), http.StatusOK, "json", &group)

		request := fmt.Sprintf(`<request protocol="3.0"><os platform="CoreOS" version="lsb" arch="x64"></os><app appid="%s" version="0.0.0" track="%s" machineid="labels-test-machine"><updatecheck></updatecheck></app></request>`, app.ID, group.Track)
		var resp omaha.Response
		httpDo(t, fmt.Sprintf("%s/v1/update", serverURL), "POST", strings.NewReader(request), http.StatusOK, "xml", &resp)

		require.Len(t, resp.Apps, 1)
		require.NotNil(t, resp.Apps[0].UpdateCheck)
		require.Equal(t, omaha.UpdateOK, resp.Apps[0].UpdateCheck.Status)
		pkgLabels := resp.Apps[0].UpdateCheck.Manifests[0].Packages[0].Labels
		value, ok := pkgLabels.Get("image.metadata.sha384")
		assert.True(t, ok)
		assert.Equal(t, labels["image.metadata.sha384"], value)
	})

	t.Run("update_with_empty_labels_removes_them", func(t *testing.T) {
		var updated types.Package
		httpDo(t, packageURL, "PUT", body(`,"labels":{}`), http.StatusOK, "json", &updated)
		assert.Empty(t, updated.Labels)
	})
}
