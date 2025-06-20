package api_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

func TestConcurrentReadRequests(t *testing.T) {
	t.Run("handle_concurrent_read_requests", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		
		const numRequests = 10
		var wg sync.WaitGroup
		results := make([]int, numRequests)
		errors := make([]error, numRequests)

		// make concurrent GET requests
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					errors[index] = err
					return
				}

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					errors[index] = err
					return
				}
				defer resp.Body.Close()

				results[index] = resp.StatusCode
			}(i)
		}

		wg.Wait()

		// all requests should succeed
		for i := 0; i < numRequests; i++ {
			assert.NoError(t, errors[i], "request %d should not have errors", i)
			assert.Equal(t, http.StatusOK, results[i], "request %d should return 200", i)
		}
	})
}

func TestConcurrentWriteRequests(t *testing.T) {
	t.Run("handle_concurrent_write_requests", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		
		const numRequests = 5
		var wg sync.WaitGroup
		results := make([]int, numRequests)
		errors := make([]error, numRequests)
		createdApps := make([]*api.Application, numRequests)

		// make concurrent POST requests to create apps
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				appName := fmt.Sprintf("concurrent_app_%d_%d", index, time.Now().UnixNano())
				payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

				req, err := http.NewRequest("POST", url, payload)
				if err != nil {
					errors[index] = err
					return
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					errors[index] = err
					return
				}
				defer resp.Body.Close()

				results[index] = resp.StatusCode

				// if successful, decode the response
				if resp.StatusCode == http.StatusOK {
					var app api.Application
					err := httpDecodeResponse(resp, &app)
					if err == nil {
						createdApps[index] = &app
					}
				}
			}(i)
		}

		wg.Wait()

		// check results
		successCount := 0
		for i := 0; i < numRequests; i++ {
			assert.NoError(t, errors[i], "request %d should not have errors", i)
			
			// all requests should either succeed or fail gracefully (no server errors)
			assert.NotEqual(t, http.StatusInternalServerError, results[i], 
				"request %d should not return 500", i)
			
			if results[i] == http.StatusOK {
				successCount++
				assert.NotNil(t, createdApps[i], "successful request %d should have created an app", i)
			}
		}

		// at least some requests should succeed
		assert.Greater(t, successCount, 0, "at least some concurrent requests should succeed")
	})
}

func TestConcurrentMixedRequests(t *testing.T) {
	t.Run("handle_concurrent_mixed_requests", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		const numReads = 5
		const numWrites = 3
		const totalRequests = numReads + numWrites
		
		var wg sync.WaitGroup
		results := make([]int, totalRequests)
		errors := make([]error, totalRequests)

		// start concurrent read requests
		for i := 0; i < numReads; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					errors[index] = err
					return
				}

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					errors[index] = err
					return
				}
				defer resp.Body.Close()

				results[index] = resp.StatusCode
			}(i)
		}

		// start concurrent write requests
		for i := 0; i < numWrites; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				realIndex := numReads + index
				appName := fmt.Sprintf("mixed_concurrent_app_%d_%d", index, time.Now().UnixNano())
				payload := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, appName))

				url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
				req, err := http.NewRequest("POST", url, payload)
				if err != nil {
					errors[realIndex] = err
					return
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					errors[realIndex] = err
					return
				}
				defer resp.Body.Close()

				results[realIndex] = resp.StatusCode
			}(i)
		}

		wg.Wait()

		// all requests should complete without errors
		for i := 0; i < totalRequests; i++ {
			assert.NoError(t, errors[i], "request %d should not have errors", i)
			assert.NotEqual(t, http.StatusInternalServerError, results[i], 
				"request %d should not return 500", i)
		}

		// all read requests should succeed
		for i := 0; i < numReads; i++ {
			assert.Equal(t, http.StatusOK, results[i], "read request %d should succeed", i)
		}
	})
}

