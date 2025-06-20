package api_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeValidation(t *testing.T) {
	t.Run("reject_wrong_content_type_POST", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with wrong content-type
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		appName := "content_type_test"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		// make request with text/plain content-type instead of application/json
		headers := map[string]string{"Content-Type": "text/plain"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()

		// should fail with 400 bad request or 415 unsupported media type
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnsupportedMediaType}, resp.StatusCode)
	})

	t.Run("reject_wrong_content_type_PUT", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get random app from db to update
		app := getRandomApp(t, db)

		// update app request with wrong content-type
		url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
		name := "content_type_updated"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","description":"%s","id":"%s"}`, name, app.Description, app.ID))

		// make request with application/xml content-type instead of application/json
		headers := map[string]string{"Content-Type": "application/xml"}
		resp := httpMakeRequest(t, "PUT", url, payload, headers)
		defer resp.Body.Close()

		// should fail with 400 bad request or 415 unsupported media type
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnsupportedMediaType}, resp.StatusCode)
	})

	t.Run("accept_correct_content_type", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with correct content-type
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		appName := "content_type_valid"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		// use the standard httpDo helper with correct content-type
		var response interface{}
		httpDo(t, url, "POST", payload, http.StatusOK, "json", &response)
	})
}

func TestMalformedRequestHandling(t *testing.T) {
	t.Run("reject_invalid_json", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with invalid json
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		invalidJSON := `{"name":"test_app","description":}`
		payload := strings.NewReader(invalidJSON)

		headers := map[string]string{"Content-Type": "application/json"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()

		// should fail with 400 bad request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("reject_incomplete_json", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with incomplete json (missing closing brace)
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		incompleteJSON := `{"name":"test_app","description":"test"`
		payload := strings.NewReader(incompleteJSON)

		headers := map[string]string{"Content-Type": "application/json"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()

		// should fail with 400 bad request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("reject_non_json_with_json_header", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with non-json content but json content-type
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		nonJSON := `name=test_app&description=test`
		payload := strings.NewReader(nonJSON)

		headers := map[string]string{"Content-Type": "application/json"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()

		// should fail with 400 bad request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestLargePayloadHandling(t *testing.T) {
	t.Run("handle_large_but_valid_payload", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with large description (10KB)
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		appName := "large_payload_test"
		largeDescription := strings.Repeat("A", 10*1024)
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","description":"%s"}`, appName, largeDescription))

		headers := map[string]string{"Content-Type": "application/json"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()

		// should handle large payloads appropriately - either accept or reject based on size limits
		// should not fail with internal server errors
		assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, []int{
			http.StatusOK,                    // accepted
			http.StatusCreated,               // created
			http.StatusBadRequest,            // rejected due to validation
			http.StatusRequestEntityTooLarge, // rejected due to size
			http.StatusUnprocessableEntity,   // rejected due to business logic
		}, resp.StatusCode)
	})

	t.Run("reject_extremely_large_payload", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with extremely large payload (1MB)
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		appName := "huge_payload_test"
		hugeDescription := strings.Repeat("A", 1024*1024)
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","description":"%s"}`, appName, hugeDescription))

		headers := map[string]string{"Content-Type": "application/json"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()

		// should reject extremely large payloads
		assert.Contains(t, []int{
			http.StatusBadRequest,            // rejected due to validation
			http.StatusRequestEntityTooLarge, // rejected due to size
			http.StatusUnprocessableEntity,   // rejected due to business logic
		}, resp.StatusCode)
	})
}

func TestHTTPMethodValidation(t *testing.T) {
	t.Run("handle_HEAD_requests", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test HEAD request to apps endpoint
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		resp := httpMakeRequest(t, "HEAD", url, nil, nil)
		defer resp.Body.Close()

		// HEAD should be handled appropriately (200 or 405 method not allowed)
		assert.Contains(t, []int{http.StatusOK, http.StatusMethodNotAllowed}, resp.StatusCode)
	})

	t.Run("handle_OPTIONS_requests", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test OPTIONS request for CORS preflight
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		headers := map[string]string{
			"Origin":                         "http://localhost:3000",
			"Access-Control-Request-Method":  "POST",
			"Access-Control-Request-Headers": "Content-Type",
		}
		resp := httpMakeRequest(t, "OPTIONS", url, nil, headers)
		defer resp.Body.Close()

		// OPTIONS should be handled appropriately (200, 204, or 405)
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent, http.StatusMethodNotAllowed}, resp.StatusCode)
	})

	t.Run("reject_unsupported_methods", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test unsupported method
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		payload := strings.NewReader(`{"name":"test"}`)
		headers := map[string]string{"Content-Type": "application/json"}
		resp := httpMakeRequest(t, "PATCH", url, payload, headers)
		defer resp.Body.Close()

		// should reject unsupported methods
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

func TestHTTPHeaderValidation(t *testing.T) {
	t.Run("handle_missing_content_type", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request without content-type header
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		appName := "no_content_type"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		// intentionally not setting content-type
		resp := httpMakeRequest(t, "POST", url, payload, nil)
		defer resp.Body.Close()

		// should handle missing content-type appropriately
		assert.Contains(t, []int{
			http.StatusBadRequest,            // rejected due to missing header
			http.StatusUnsupportedMediaType,  // rejected due to missing/wrong content type
			http.StatusUnprocessableEntity,   // rejected due to processing issues
		}, resp.StatusCode)
	})

	t.Run("handle_charset_in_content_type", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create app request with charset in content-type
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		appName := "charset_test"
		payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

		headers := map[string]string{"Content-Type": "application/json; charset=utf-8"}
		resp := httpMakeRequest(t, "POST", url, payload, headers)
		defer resp.Body.Close()

		// should accept content-type with charset
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}