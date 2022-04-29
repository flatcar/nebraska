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
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func TestListGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// get groups from DB for app
		groupsDB, err := db.GetGroups(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, groupsDB)

		// fetch groups from API
		url := fmt.Sprintf("%s/api/apps/%s/groups", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "GET"

		var groupResp codegen.GroupPage

		httpDo(t, url, method, nil, http.StatusOK, "json", &groupResp)

		assert.NotEqual(t, 0, len(groupResp.Groups))
		assert.Equal(t, len(groupsDB), len(groupResp.Groups))

		for i := range groupsDB {
			assert.Equal(t, groupsDB[i].ID, groupResp.Groups[i].Id)
			assert.Equal(t, groupsDB[i].Name, groupResp.Groups[i].Name)
		}
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// get groups from DB for app
		groupsDB, err := db.GetGroups(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, groupsDB)

		// fetch groups from API
		url := fmt.Sprintf("%s/api/apps/%s/groups", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "GET"

		var groupResp codegen.GroupPage

		httpDo(t, url, method, nil, http.StatusOK, "json", &groupResp)

		assert.NotEqual(t, 0, len(groupResp.Groups))
		assert.Equal(t, len(groupsDB), len(groupResp.Groups))

		for i := range groupsDB {
			assert.Equal(t, groupsDB[i].ID, groupResp.Groups[i].Id)
			assert.Equal(t, groupsDB[i].Name, groupResp.Groups[i].Name)
		}
	})

}

func TestCreateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// create group using the API
		url := fmt.Sprintf("%s/api/apps/%s/groups", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "POST"

		groupName := "test_group"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","policy_max_updates_per_period":1,"policy_period_interval":"1 hours","policy_update_timeout":"1 days","policy_timezone":"Asia/Calcutta","application_id":"%s"}`, groupName, app.ID))

		// response
		var group api.Group

		httpDo(t, url, method, payload, http.StatusOK, "json", &group)

		assert.Equal(t, groupName, group.Name)

		// check group exists in DB
		groupDB, err := db.GetGroup(group.ID)
		assert.NoError(t, err)
		assert.NotNil(t, groupDB)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// create group using the API
		url := fmt.Sprintf("%s/api/apps/%s/groups", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "POST"

		groupName := "test_group_product_id"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","policy_max_updates_per_period":1,"policy_period_interval":"1 hours","policy_update_timeout":"1 days","policy_timezone":"Asia/Calcutta","application_id":"%s"}`, groupName, app.ID))

		// response
		var group api.Group

		httpDo(t, url, method, payload, http.StatusOK, "json", &group)

		assert.Equal(t, groupName, group.Name)

		// check group exists in DB
		groupDB, err := db.GetGroup(group.ID)
		assert.NoError(t, err)
		assert.NotNil(t, groupDB)
	})
}

func TestGetGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// fetch group by id request
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, app.Groups[0].ID)
		method := "GET"

		// response
		var group api.Group

		httpDo(t, url, method, nil, http.StatusOK, "json", &group)

		assert.Equal(t, app.Groups[0].Name, group.Name)
	})

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// fetch group by id request
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, app.Groups[0].ID)
		method := "GET"

		// response
		var group api.Group

		httpDo(t, url, method, nil, http.StatusOK, "json", &group)

		assert.Equal(t, app.Groups[0].Name, group.Name)
	})
}

func TestUpdateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// update group request
		var groupDB api.Group
		err := copier.Copy(&groupDB, app.Groups[0])
		require.NoError(t, err)

		groupName := "test_group"
		groupDB.Name = groupName

		payload, err := json.Marshal(groupDB)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, groupDB.ID)
		method := "PUT"

		// response
		var group api.Group
		httpDo(t, url, method, bytes.NewReader(payload), http.StatusOK, "json", &group)

		assert.Equal(t, groupName, group.Name)

		// check name in db
		updatedGroupDB, err := db.GetGroup(groupDB.ID)
		require.NoError(t, err)
		assert.Equal(t, groupName, updatedGroupDB.Name)
	})

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// update group request
		var groupDB api.Group
		err := copier.Copy(&groupDB, app.Groups[0])
		require.NoError(t, err)

		groupName := "test_group"
		groupDB.Name = groupName

		payload, err := json.Marshal(groupDB)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, groupDB.ID)
		method := "PUT"

		// response
		var group api.Group
		httpDo(t, url, method, bytes.NewReader(payload), http.StatusOK, "json", &group)

		assert.Equal(t, groupName, group.Name)

		// check name in db
		updatedGroupDB, err := db.GetGroup(groupDB.ID)
		require.NoError(t, err)
		assert.Equal(t, groupName, updatedGroupDB.Name)
	})

}

func TestDeleteGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		groupDB := app.Groups[0]
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, groupDB.ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		// check if app doesn't exists in DB
		group, err := db.GetGroup(groupDB.ID)
		assert.Error(t, err)
		assert.Nil(t, group)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		groupDB := app.Groups[0]
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, groupDB.ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		// check if app doesn't exists in DB
		group, err := db.GetGroup(groupDB.ID)
		assert.Error(t, err)
		assert.Nil(t, group)
	})

}
