package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/cmd/nebraska/auth"
	"github.com/kinvolk/nebraska/pkg/api"
)

func TestGetRequestIP(t *testing.T) {
	testCases := []struct {
		remoteAddr     string
		xForwardedFor  string
		expectedOutput string
	}{
		{"", "", ""},
		{"1.1.1.1:12345", "", "1.1.1.1"},
		{"1.1.1.1:12345", "2.2.2.2.2", "1.1.1.1"},
		{"1.1.1.1:12345", "2.2.2.2", "2.2.2.2"},
		{"1.1.1.1:12345", "3.3.3.3, 4.4.4.4", "3.3.3.3"},
	}

	for _, tc := range testCases {
		r, _ := http.NewRequest("POST", "/v1/update", nil)
		r.RemoteAddr = tc.remoteAddr
		r.Header.Set("X-Forwarded-For", tc.xForwardedFor)
		assert.Equal(t, tc.expectedOutput, getRequestIP(r))
	}
}

func TestClientConfig(t *testing.T) {
	// Test the client configuration with the "noop" auth backend
	noopAuthConfig := &auth.NoopAuthConfig{}
	conf := &controllerConfig{
		noopAuthConfig: noopAuthConfig,
	}

	ctl, err := newController(conf)

	assert.NoError(t, err, "Couldn't create controller")

	assert.Equal(t, "",
		ctl.clientConfig.AccessManagementURL,
		"AccessManagementURL should be empty!")

	// Test the client configuration with the Github auth backend
	ghAuthConfig := &auth.GithubAuthConfig{}
	conf = &controllerConfig{
		githubAuthConfig: ghAuthConfig,
	}

	ctl, err = newController(conf)

	assert.NoError(t, err, "Couldn't create controller")

	assert.Equal(t, "https://github.com/settings/apps/authorizations",
		ctl.clientConfig.AccessManagementURL)

	// Test the client configuration with the Github Enterprise auth backend
	const phonyURL = "https://phony-ghe.url"

	ghAuthConfig = &auth.GithubAuthConfig{
		EnterpriseURL: phonyURL,
	}
	conf = &controllerConfig{
		githubAuthConfig: ghAuthConfig,
	}

	ctl, err = newController(conf)

	assert.NoError(t, err, "Couldn't create controller")

	assert.Equal(t, phonyURL+"/settings/apps/authorizations",
		ctl.clientConfig.AccessManagementURL)

	// Check that a version is set in the client configuration
	assert.NotEmpty(t, ctl.clientConfig.NebraskaVersion)
}

func TestOmahaRequestSizeLimitation(t *testing.T) {
	type testCase struct {
		bodyLength int
		status     int
	}

	a, err := api.NewForTest(api.OptionInitDB, api.OptionDisableUpdatesOnFailedRollout)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()

	for _, tc := range []testCase{
		{
			bodyLength: UpdateMaxRequestSize,
			status:     http.StatusOK,
		}, {
			bodyLength: UpdateMaxRequestSize + 1,
			status:     http.StatusBadRequest,
		},
	} {
		noopAuthConfig := &auth.NoopAuthConfig{}
		conf := &controllerConfig{
			noopAuthConfig: noopAuthConfig,
			api:            a,
		}
		ctl, err := newController(conf)
		require.NoError(t, err)
		gin.SetMode(gin.TestMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{}
		xmlRequest := bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>
	<request protocol="3.0" version="update_engine-0.4.10" updaterversion="update_engine-0.4.10" installsource="scheduler" ismachine="1">
		<os version="Chateau" platform="CoreOS" sp="2512.2.0_x86_64"></os>
		<app appid="e96281a6-d1af-4bde-9a0a-97b76e56dc57" version="1.2.3" track="stable" bootid="{965fb4c5-ad3e-4eb7-a4c2-ca0c0e31ec84}" oem="ami" oemversion="0.1.1-r1" alephversion="1688.5.3" machineid="INSERTINSTANCEIDHERE" lang="en-US" board="amd64-usr" hardware_class="" delta_okay="false" >
			<ping active="1"></ping>
			<updatecheck></updatecheck>
			<event eventtype="3" eventresult="1"></event>
		</app>`)
		xmlRequest.WriteString("<!--")
		tempBuffer := bytes.NewBufferString("--></request>")
		lenRequestXML := xmlRequest.Len()
		//insert (64kb - (the valid xml + the bytes for xml comments)) bytes into the large xml dummy request
		for i := 1; i <= tc.bodyLength-(lenRequestXML+tempBuffer.Len()); i++ {
			xmlRequest.WriteString("x")
		}
		xmlRequest.WriteString(tempBuffer.String())
		assert.Equal(t, tc.bodyLength, xmlRequest.Len())
		c.Request.Body = ioutil.NopCloser(xmlRequest)
		ctl.processOmahaRequest(c)
		assert.Equal(t, tc.status, w.Code)
	}
}
