package api_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/server"
)

const (
	testServerURL = "http://localhost:6000"
	serverPort    = uint(6000)
)

var serverPortStr = fmt.Sprintf(":%d", serverPort)

var conf = &config.Config{
	EnableSyncer: true,
	NebraskaURL:  testServerURL,
	HTTPLog:      true,
	AuthMode:     "noop",
	Debug:        true,
	ServerPort:   serverPort,
}

func TestAPIEndpointSecret(t *testing.T) {
	// establish db connection
	db := newDBForTest(t)
	defer db.Close()

	app := getAppWithInstance(t, db)

	// increase max update for the group
	group := app.Groups[0]
	group.PolicyMaxUpdatesPerPeriod = 1000
	err := db.UpdateGroup(group)
	require.NoError(t, err)

	tt := []struct {
		name               string
		secret             string
		url                string
		expectedStatusCode int
	}{
		{
			"success_with_slash_as_secret",
			"/",
			fmt.Sprintf("%s/v1/update", testServerURL),
			http.StatusOK,
		},
		{
			"success_with_slash_as_secret_and_path",
			"/",
			fmt.Sprintf("%s/v1/update/", testServerURL),
			http.StatusOK,
		},
		{
			"success_secret_with_no_pre_slash",
			"test/this",
			fmt.Sprintf("%s/v1/update/test/this", testServerURL),
			http.StatusOK,
		},
		{
			"success_secret_with_two_pre_slash",
			"//test/this",
			fmt.Sprintf("%s/v1/update//test/this", testServerURL),
			http.StatusOK,
		},
		{
			"success_secret_with_two_pre_slash_and_path_with_trailing_slash",
			"//test/this",
			fmt.Sprintf("%s/v1/update//test/this/", testServerURL),
			http.StatusOK,
		},
		{
			"success_with_two_trailing_slash",
			"/test//",
			fmt.Sprintf("%s/v1/update/test//", testServerURL),
			http.StatusOK,
		},
		{
			"success_with_secret",
			"/test",
			fmt.Sprintf("%s/v1/update/test", testServerURL),
			http.StatusOK,
		},
		{
			"failure_with_secret",
			"/test",
			fmt.Sprintf("%s/v1/update/failure", testServerURL),
			http.StatusNotImplemented,
		},
		{
			"success_secret_and_path_with_trailing_slash",
			"/test/",
			fmt.Sprintf("%s/v1/update/test/", testServerURL),
			http.StatusOK,
		},
		{
			"success_secret_with_trailing_slash",
			"/test/",
			fmt.Sprintf("%s/v1/update/test", testServerURL),
			http.StatusOK,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			track := group.Track

			var testConfig config.Config
			err := copier.Copy(&testConfig, conf)
			require.NoError(t, err)

			testConfig.APIEndpointSuffix = tc.secret
			server, err := server.New(&testConfig, db)
			assert.NoError(t, err)

			//nolint:errcheck
			go server.Start(serverPortStr)

			//nolint:errcheck
			defer server.Shutdown(context.Background())
			_, err = waitServerReady(testConfig.NebraskaURL)
			require.NoError(t, err)

			method := "POST"

			instanceID := uuid.New().String()
			payload := strings.NewReader(fmt.Sprintf(`
		<request protocol="3.0" installsource="scheduler">
		<os platform="CoreOS" version="lsb"></os>
		<app appid="%s" version="0.0.0" track="%s" bootid="3c9c0e11-6c37-4e47-9f60-6d06b421286d" machineid="%s" oem="fakeclient">
		 <ping r="1" status="1"></ping>
		</app>
	   	</request>`, app.ID, track, instanceID))

			// response
			if tc.expectedStatusCode == http.StatusOK {
				var omahaResp omaha.Response
				httpDo(t, tc.url, method, payload, tc.expectedStatusCode, "xml", &omahaResp)
				assert.Equal(t, "ok", omahaResp.Apps[0].Ping.Status)
			} else {
				httpDo(t, tc.url, method, payload, tc.expectedStatusCode, "", nil)
			}
		})
	}
}
