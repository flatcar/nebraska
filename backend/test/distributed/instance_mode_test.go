package distributed_test

import (
	"encoding/json"
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

// TestSavingAGroupKeepsTheLocalBrakeLocal saves a group the way the edit dialog
// does, sending back every field it read, on a node whose brake has tripped.
func TestSavingAGroupKeepsTheLocalBrakeLocal(t *testing.T) {
	for _, tc := range []struct {
		name         string
		url, dbURL   func() string
		showsBrake   bool
		keepsDefault bool
	}{
		// The control node's default replicates to every edge.
		{"control", controlURL, controlDBURL, false, true},
		// A single node has one switch, so saving it as shown keeps it paused.
		{"single", singleURL, singleDBURL, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base, dbURL := tc.url(), tc.dbURL()
			groupID := createGroup(t, base)
			group := fmt.Sprintf("%s/api/apps/%s/groups/%s", base, seededAppID, groupID)

			setUpdatesOverride(t, dbURL, groupID, false)

			code, body := request(t, "GET", group, nil)
			require.Equal(t, http.StatusOK, code, body)
			assert.Equal(t, !tc.showsBrake, strings.Contains(body, `"policy_updates_enabled":true`), body)
			assert.Equal(t, !tc.showsBrake, strings.Contains(body, `"local_overrides"`), body)

			code, body = request(t, "PUT", group, strings.NewReader(editDialogSave(t, body)))
			require.Equal(t, http.StatusOK, code, body)

			assert.Equal(t, tc.keepsDefault, adminUpdatesEnabled(t, dbURL, groupID))
		})
	}

	// UpdateGroup is the one admin endpoint that also writes runtime-owned state,
	// which it reaches through LocalUpdatesBrake.
	t.Run("single_update_group_clears_the_override", func(t *testing.T) {
		base, dbURL := singleURL(), singleDBURL()
		group := fmt.Sprintf("/api/apps/%s/groups/%s", seededAppID, seededGroupID)

		setUpdatesOverride(t, dbURL, seededGroupID, false)
		t.Cleanup(func() { setUpdatesOverride(t, dbURL, seededGroupID, nil) })

		code, before := request(t, "GET", base+group, nil)
		require.Equal(t, http.StatusOK, code, before)
		require.Contains(t, before, `"policy_updates_enabled":false`)

		// updateGroup replaces the whole row, so the body has to carry every
		// field back or the omitted ones are written as their zero value.
		var cfg map[string]any
		require.NoError(t, json.Unmarshal([]byte(before), &cfg))
		cfg["policy_updates_enabled"] = true

		payload, err := json.Marshal(cfg)
		require.NoError(t, err)

		code, body := request(t, "PUT", base+group, strings.NewReader(string(payload)))
		require.Equal(t, http.StatusOK, code, body)

		// The group default is already true, so this can only read true if the
		// node-local override was cleared.
		code, after := request(t, "GET", base+group, nil)
		require.Equal(t, http.StatusOK, code, after)
		assert.Contains(t, after, `"policy_updates_enabled":true`)
	})
}

func TestConfigReportsTheInstanceMode(t *testing.T) {
	for _, tc := range []struct {
		mode string
		url  func() string
	}{
		{"single", singleURL},
		{"control", controlURL},
		{"edge", edgeURL},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			code, body := request(t, "GET", tc.url()+"/config", nil)
			require.Equal(t, http.StatusOK, code, body)

			var config struct {
				InstanceMode string `json:"instance_mode"`
			}
			require.NoError(t, json.Unmarshal([]byte(body), &config))
			assert.Equal(t, tc.mode, config.InstanceMode)
		})
	}
}

