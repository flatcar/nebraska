package api_test

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

// newDBForTest is a helper function that
// establishes connection with test db and returns the db struct.
func newDBForTest(t *testing.T) *api.API {
	t.Helper()
	db, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	require.NotNil(t, db)
	return db
}

// getTeamID is a helper function that
// takes the db connection and returns the default teamID.
func getTeamID(t *testing.T, db *api.API) string {
	t.Helper()
	team, err := db.GetTeam()
	require.NoError(t, err)
	require.NotNil(t, team)
	return team.ID
}

// getApps is a helper function that
// takes an active db connection and retuns the first 10 applications.
func getApps(t *testing.T, db *api.API) []*api.Application {
	t.Helper()

	teamID := getTeamID(t, db)
	apps, err := db.GetApps(teamID, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, apps)
	return apps
}

// getRandomApp is a helper function that
// takes an active db connection and returns a random app.
func getRandomApp(t *testing.T, db *api.API) *api.Application {
	t.Helper()
	apps := getApps(t, db)
	rand.Seed(time.Now().Unix())
	return apps[rand.Intn(len(apps))]
}

// getAppWithInstance is a helper function that
// takes an active db connection and returns an app
// that has instances.
func getAppWithInstance(t *testing.T, db *api.API) *api.Application {
	t.Helper()

	apps := getApps(t, db)

	for _, app := range apps {
		if app.Instances.Count != 0 {
			return app
		}
	}
	t.Error("couldn't get app with instance")
	return nil
}

// httpDo is a helper function that takes all request related info and
// makes the http request and returns the unmarshalled response body based
// on the responseType.
func httpDo(t *testing.T, url string, method string, payload io.Reader, statuscode int, responseType string, response interface{}) {
	t.Helper()

	req, err := http.NewRequest(method, url, payload)
	require.NoError(t, err)
	require.NotNil(t, req)

	req.Header = http.Header{
		"Content-Type": []string{fmt.Sprintf("application/%s", responseType)},
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	if statuscode != 0 {
		require.Equal(t, statuscode, resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	switch responseType {
	case "json":
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
	case "xml":
		err = xml.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
	}
}

var ErrOutOfRetries = errors.New("test: out of retries")

func waitServerReady(serverURL string) (bool, error) {
	retries := 5
	for i := 0; i < retries; i++ {
		if i != 0 {
			time.Sleep(100 * time.Millisecond)
		}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/health", serverURL), nil)
		if err != nil {
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		if (http.StatusOK == resp.StatusCode) && ("OK" == string(bodyBytes)) {
			return true, nil
		}
	}
	return false, ErrOutOfRetries
}
