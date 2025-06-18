package auth_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/server"
)

func TestOIDCAuthModeSetup(t *testing.T) {
	t.Run("oidc_server_not_reachable_on_server_setup", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		server, err := server.New(conf, db)
		assert.Nil(t, server)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Error setting up oidc provider")
		assert.Contains(t, err.Error(), "connect: connection refused")
	})

	t.Run("invalid_oidc_server_url", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		var testConfig config.Config
		err := copier.Copy(&testConfig, conf)
		require.NoError(t, err)

		testConfig.OidcIssuerURL = "http://127.0.0.1:8080/"

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		server, err := server.New(&testConfig, db)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "Error setting up oidc provider")
		assert.Contains(t, err.Error(), "404 page not found")

		err = oidcServer.Shutdown()
		require.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		server, err := server.New(conf, db)
		assert.NotNil(t, server)
		assert.NoError(t, err)

		err = oidcServer.Shutdown()
		require.NoError(t, err)
	})
}

var ErrOutOfRetries = errors.New("test: out of retries")

func waitServerReady() (bool, error) {
	retries := 5
	for i := 0; i < retries; i++ {
		if i != 0 {
			time.Sleep(100 * time.Millisecond)
		}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/health", testServerURL), nil)
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

		if (http.StatusOK == resp.StatusCode) && ("OK" == string(bodyBytes)) {
			return true, nil
		}
	}
	return false, ErrOutOfRetries
}
func TestLoginEndpointsNotImplemented(t *testing.T) {
	t.Run("login_endpoint_not_implemented", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 404 since endpoint is removed
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("login_cb_endpoint_not_implemented", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login/cb", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return 404 since endpoint is removed
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid_oidc_client_id", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, string(bodyBytes), `Invalid client id: clientID`)
	})

	t.Run("invalid_oidc_scope", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		startOIDCMockServer(t, oidcServer)

		// change the default config
		var testConfig config.Config
		err := copier.Copy(&testConfig, conf)
		require.NoError(t, err)

		testConfig.OidcScopes = ""

		// start nebraska server
		server, err := server.New(&testConfig, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		client := &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Contains(t, string(bodyBytes), "The request is missing the required parameter: scope")
	})

	t.Run("invalid_callback_state", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		client := &http.Client{
			CheckRedirect: func(req *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())

				// change the state to test the case when the state is not
				// stored in the nebraska server
				if _, ok := req.URL.Query()["state"]; ok {
					query := req.URL.Query()
					query.Set("state", uuid.NewString())
					req.URL.RawQuery = query.Encode()
				}
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		t.Log(resp.StatusCode)
		t.Log(string(bodyBytes))

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("invalid_client_secret", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = uuid.NewString()
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		client := &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("token_exchange_error", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		client := &http.Client{
			CheckRedirect: func(req *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())

				// when the nebraska server tries to exchang the code
				// for token in the /login/cb endpoint this will
				// return an error
				if strings.Contains(req.URL.Path, "/login/cb") {
					oidcServer.QueueError(&mockoidc.ServerError{
						Code:        http.StatusBadRequest,
						Error:       "invalid request",
						Description: "invalid request to exchange token",
					})
					oidcServer.QueueError(&mockoidc.ServerError{
						Code:        http.StatusBadRequest,
						Error:       "invalid request",
						Description: "invalid request to exchange token",
					})
					t.Log("Error queued in mock oidc server")
				}
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("no_roles_in_jwt", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		client := &http.Client{
			Jar: jar,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		t.Log(resp.StatusCode)
		t.Log(string(bodyBytes))
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		accessToken := resp.Request.URL.Query().Get("access_token")
		tokenType := resp.Request.URL.Query().Get("token_type")
		require.NotEmpty(t, accessToken)
		require.Equal(t, "Bearer", tokenType)

		req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/apps", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err = client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("invalid_access", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		oidcServer.QueueUser(&mockoidc.MockUser{
			Groups: []string{"nebraska-member"},
		})
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		client := &http.Client{
			Jar: jar,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		t.Log(resp.StatusCode)
		t.Log(string(bodyBytes))
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		accessToken := resp.Request.URL.Query().Get("access_token")
		tokenType := resp.Request.URL.Query().Get("token_type")
		require.NotEmpty(t, accessToken)
		require.Equal(t, "Bearer", tokenType)

		payload := strings.NewReader(`{"name":"someApp"}`)
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/apps", testServerURL), payload)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err = client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("invalid_token", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		oidcServer.QueueUser(&mockoidc.MockUser{
			Groups: []string{"nebraska-admin"},
		})
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/apps", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", uuid.NewString()))

		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		client := &http.Client{
			Jar: jar,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		oidcServer.QueueUser(&mockoidc.MockUser{
			Groups: []string{"nebraska-admin"},
		})
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		client := &http.Client{
			Jar: jar,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		t.Log(resp.StatusCode)
		t.Log(string(bodyBytes))
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		accessToken := resp.Request.URL.Query().Get("access_token")
		tokenType := resp.Request.URL.Query().Get("token_type")
		require.NotEmpty(t, accessToken)
		require.Equal(t, "Bearer", tokenType)

		req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/apps", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err = client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(bodyBytes), "totalCount")
		assert.Contains(t, string(bodyBytes), "count")
	})
}

func TestValidateToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish db connection
		db := newDBForTest(t)

		// setup and run mock server
		oidcServer := newOIDCMockServer(t)
		oidcServer.ClientID = clientID
		oidcServer.ClientSecret = clientSecret
		oidcServer.QueueUser(&mockoidc.MockUser{
			Groups: []string{"nebraska-admin"},
		})
		startOIDCMockServer(t, oidcServer)

		// start nebraska server
		server, err := server.New(conf, db)
		t.Log(err)
		require.NotNil(t, server)
		require.NoError(t, err)

		//nolint:errcheck
		go server.Start(serverPortStr)

		//nolint:errcheck
		defer server.Shutdown(context.Background())
		//nolint:errcheck
		defer oidcServer.Shutdown()

		_, err = waitServerReady()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/login?login_redirect_url=%s/", testServerURL, testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		client := &http.Client{
			Jar: jar,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// uncomment to debug redirect flow
				// t.Log("req:", req.URL.String())
				return nil
			},
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		t.Log(resp.StatusCode)
		t.Log(string(bodyBytes))
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		accessToken := resp.Request.URL.Query().Get("access_token")
		tokenType := resp.Request.URL.Query().Get("token_type")
		require.NotEmpty(t, accessToken)
		require.Equal(t, "Bearer", tokenType)

		req, err = http.NewRequest("GET", fmt.Sprintf("%s/login/validate_token", testServerURL), nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err = client.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		bodyBytes, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(bodyBytes), "valid")
		assert.Contains(t, string(bodyBytes), "true")
	})
}
