package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		url := fmt.Sprintf("%s/config", testServerURL)
		method := "GET"
		// response
		rMap := make(map[string]interface{})

		httpDo(t, url, method, nil, http.StatusOK, "json", &rMap)

		assert.Equal(t, "noop", rMap["auth_mode"])
		assert.Equal(t, "", rMap["access_management_url"])
		assert.Equal(t, "", rMap["login_url"])
		assert.Equal(t, "", rMap["logout_url"])
	})
}
