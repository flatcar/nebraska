package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
)

// The paginated list endpoints declare their item field as a required array in
// api/spec.yaml. A nil slice marshals to JSON null, which violates that
// contract and breaks clients generated from the spec. These assertions look at
// the raw JSON because unmarshalling into a Go slice treats null and [] alike.
func TestEmptyListsAreEmptyArraysNotNull(t *testing.T) {
	serverURL := os.Getenv("NEBRASKA_TEST_SERVER_URL")

	// A freshly created app has no groups, channels or packages.
	var app api.Application
	httpDo(t, serverURL+"/api/apps", http.MethodPost,
		strings.NewReader(`{"name":"empty-list-serialization-test"}`),
		http.StatusOK, "json", &app)
	require.NotEmpty(t, app.ID)

	for _, tc := range []struct {
		field string
		path  string
	}{
		{"groups", fmt.Sprintf("/api/apps/%s/groups", app.ID)},
		{"channels", fmt.Sprintf("/api/apps/%s/channels", app.ID)},
		{"packages", fmt.Sprintf("/api/apps/%s/packages", app.ID)},
	} {
		t.Run(tc.field, func(t *testing.T) {
			var page map[string]json.RawMessage
			httpDo(t, serverURL+tc.path, http.MethodGet, nil, http.StatusOK, "json", &page)

			raw, ok := page[tc.field]
			require.True(t, ok, "response is missing the %q field", tc.field)
			assert.JSONEq(t, "[]", string(raw),
				"%s must serialize as an empty array, not null", tc.field)
		})
	}
}
