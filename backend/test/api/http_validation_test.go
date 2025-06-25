package api_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentTypeValidation(t *testing.T) {
	t.Run("content_type_validation", func(t *testing.T) {
		db := newDBForTest(t)
		defer db.Close()

		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		payload := strings.NewReader(`{"name":"test_app"}`)

		// Should reject wrong content-type
		headers := map[string]string{"Content-Type": "text/plain"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnsupportedMediaType}, resp.StatusCode)

		// Should accept correct content-type
		headers = map[string]string{"Content-Type": "application/json"}
		resp = httpMakeRequest(t, "POST", url, strings.NewReader(`{"name":"valid_test"}`), headers)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Should accept content-type with charset
		headers = map[string]string{"Content-Type": "application/json; charset=utf-8"}
		resp = httpMakeRequest(t, "POST", url, strings.NewReader(`{"name":"charset_test"}`), headers)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestMalformedRequestHandling(t *testing.T) {
	t.Run("reject_malformed_json", func(t *testing.T) {
		db := newDBForTest(t)
		defer db.Close()

		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		headers := map[string]string{"Content-Type": "application/json"}

		// Test various malformed JSON cases
		malformedCases := []string{
			`{"name":"test_app","description":}`,      // Invalid JSON syntax
			`{"name":"test_app","description":"test"`, // Incomplete JSON
			`name=test_app&description=test`,          // Non-JSON content
		}

		for _, malformedJSON := range malformedCases {
			resp := httpMakeRequest(t, "POST", url, strings.NewReader(malformedJSON), headers)
			resp.Body.Close()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		}
	})
}

func TestHTTPRequestValidation(t *testing.T) {
	t.Run("basic_http_request_handling", func(t *testing.T) {
		db := newDBForTest(t)
		defer db.Close()

		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))

		// Test GET requests work
		resp := httpMakeRequest(t, "GET", url, nil, nil)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Test POST with missing content-type is rejected
		resp = httpMakeRequest(t, "POST", url, strings.NewReader(`{"name":"test"}`), nil)
		resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Test POST with correct content-type works
		headers := map[string]string{"Content-Type": "application/json"}
		resp = httpMakeRequest(t, "POST", url, strings.NewReader(`{"name":"test_post"}`), headers)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
