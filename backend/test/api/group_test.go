package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app from DB
		app := getRandomApp(t, db)

		// get groups from DB for app
		groupsDB, err := db.GetGroups(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, groupsDB)

		// fetch groups from API
		url := fmt.Sprintf("%s/api/apps/%s/groups", testServerURL, app.ID)
		method := "GET"

		// response
		// TODO: will require change as response struct is changed in POC2 branch
		var groups []*api.Group

		httpDo(t, url, method, nil, http.StatusOK, "json", &groups)

		assert.NotEqual(t, 0, len(groups))
		assert.Equal(t, len(groupsDB), len(groups))
	})
}

func TestCreateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app from DB
		app := getRandomApp(t, db)

		// create group using the API
		url := fmt.Sprintf("%s/api/apps/%s/groups", testServerURL, app.ID)
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
}

func TestGetGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get random app from DB
		app := getRandomApp(t, db)

		// fetch group by id request
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", testServerURL, app.ID, app.Groups[0].ID)
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

		// get random app from DB
		app := getRandomApp(t, db)

		// update group request
		var groupDB api.Group
		copier.Copy(&groupDB, app.Groups[0])

		groupName := "test_group"
		groupDB.Name = groupName

		payload, err := json.Marshal(groupDB)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", testServerURL, groupDB.ApplicationID, groupDB.ID)
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

		// get random app from DB
		app := getRandomApp(t, db)

		groupDB := app.Groups[0]
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s", testServerURL, groupDB.ApplicationID, groupDB.ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		// check if app doesn't exists in DB
		group, err := db.GetGroup(groupDB.ID)
		assert.Error(t, err)
		assert.Nil(t, group)
	})
}
