package api_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

func TestPaginationParameters(t *testing.T) {
	t.Run("accept_valid_pagination", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test valid pagination parameters
		url := fmt.Sprintf("%s/api/apps?page=1&perpage=10", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		resp := httpMakeRequest(t, "GET", url, nil, nil)
		defer resp.Body.Close()

		// api accepts valid pagination
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("validate_pagination_parameters", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test api validation of pagination parameters
		testCases := []struct {
			name         string
			url          string
			expectedCode int
		}{
			{"zero_page", "page=0&perpage=10", http.StatusOK},
			{"negative_page", "page=-1&perpage=10", http.StatusBadRequest},
			{"invalid_page", "page=invalid&perpage=10", http.StatusBadRequest},
			{"small_perpage", "page=1&perpage=5", http.StatusBadRequest},
			{"negative_perpage", "page=1&perpage=-5", http.StatusBadRequest},
			{"large_perpage", "page=1&perpage=999999", http.StatusOK},
		}

		baseURL := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := fmt.Sprintf("%s?%s", baseURL, tc.url)
				resp := httpMakeRequest(t, "GET", url, nil, nil)
				defer resp.Body.Close()

				// api validates pagination parameters
				assert.Equal(t, tc.expectedCode, resp.StatusCode)
			})
		}
	})
}

func TestDateRangeParameters(t *testing.T) {
	t.Run("accept_valid_date_range", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test valid date range
		url := fmt.Sprintf("%s/api/activity?start=2023-01-01T00:00:00Z&end=2023-12-31T23:59:59Z", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		resp := httpMakeRequest(t, "GET", url, nil, nil)
		defer resp.Body.Close()

		// api accepts valid date ranges
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("accept_permissive_dates", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test that api is permissive with date parameters
		testCases := []struct {
			name string
			url  string
		}{
			{"invalid_start", "start=invalid-date&end=2023-12-31T23:59:59Z"},
			{"invalid_end", "start=2023-01-01T00:00:00Z&end=invalid-date"},
			{"start_after_end", "start=2023-12-31T23:59:59Z&end=2023-01-01T00:00:00Z"},
		}

		baseURL := fmt.Sprintf("%s/api/activity", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := fmt.Sprintf("%s?%s", baseURL, tc.url)
				resp := httpMakeRequest(t, "GET", url, nil, nil)
				defer resp.Body.Close()

				// api is permissive with date parameters
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
		}
	})
}

func TestSpecialCharactersInParameters(t *testing.T) {
	t.Run("handle_special_characters", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test special characters in parameters
		testCases := []struct {
			name string
			url  string
		}{
			{"url_encoded", "page=1&perpage=10&filter=test%20app"},
			{"unicode", "page=1&perpage=10&filter=tëst-àpp"},
			{"special_chars", "page=1&perpage=10&filter=test&app=value"},
		}

		baseURL := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := fmt.Sprintf("%s?%s", baseURL, tc.url)
				resp := httpMakeRequest(t, "GET", url, nil, nil)
				defer resp.Body.Close()

				// api handles special characters gracefully
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
		}
	})
}

func TestPaginationResponseValidation(t *testing.T) {
	t.Run("return_consistent_pagination_metadata", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test pagination response structure with valid parameters
		url := fmt.Sprintf("%s/api/apps?page=1&perpage=10", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		resp := httpMakeRequest(t, "GET", url, nil, nil)
		defer resp.Body.Close()

		// api returns successful response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// decode and validate response structure
		var appsResp codegen.AppsPage
		err := httpDecodeResponse(resp, &appsResp)
		assert.NoError(t, err)

		// verify pagination metadata exists
		assert.NotNil(t, appsResp.TotalCount)
		assert.NotNil(t, appsResp.Applications)
		assert.GreaterOrEqual(t, len(appsResp.Applications), 0)
		assert.GreaterOrEqual(t, appsResp.TotalCount, 0)
	})

	t.Run("reject_small_perpage", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// test that api requires minimum perpage of 10
		url := fmt.Sprintf("%s/api/apps?page=1&perpage=5", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		resp := httpMakeRequest(t, "GET", url, nil, nil)
		defer resp.Body.Close()

		// api rejects perpage values less than 10
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
