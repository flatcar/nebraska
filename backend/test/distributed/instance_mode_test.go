package distributed_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeRefusesAdminOperations(t *testing.T) {
	base := edgeURL()

	// A sample: one middleware guards every admin op; pkg/middleware checks the spec.
	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{"createApp", "POST", "/api/apps"},
		{"updateApp", "PUT", "/api/apps/" + seededAppID},
		{"deleteApp", "DELETE", "/api/apps/" + seededAppID},
		{"createGroup", "POST", fmt.Sprintf("/api/apps/%s/groups", seededAppID)},
		{"deleteGroup", "DELETE", fmt.Sprintf("/api/apps/%s/groups/%s", seededAppID, seededGroupID)},
		{"deletePackage", "DELETE", fmt.Sprintf("/api/apps/%s/packages/%s", seededAppID, unknownID)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The guard runs before any lookup, so unknown IDs are refused too.
			code, body := request(t, tc.method, base+tc.path, nil)

			assert.Equal(t, http.StatusForbidden, code)
			assert.Contains(t, body, "admin_operation_not_accepted")
			assert.Contains(t, body, "control node")
		})
	}
}

func TestAdminOperationsStayOpenElsewhere(t *testing.T) {
	for _, tc := range []struct {
		name string
		url  func() string
	}{
		{"control", controlURL},
		{"single", singleURL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := tc.url()
			// Every segment of a product ID has to start with a letter.
			suffix := fmt.Sprintf("n%d", time.Now().UnixNano())
			productID := fmt.Sprintf("io.test.%s.%s", tc.name, suffix)
			payload := fmt.Sprintf(`{"name":%q,"description":"instance mode test","product_id":%q}`,
				tc.name+"-"+suffix, productID)

			code, body := request(t, "POST", base+"/api/apps", strings.NewReader(payload))

			require.Equal(t, http.StatusOK, code, body)
			assert.Contains(t, body, productID)

			t.Cleanup(func() {
				code, body := request(t, "DELETE", base+"/api/apps/"+productID, nil)
				assert.Equal(t, http.StatusNoContent, code, body)
			})
		})
	}
}

func TestEdgeStillServesReads(t *testing.T) {
	base := edgeURL()

	for _, tc := range []struct {
		path string
		want string
	}{
		{"/health", "OK"},
		{"/api/apps", seededAppID},
		{"/api/apps/" + seededAppID, seededAppID},
	} {
		t.Run(strings.ReplaceAll(strings.TrimPrefix(tc.path, "/"), "/", "_"), func(t *testing.T) {
			code, body := request(t, "GET", base+tc.path, nil)
			assert.Equal(t, http.StatusOK, code, body)
			assert.Contains(t, body, tc.want)
		})
	}
}

func TestClearGroupUpdatesOverride(t *testing.T) {
	group := fmt.Sprintf("/api/apps/%s/groups/%s", seededAppID, seededGroupID)
	override := group + "/updates_override"

	for _, tc := range []struct {
		name  string
		url   func() string
		dbURL func() string
	}{
		{"edge", edgeURL, edgeDBURL},
		{"control", controlURL, controlDBURL},
	} {
		t.Run(tc.name+"_clears_its_own_override", func(t *testing.T) {
			base, dbURL := tc.url(), tc.dbURL()

			setUpdatesOverride(t, dbURL, seededGroupID, false)
			t.Cleanup(func() { setUpdatesOverride(t, dbURL, seededGroupID, nil) })

			code, before := request(t, "GET", base+group, nil)
			require.Equal(t, http.StatusOK, code, before)
			require.Contains(t, before, `"policy_updates_enabled":false`)

			code, body := request(t, "DELETE", base+override, nil)
			require.Equal(t, http.StatusNoContent, code, body)
			assert.Empty(t, body)

			// Clearing the override has to restore the admin default rather
			// than write a value of its own.
			code, after := request(t, "GET", base+group, nil)
			require.Equal(t, http.StatusOK, code, after)
			assert.Contains(t, after, `"policy_updates_enabled":true`)
		})
	}

	t.Run("unknown_group_is_not_found", func(t *testing.T) {
		base := edgeURL()

		code, _ := request(t, "DELETE", base+fmt.Sprintf("/api/apps/%s/groups/%s/updates_override", seededAppID, unknownID), nil)
		assert.Equal(t, http.StatusNotFound, code)
	})

	t.Run("single_has_no_local_override", func(t *testing.T) {
		code, body := request(t, "DELETE", singleURL()+override, nil)

		assert.Equal(t, http.StatusNotImplemented, code)
		assert.Contains(t, body, "local_override_not_supported")
	})
}
