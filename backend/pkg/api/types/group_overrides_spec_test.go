package types_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

// TestGroupPolicyOverridesPatchMatchesTheSpec keeps the patch type in step with
// the fields and interval pattern the API spec states.
func TestGroupPolicyOverridesPatchMatchesTheSpec(t *testing.T) {
	swagger, err := codegen.GetSwagger()
	require.NoError(t, err)
	schemaRef, ok := swagger.Components.Schemas["groupLocalOverridesPatch"]
	require.True(t, ok)
	schema := schemaRef.Value

	// Every field the spec names is one the patch takes, and there are as many.
	body := map[string]any{}
	for name := range schema.Properties {
		body[name] = nil
	}
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	var patch types.GroupPolicyOverridesPatch
	require.NoError(t, json.Unmarshal(encoded, &patch))
	assert.Len(t, patch.Changes(), len(schema.Properties))
	assert.Equal(t, reflect.TypeOf(patch).NumField(), len(schema.Properties))

	for _, name := range []string{"policy_period_interval", "policy_update_timeout"} {
		property, ok := schema.Properties[name]
		require.True(t, ok, name)
		assert.Equal(t, types.OverrideIntervalPattern, property.Value.Pattern, name)
	}
}
