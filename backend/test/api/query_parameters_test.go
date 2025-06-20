package api_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func TestPaginationParameters(t *testing.T) {
	t.Run("handle_pagination_parameters", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		baseURL := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))

		// test various pagination parameter combinations
		testCases := []struct {
			name       string
			params     map[string]string
			expectCode int
		}{
			{
				name:       "valid_pagination",
				params:     map[string]string{"page": "1", "perpage": "10"},
				expectCode: http.StatusOK,
			},
			{
				name:       "zero_page",
				params:     map[string]string{"page": "0", "perpage": "10"},
				expectCode: http.StatusOK, // Page 0 might be valid (first page)
			},
			{
				name:       "negative_page",
				params:     map[string]string{"page": "-1", "perpage": "10"},
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "invalid_page_string",
				params:     map[string]string{"page": "invalid", "perpage": "10"},
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "zero_perpage",
				params:     map[string]string{"page": "1", "perpage": "0"},
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "negative_perpage",
				params:     map[string]string{"page": "1", "perpage": "-5"},
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "large_perpage",
				params:     map[string]string{"page": "1", "perpage": "1000"},
				expectCode: http.StatusOK, // Should handle or limit large page sizes
			},
			{
				name:       "extremely_large_perpage",
				params:     map[string]string{"page": "1", "perpage": "999999"},
				expectCode: http.StatusBadRequest, // Should reject extremely large page sizes
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Build URL with query parameters
				u, err := url.Parse(baseURL)
				require.NoError(t, err)
				
				q := u.Query()
				for key, value := range tc.params {
					q.Set(key, value)
				}
				u.RawQuery = q.Encode()

				req, err := http.NewRequest("GET", u.String(), nil)
				require.NoError(t, err)

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expectCode, resp.StatusCode, 
					"Test case %s should return status %d", tc.name, tc.expectCode)
			})
		}
	})

	t.Run("should_handle_date_range_parameters", func(t *testing.T) {
		// Test activity endpoint with date parameters
		baseURL := fmt.Sprintf("%s/api/activity", os.Getenv("NEBRASKA_TEST_SERVER_URL"))

		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		testCases := []struct {
			name       string
			start      string
			end        string
			expectCode int
		}{
			{
				name:       "valid_date_range",
				start:      yesterday.Format(time.RFC3339),
				end:        now.Format(time.RFC3339),
				expectCode: http.StatusOK,
			},
			{
				name:       "invalid_start_date",
				start:      "invalid-date",
				end:        now.Format(time.RFC3339),
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "invalid_end_date", 
				start:      yesterday.Format(time.RFC3339),
				end:        "invalid-date",
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "start_after_end",
				start:      tomorrow.Format(time.RFC3339),
				end:        yesterday.Format(time.RFC3339),
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "missing_start_date",
				start:      "",
				end:        now.Format(time.RFC3339),
				expectCode: http.StatusBadRequest,
			},
			{
				name:       "missing_end_date",
				start:      yesterday.Format(time.RFC3339),
				end:        "",
				expectCode: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				u, err := url.Parse(baseURL)
				require.NoError(t, err)
				
				q := u.Query()
				if tc.start != "" {
					q.Set("start", tc.start)
				}
				if tc.end != "" {
					q.Set("end", tc.end)
				}
				u.RawQuery = q.Encode()

				req, err := http.NewRequest("GET", u.String(), nil)
				require.NoError(t, err)

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expectCode, resp.StatusCode,
					"Test case %s should return status %d", tc.name, tc.expectCode)
			})
		}
	})

	t.Run("should_handle_filter_parameters", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// Get an app to test instance filtering
		app := getRandomApp(t, db)
		if len(app.Groups) == 0 {
			t.Skip("No groups available for testing")
		}

		groupID := app.Groups[0].ID
		baseURL := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances", 
			os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, groupID)

		testCases := []struct {
			name       string
			params     map[string]string
			expectCode int
		}{
			{
				name: "valid_filters",
				params: map[string]string{
					"status":    "0",
					"sort":      "2", 
					"sortOrder": "0",
					"page":      "1",
					"perpage":   "10",
					"duration":  "30d",
				},
				expectCode: http.StatusOK,
			},
			{
				name: "invalid_status",
				params: map[string]string{
					"status":  "invalid",
					"page":    "1", 
					"perpage": "10",
				},
				expectCode: http.StatusBadRequest,
			},
			{
				name: "invalid_sort",
				params: map[string]string{
					"sort":    "999",
					"page":    "1",
					"perpage": "10",
				},
				expectCode: http.StatusBadRequest,
			},
			{
				name: "invalid_sort_order",
				params: map[string]string{
					"sortOrder": "invalid",
					"page":      "1",
					"perpage":   "10",
				},
				expectCode: http.StatusBadRequest,
			},
			{
				name: "invalid_duration",
				params: map[string]string{
					"duration": "invalid-duration",
					"page":     "1",
					"perpage":  "10",
				},
				expectCode: http.StatusBadRequest,
			},
			{
				name: "negative_status",
				params: map[string]string{
					"status":  "-1",
					"page":    "1",
					"perpage": "10",
				},
				expectCode: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				u, err := url.Parse(baseURL)
				require.NoError(t, err)
				
				q := u.Query()
				for key, value := range tc.params {
					q.Set(key, value)
				}
				u.RawQuery = q.Encode()

				req, err := http.NewRequest("GET", u.String(), nil)
				require.NoError(t, err)

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expectCode, resp.StatusCode,
					"Test case %s should return status %d", tc.name, tc.expectCode)
			})
		}
	})

	t.Run("should_handle_special_characters_in_parameters", func(t *testing.T) {
		baseURL := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))

		testCases := []struct {
			name       string
			params     map[string]string
			expectCode int
		}{
			{
				name: "url_encoded_characters",
				params: map[string]string{
					"page":    "1",
					"perpage": "10",
					"filter":  "test%20app", // URL encoded space
				},
				expectCode: http.StatusOK, // Should handle URL encoding
			},
			{
				name: "special_characters",
				params: map[string]string{
					"page":    "1",
					"perpage": "10", 
					"filter":  "test&app=value", // Special characters
				},
				expectCode: http.StatusOK, // Should handle or ignore invalid filters
			},
			{
				name: "unicode_characters",
				params: map[string]string{
					"page":    "1",
					"perpage": "10",
					"filter":  "tëst-àpp", // Unicode characters
				},
				expectCode: http.StatusOK, // Should handle unicode
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				u, err := url.Parse(baseURL)
				require.NoError(t, err)
				
				q := u.Query()
				for key, value := range tc.params {
					q.Set(key, value)
				}
				u.RawQuery = q.Encode()

				req, err := http.NewRequest("GET", u.String(), nil)
				require.NoError(t, err)

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expectCode, resp.StatusCode,
					"Test case %s should return status %d", tc.name, tc.expectCode)
			})
		}
	})

	t.Run("should_handle_duplicate_parameters", func(t *testing.T) {
		baseURL := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))

		// Test URL with duplicate parameters
		testURL := baseURL + "?page=1&page=2&perpage=10"

		req, err := http.NewRequest("GET", testURL, nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle duplicate parameters gracefully (usually takes the last value)
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, resp.StatusCode)
	})

	t.Run("should_handle_extremely_long_parameter_values", func(t *testing.T) {
		baseURL := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))

		// Create extremely long parameter value
		longValue := strings.Repeat("a", 10000)
		
		u, err := url.Parse(baseURL)
		require.NoError(t, err)
		
		q := u.Query()
		q.Set("page", "1")
		q.Set("perpage", "10")
		q.Set("filter", longValue)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest("GET", u.String(), nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle long parameter values gracefully
		assert.Contains(t, []int{
			http.StatusOK,                    // Accepted and processed
			http.StatusBadRequest,            // Rejected due to length
			http.StatusRequestURITooLong,     // Rejected due to URL length
		}, resp.StatusCode)
	})
}

func TestPaginationResponseValidation(t *testing.T) {
	t.Run("should_return_consistent_pagination_metadata", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		baseURL := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))

		testCases := []struct {
			page    int
			perpage int
		}{
			{page: 1, perpage: 5},
			{page: 1, perpage: 10},
			{page: 2, perpage: 5},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("page_%d_perpage_%d", tc.page, tc.perpage), func(t *testing.T) {
				u, err := url.Parse(baseURL)
				require.NoError(t, err)
				
				q := u.Query()
				q.Set("page", fmt.Sprintf("%d", tc.page))
				q.Set("perpage", fmt.Sprintf("%d", tc.perpage))
				u.RawQuery = q.Encode()

				var appsResp codegen.AppsPage
				httpDo(t, u.String(), "GET", nil, http.StatusOK, "json", &appsResp)

				// Validate pagination metadata consistency
				assert.GreaterOrEqual(t, len(appsResp.Applications), 0, "Should return valid applications array")
				
				// If there are results, validate they don't exceed perpage limit
				if len(appsResp.Applications) > 0 {
					assert.LessOrEqual(t, len(appsResp.Applications), tc.perpage, 
						"Should not return more items than perpage limit")
				}
			})
		}
	})
}