func TestConcurrentSameResourceUpdates(t *testing.T) {
	t.Run("handle_concurrent_updates_to_same_resource", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// get an existing app to update concurrently
		app := getRandomApp(t, db)

		const numRequests = 5
		var wg sync.WaitGroup
		results := make([]int, numRequests)
		errors := make([]error, numRequests)

		// make concurrent PUT requests to the same app
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
				name := fmt.Sprintf("updated_concurrent_%d_%d", index, time.Now().UnixNano())
				payload := strings.NewReader(fmt.Sprintf(`{"name":"%s","description":"%s","id":"%s"}`, 
					name, app.Description, app.ID))

				req, err := http.NewRequest("PUT", url, payload)
				if err != nil {
					errors[index] = err
					return
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					errors[index] = err
					return
				}
				defer resp.Body.Close()

				results[index] = resp.StatusCode
			}(i)
		}

		wg.Wait()

		// check that all requests completed
		successCount := 0
		for i := 0; i < numRequests; i++ {
			assert.NoError(t, errors[i], "request %d should not have errors", i)
			
			// should handle concurrent updates gracefully
			assert.Contains(t, []int{
				http.StatusOK,                  // successful update
				http.StatusConflict,            // conflict due to concurrent modification
				http.StatusUnprocessableEntity, // business logic error
			}, results[i], "request %d should return appropriate status", i)

			if results[i] == http.StatusOK {
				successCount++
			}
		}

		// at least one update should succeed
		assert.Greater(t, successCount, 0, "at least one concurrent update should succeed")
	})
}

func TestConnectionPooling(t *testing.T) {
	t.Run("not_exhaust_connection_pool", func(t *testing.T) {
		// this test ensures the database connection pool can handle many concurrent requests
		// without exhausting connections
		
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()
		
		url := fmt.Sprintf("%s/api/apps", os.Getenv("NEBRASKA_TEST_SERVER_URL"))
		
		const numRequests = 20
		var wg sync.WaitGroup
		results := make([]int, numRequests)
		errors := make([]error, numRequests)
		durations := make([]time.Duration, numRequests)

		// make many concurrent requests
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				start := time.Now()
				
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					errors[index] = err
					return
				}

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					errors[index] = err
					return
				}
				defer resp.Body.Close()

				results[index] = resp.StatusCode
				durations[index] = time.Since(start)
			}(i)
		}

		wg.Wait()

		// all requests should complete successfully
		for i := 0; i < numRequests; i++ {
			assert.NoError(t, errors[i], "request %d should not have errors", i)
			assert.Equal(t, http.StatusOK, results[i], "request %d should succeed", i)
			
			// requests should complete in reasonable time (not timing out due to pool exhaustion)
			assert.Less(t, durations[i], 10*time.Second, "request %d should complete quickly", i)
		}
	})
}

func TestRaceConditionHandling(t *testing.T) {
	t.Run("handle_create_and_delete_race", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)
		defer db.Close()

		// create an app first
		app := getRandomApp(t, db)

		var wg sync.WaitGroup
		deleteResult := 0
		readResults := make([]int, 3)
		deleteError := error(nil)
		readErrors := make([]error, 3)

		// start delete request
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
			req, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				deleteError = err
				return
			}

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				deleteError = err
				return
			}
			defer resp.Body.Close()

			deleteResult = resp.StatusCode
		}()

		// start concurrent read requests
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				// add small delay to ensure some requests happen after delete starts
				time.Sleep(time.Duration(index*10) * time.Millisecond)
				
				url := fmt.Sprintf("%s/api/apps/%s", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID)
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					readErrors[index] = err
					return
				}

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					readErrors[index] = err
					return
				}
				defer resp.Body.Close()

				readResults[index] = resp.StatusCode
			}(i)
		}

		wg.Wait()

		// delete should succeed
		assert.NoError(t, deleteError)
		assert.Equal(t, http.StatusNoContent, deleteResult)

		// read requests should either succeed (if they executed before delete) 
		// or return 404 (if they executed after delete)
		for i := 0; i < 3; i++ {
			assert.NoError(t, readErrors[i], "read request %d should not have errors", i)
			assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, readResults[i],
				"read request %d should return 200 or 404", i)
		}
	})
}