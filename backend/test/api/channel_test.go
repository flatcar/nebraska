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

func TestListChannels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// get channels from DB for app
		channelsDB, err := db.GetChannels(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, channelsDB)

		// fetch channels from API
		url := fmt.Sprintf("%s/api/apps/%s/channels", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "GET"

		var channelsResp codegen.ChannelPage

		httpDo(t, url, method, nil, http.StatusOK, "json", &channelsResp)

		for i := range channelsDB {
			assert.NotEqual(t, 0, len(channelsResp.Channels))
			assert.Equal(t, len(channelsDB), len(channelsResp.Channels))
			assert.Equal(t, channelsDB[i].ApplicationID, channelsResp.Channels[i].ApplicationID)
		}
	})

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// get channels from DB for app
		channelsDB, err := db.GetChannels(app.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, channelsDB)

		// fetch channels from API
		url := fmt.Sprintf("%s/api/apps/%s/channels", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "GET"

		var channelsResp codegen.ChannelPage

		httpDo(t, url, method, nil, http.StatusOK, "json", &channelsResp)

		for i := range channelsDB {
			assert.NotEqual(t, 0, len(channelsResp.Channels))
			assert.Equal(t, len(channelsDB), len(channelsResp.Channels))
			assert.Equal(t, channelsDB[i].ApplicationID, channelsResp.Channels[i].ApplicationID)
		}
	})
}

func TestCreateChannel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// create channel using the API
		url := fmt.Sprintf("%s/api/apps/%s/channels", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		method := "POST"

		channelName := "test_channel"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","arch":0,"color":"","application_id":"%s"}`, channelName, app.ID))

		var channel api.Channel

		httpDo(t, url, method, payload, http.StatusOK, "json", &channel)

		assert.Equal(t, channelName, channel.Name)

		// check channel exists in DB
		channelsDB, err := db.GetChannel(channel.ID)
		assert.NoError(t, err)
		assert.NotNil(t, channelsDB)
		assert.Equal(t, channelName, channelsDB.Name)
	})

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// create channel using the API
		url := fmt.Sprintf("%s/api/apps/%s/channels", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String)
		method := "POST"

		channelName := "test_channel"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","arch":0,"color":"","application_id":"%s"}`, channelName, app.ID))

		var channel api.Channel

		httpDo(t, url, method, payload, http.StatusOK, "json", &channel)

		assert.Equal(t, channelName, channel.Name)

		// check channel exists in DB
		channelsDB, err := db.GetChannel(channel.ID)
		assert.NoError(t, err)
		assert.NotNil(t, channelsDB)
		assert.Equal(t, channelName, channelsDB.Name)
	})
}

func TestGetChannel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// fetch channel by id request
		url := fmt.Sprintf("%s/api/apps/%s/channels/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, app.Channels[0].ID)
		method := "GET"

		// response
		var channel api.Channel

		httpDo(t, url, method, nil, http.StatusOK, "json", &channel)

		assert.Equal(t, app.Channels[0].Name, channel.Name)
		assert.Equal(t, app.Channels[0].ID, channel.ID)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		// fetch channel by id request
		url := fmt.Sprintf("%s/api/apps/%s/channels/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, app.Channels[0].ID)
		method := "GET"

		// response
		var channel api.Channel

		httpDo(t, url, method, nil, http.StatusOK, "json", &channel)

		assert.Equal(t, app.Channels[0].Name, channel.Name)
		assert.Equal(t, app.Channels[0].ID, channel.ID)
	})
}

func TestUpdateChannel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		// update channel request
		var channelDB api.Channel
		err := copier.Copy(&channelDB, app.Channels[0])
		require.NoError(t, err)

		channelName := "test_channel"
		channelDB.Name = channelName

		payload, err := json.Marshal(channelDB)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/api/apps/%s/channels/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, channelDB.ID)
		method := "PUT"

		// response
		var channel api.Channel

		httpDo(t, url, method, bytes.NewReader(payload), http.StatusOK, "json", &channel)
		assert.Equal(t, channelName, channel.Name)

		// check name in DB
		updatedChannelDB, err := db.GetChannel(channel.ID)
		require.NoError(t, err)

		assert.Equal(t, channelName, updatedChannelDB.Name)
	})

	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)
		defer db.Close()

		// update channel request
		var channelDB api.Channel
		err := copier.Copy(&channelDB, app.Channels[0])
		require.NoError(t, err)

		channelName := "test_channel"
		channelDB.Name = channelName

		payload, err := json.Marshal(channelDB)
		require.NoError(t, err)

		url := fmt.Sprintf("%s/api/apps/%s/channels/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, channelDB.ID)
		method := "PUT"

		// response
		var channel api.Channel

		httpDo(t, url, method, bytes.NewReader(payload), http.StatusOK, "json", &channel)
		assert.Equal(t, channelName, channel.Name)

		// check name in DB
		updatedChannelDB, err := db.GetChannel(channel.ID)
		require.NoError(t, err)

		assert.Equal(t, channelName, updatedChannelDB.Name)
	})
}

func TestDeleteChannel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from DB
		app := getRandomApp(t, db)

		channelDB := app.Channels[0]
		url := fmt.Sprintf("%s/api/apps/%s/channels/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, channelDB.ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		channel, err := db.GetChannel(channelDB.ID)
		assert.Error(t, err)
		assert.Nil(t, channel)
	})
	t.Run("success_product_id", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get app with product id from DB
		app := getAppWithProductID(t, db)

		channelDB := app.Channels[0]
		url := fmt.Sprintf("%s/api/apps/%s/channels/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ProductID.String, channelDB.ID)
		method := "DELETE"

		httpDo(t, url, method, nil, http.StatusNoContent, "", nil)

		channel, err := db.GetChannel(channelDB.ID)
		assert.Error(t, err)
		assert.Nil(t, channel)
	})
}
