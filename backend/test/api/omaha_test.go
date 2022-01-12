package api_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kinvolk/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOmaha(t *testing.T) {
	// establish db connection
	db := newDBForTest(t)

	app := getAppWithInstance(t, db)

	t.Run("success", func(t *testing.T) {
		track := app.Groups[0].Track

		url := fmt.Sprintf("%s/omaha", testServerURL)

		method := "POST"

		instanceID := uuid.New().String()
		payload := strings.NewReader(fmt.Sprintf(`
		<request protocol="3.0" installsource="scheduler">
		<os platform="CoreOS" version="lsb"></os>
		<app appid="%s" version="0.0.0" track="%s" bootid="3c9c0e11-6c37-4e47-9f60-6d06b421286d" machineid="%s" oem="fakeclient">
		 <ping r="1" status="1"></ping>
		 <updatecheck></updatecheck>
		 <event eventtype="3" eventresult="1"></event>
		</app>
	   	</request>`, app.ID, track, instanceID))

		// response
		var omahaResp omaha.Response

		httpDo(t, url, method, payload, 200, "xml", &omahaResp)

		assert.Equal(t, "ok", omahaResp.Apps[0].Ping.Status)
		assert.Equal(t, "omaha: update status ok", omahaResp.Apps[0].UpdateCheck.Status.Error())

		// check if instance exists in the DB
		instance, err := db.GetInstance(instanceID, app.ID)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("large_request_body", func(t *testing.T) {
		url := fmt.Sprintf("%s/omaha", testServerURL)

		method := "POST"

		payload, err := ioutil.ReadFile("./big_omaha_request.xml")
		require.NoError(t, err)

		httpDo(t, url, method, bytes.NewReader(payload), 400, "", nil)
	})
}
