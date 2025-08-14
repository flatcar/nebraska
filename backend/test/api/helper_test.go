package api_test

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
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
	rand.New(rand.NewSource(time.Now().Unix()))
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

func getAppWithProductID(t *testing.T, db *api.API) *api.Application {
	t.Helper()

	apps := getApps(t, db)

	for _, app := range apps {
		if app.ProductID.Valid {
			return app
		}
	}
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

	bodyBytes, err := io.ReadAll(resp.Body)
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

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		if (http.StatusOK == resp.StatusCode) && (string(bodyBytes) == "OK") {
			return true, nil
		}
	}
	return false, ErrOutOfRetries
}

// httpDecodeResponse is a helper function that decodes HTTP response body
// into the provided interface based on Content-Type
func httpDecodeResponse(resp *http.Response, response interface{}) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	contentType := resp.Header.Get("Content-Type")
	switch contentType {
	case "application/json", "":
		return json.Unmarshal(bodyBytes, response)
	case "application/xml":
		return xml.Unmarshal(bodyBytes, response)
	}

	return fmt.Errorf("unsupported content type %q", contentType)
}

// httpMakeRequest is a helper function for making HTTP requests with custom headers
func httpMakeRequest(t *testing.T, method, url string, payload io.Reader, headers map[string]string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, payload)
	require.NoError(t, err)

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}
