package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kinvolk/nebraska/cmd/nebraska/auth"
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
