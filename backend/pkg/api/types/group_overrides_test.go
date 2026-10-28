package types

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}
	os.Exit(m.Run())
}

func decodePatch(t *testing.T, body string) (*GroupPolicyOverridesPatch, error) {
	t.Helper()

	var patch GroupPolicyOverridesPatch
	err := json.Unmarshal([]byte(body), &patch)
	return &patch, err
}

func TestGroupPolicyOverridesPatchTellsLeftOutFromNull(t *testing.T) {
	patch, err := decodePatch(t, `{"policy_safe_mode": null, "policy_max_updates_per_period": 3, "policy_timezone": "UTC"}`)
	require.NoError(t, err)

	assert.Nil(t, patch.PolicyUpdatesEnabled, "a field left out keeps its override")
	assert.Equal(t, &null.Bool{}, patch.PolicySafeMode, "null removes the override")
	assert.Equal(t, null.IntFrom(3), *patch.PolicyMaxUpdatesPerPeriod)
	assert.Equal(t, map[string]any{
		"policy_safe_mode":              null.Bool{},
		"policy_max_updates_per_period": null.IntFrom(3),
		"policy_timezone":               null.StringFrom("UTC"),
	}, patch.Changes())
}

// The patch names its fields by hand in its JSON tags, UnmarshalJSON and
// Changes, so this keeps all three in step with the overrides a group can have.
func TestGroupPolicyOverridesPatchCoversEveryOverride(t *testing.T) {
	overrides := reflect.TypeOf(GroupPolicyOverrides{})
	every := map[string]any{}
	for i := 0; i < overrides.NumField(); i++ {
		name, _, _ := strings.Cut(overrides.Field(i).Tag.Get("json"), ",")
		every[name] = nil
	}

	body, err := json.Marshal(every)
	require.NoError(t, err)
	patch, err := decodePatch(t, string(body))
	require.NoError(t, err)

	changes, err := json.Marshal(patch.Changes())
	require.NoError(t, err)
	assert.JSONEq(t, string(body), string(changes))
	encoded, err := json.Marshal(patch)
	require.NoError(t, err)
	assert.JSONEq(t, string(body), string(encoded))
	assert.Equal(t, overrides.NumField(), reflect.TypeOf(*patch).NumField())

	empty, err := decodePatch(t, `{}`)
	require.NoError(t, err)
	assert.Empty(t, empty.Changes())
}

// The generated API client marshals the patch, so a field it was not given must
// stay out of the body rather than be sent as a null that removes it.
func TestGroupPolicyOverridesPatchSendsOnlyWhatChanges(t *testing.T) {
	for _, body := range []string{
		`{}`,
		`{"policy_safe_mode":true}`,
		`{"policy_updates_enabled":null,"policy_max_updates_per_period":5,"policy_update_timeout":"2 hours"}`,
	} {
		patch, err := decodePatch(t, body)
		require.NoError(t, err)

		encoded, err := json.Marshal(*patch)
		require.NoError(t, err)
		assert.JSONEq(t, body, string(encoded))

		encoded, err = json.Marshal(patch)
		require.NoError(t, err)
		assert.JSONEq(t, body, string(encoded))
	}
}

func TestGroupPolicyOverridesPatchRefusesWhatItCannotRead(t *testing.T) {
	for name, body := range map[string]string{
		"unknown_field": `{"policy_safe_mod": true}`,
		"wrong_type":    `{"policy_safe_mode": "yes"}`,
		"null":          `null`,
		"not_an_object": `[]`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := decodePatch(t, body)
			assert.ErrorIs(t, err, ErrInvalidLocalOverrides)
		})
	}
}

func TestGroupPolicyOverridesPatchValidate(t *testing.T) {
	for name, body := range map[string]string{
		"nothing":               `{}`,
		"pause":                 `{"policy_updates_enabled": false}`,
		"release_the_pause":     `{"policy_updates_enabled": null}`,
		"time_zone":             `{"policy_timezone": "Europe/Berlin"}`,
		"no_time_zone":          `{"policy_timezone": null}`,
		"one_update_per_period": `{"policy_max_updates_per_period": 1}`,
		"most_the_column_holds": `{"policy_max_updates_per_period": 2147483647}`,
		"minutes_interval":      `{"policy_period_interval": "15 minutes", "policy_update_timeout": null}`,
		"longest_interval":      `{"policy_period_interval": "1 hour", "policy_update_timeout": "999999 days"}`,
	} {
		t.Run(name, func(t *testing.T) {
			patch, err := decodePatch(t, body)
			require.NoError(t, err)
			assert.NoError(t, patch.Validate())
		})
	}

	for name, body := range map[string]string{
		"enables_updates":            `{"policy_updates_enabled": true}`,
		"unknown_time_zone":          `{"policy_timezone": "Mars/Base"}`,
		"empty_time_zone":            `{"policy_timezone": ""}`,
		"servers_own_time_zone":      `{"policy_timezone": "Local"}`,
		"no_updates_per_period":      `{"policy_max_updates_per_period": 0}`,
		"more_than_the_column_holds": `{"policy_max_updates_per_period": 2147483648}`,
		"empty_interval":             `{"policy_update_timeout": ""}`,
		"zero_interval":              `{"policy_update_timeout": "0 minutes"}`,
		"negative_interval":          `{"policy_update_timeout": "-5 minutes"}`,
		"compound_interval":          `{"policy_period_interval": "1 hour 30 minutes"}`,
		"iso_8601_interval":          `{"policy_period_interval": "PT2H"}`,
		"unit_the_ui_cannot_show":    `{"policy_period_interval": "2 weeks"}`,
		"out_of_the_timestamp_range": `{"policy_period_interval": "3000000 days"}`,
		"not_an_interval":            `{"policy_update_timeout": "soon"}`,
	} {
		t.Run(name, func(t *testing.T) {
			patch, err := decodePatch(t, body)
			require.NoError(t, err)
			assert.ErrorIs(t, patch.Validate(), ErrInvalidLocalOverrides)
		})
	}
}