func TestUpdateGroupLocalOverrides(t *testing.T) {
	group := fmt.Sprintf("/api/apps/%s/groups/%s", seededAppID, seededGroupID)
	overrides := group + "/local_overrides"

	type presented struct {
		PolicyUpdatesEnabled      bool           `json:"policy_updates_enabled"`
		PolicySafeMode            bool           `json:"policy_safe_mode"`
		PolicyMaxUpdatesPerPeriod int            `json:"policy_max_updates_per_period"`
		LocalOverrides            map[string]any `json:"local_overrides"`
	}
	decode := func(t *testing.T, body string) presented {
		t.Helper()

		var group presented
		require.NoError(t, json.Unmarshal([]byte(body), &group))
		return group
	}

	for _, tc := range []struct {
		name          string
		url, dbURL    func() string
		other, oDBURL func() string
	}{
		{"edge", edgeURL, edgeDBURL, controlURL, controlDBURL},
		{"control", controlURL, controlDBURL, edgeURL, edgeDBURL},
	} {
		t.Run(tc.name+"_overrides_its_own_policy", func(t *testing.T) {
			base, dbURL := tc.url(), tc.dbURL()

			// Both nodes start without overrides, whatever an earlier run left.
			clearOverrides(t, dbURL, seededGroupID)
			clearOverrides(t, tc.oDBURL(), seededGroupID)
			// A brake the rollout engine tripped is not part of the change.
			setUpdatesOverride(t, dbURL, seededGroupID, false)
			t.Cleanup(func() { clearOverrides(t, dbURL, seededGroupID) })
			before := readGroupDefault(t, dbURL, seededGroupID)

			code, body := request(t, "PATCH", base+overrides,
				strings.NewReader(`{"policy_safe_mode":true,"policy_max_updates_per_period":3,"policy_period_interval":"2 hours"}`))
			require.Equal(t, http.StatusOK, code, body)

			got := decode(t, body)
			assert.Equal(t, map[string]any{
				"policy_updates_enabled": false, "policy_safe_mode": true, "policy_office_hours": nil, "policy_timezone": nil,
				"policy_period_interval": "2 hours", "policy_max_updates_per_period": float64(3), "policy_update_timeout": nil,
			}, got.LocalOverrides)
			assert.Equal(t, before.MaxUpdatesPerPeriod, got.PolicyMaxUpdatesPerPeriod, "the policy fields show the default")
			assert.Equal(t, before.UpdatesEnabled, got.PolicyUpdatesEnabled, "the default, not the brake")
			assert.Equal(t, before, readGroupDefault(t, dbURL, seededGroupID), "the default must not change")

			// The override stays on the node that set it.
			code, body = request(t, "GET", tc.other()+group, nil)
			require.Equal(t, http.StatusOK, code, body)
			assert.Nil(t, decode(t, body).LocalOverrides["policy_safe_mode"])
			assert.Equal(t, before, readGroupDefault(t, tc.oDBURL(), seededGroupID))

			code, body = request(t, "PATCH", base+overrides, strings.NewReader(`{"policy_safe_mode":null,"policy_max_updates_per_period":null,"policy_period_interval":null}`))
			require.Equal(t, http.StatusOK, code, body)
			assert.Equal(t, map[string]any{
				"policy_updates_enabled": false, "policy_safe_mode": nil, "policy_office_hours": nil, "policy_timezone": nil,
				"policy_period_interval": nil, "policy_max_updates_per_period": nil, "policy_update_timeout": nil,
			}, decode(t, body).LocalOverrides, "removing overrides leaves the brake")

			// The node releases its brake with a null override.
			code, body = request(t, "PATCH", base+overrides, strings.NewReader(`{"policy_updates_enabled":null}`))
			require.Equal(t, http.StatusOK, code, body)
			assert.Nil(t, decode(t, body).LocalOverrides["policy_updates_enabled"])
			assert.Equal(t, before, readGroupDefault(t, dbURL, seededGroupID), "releasing the brake must not change the default")
		})
	}

	t.Run("single_has_no_local_overrides", func(t *testing.T) {
		code, body := request(t, "PATCH", singleURL()+overrides, strings.NewReader(`{"policy_safe_mode":true}`))

		assert.Equal(t, http.StatusNotImplemented, code)
		assert.Contains(t, body, "local_override_not_supported")
	})

	t.Run("refuses_what_a_node_cannot_apply", func(t *testing.T) {
		for _, payload := range []string{
			// The default is the pause every node honors.
			`{"policy_updates_enabled":true}`,
			`{"policy_office_hours":true}`,
			`{"policy_timezone":"Asia/Tokyo"}`,
			// The spec refuses these before the handler runs.
			`{"policy_update_timeout":"soon"}`,
			`{"policy_period_interval":"3000000 days"}`,
			`{"policy_safe_mod":true}`,
		} {
			code, body := request(t, "PATCH", edgeURL()+overrides, strings.NewReader(payload))
			assert.Equal(t, http.StatusBadRequest, code, payload)
			assert.Contains(t, body, `"message":`, payload)
		}
	})

	t.Run("unknown_or_malformed_group", func(t *testing.T) {
		for id, want := range map[string]int{unknownID: http.StatusNotFound, "not-a-uuid": http.StatusBadRequest} {
			code, body := request(t, "PATCH", edgeURL()+fmt.Sprintf("/api/apps/%s/groups/%s/local_overrides", seededAppID, id), strings.NewReader(`{}`))
			assert.Equal(t, want, code, body)
		}
	})
}
