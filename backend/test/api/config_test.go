package api_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		url := fmt.Sprintf("%s/config", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		method := "GET"
		// response
		rMap := make(map[string]interface{})

		httpDo(t, url, method, nil, http.StatusOK, "json", &rMap)

		assert.Equal(t, "noop", rMap["auth_mode"])
		assert.Equal(t, "", rMap["access_management_url"])
		assert.Equal(t, "", rMap["login_url"])
		// OIDC fields should be present but null/empty for noop auth mode
		assert.Nil(t, rMap["oidc_logout_url"])
		assert.Nil(t, rMap["oidc_client_id"])
		assert.Nil(t, rMap["oidc_issuer_url"])
		assert.Nil(t, rMap["oidc_scopes"])
	})
}
