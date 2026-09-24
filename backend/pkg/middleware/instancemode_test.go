package middleware

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

// The admin surface, written out here independently of api/spec.yaml so that
// moving an endpoint across the boundary has to be done twice.
var adminOperations = []string{
	"DELETE /api/apps/:appIDorProductID",
	"DELETE /api/apps/:appIDorProductID/channels/:channelID",
	"DELETE /api/apps/:appIDorProductID/groups/:groupID",
	"DELETE /api/apps/:appIDorProductID/packages/:packageID",
	"DELETE /api/channels/:channelID/floors/:packageID",
	"POST /api/apps",
	"POST /api/apps/:appIDorProductID/channels",
	"POST /api/apps/:appIDorProductID/groups",
	"POST /api/apps/:appIDorProductID/packages",
	"PUT /api/apps/:appIDorProductID",
	"PUT /api/apps/:appIDorProductID/channels/:channelID",
	"PUT /api/apps/:appIDorProductID/groups/:groupID",
	"PUT /api/apps/:appIDorProductID/packages/:packageID",
	"PUT /api/channels/:channelID/floors/:packageID",
}

func testOperations(t *testing.T) Operations {
	t.Helper()

	swagger, err := codegen.GetSwagger()
	require.NoError(t, err)

	ops, err := ClassifyOperations(swagger)
	require.NoError(t, err)

	return ops
}

func TestClassifyOperationsCoversTheSpec(t *testing.T) {
	ops := testOperations(t)

	counts := map[string]int{}

	var admin []string

	for template, methods := range ops {
		for method, role := range methods {
			counts[role]++

			if role == RoleAdmin {
				admin = append(admin, method+" "+template)
			}
		}
	}

	sort.Strings(admin)
	assert.Equal(t, adminOperations, admin)
	assert.NotZero(t, counts[RoleRuntime])
	assert.NotZero(t, counts[RoleNone])
}

// routeStub satisfies codegen.ServerInterface without implementing anything.
// Registration only takes the method values, so nothing here is ever called.
type routeStub struct {
	codegen.ServerInterface
}

// The guard keys on the route template echo reports for a matched request, so
// the templates derived from the spec have to be the ones echo registered.
func TestClassificationMatchesRegisteredRoutes(t *testing.T) {
	e := echo.New()
	codegen.RegisterHandlers(e, routeStub{})

	var registered, classified []string

	for _, route := range e.Routes() {
		registered = append(registered, route.Method+" "+route.Path)
	}

	for template, methods := range testOperations(t) {
		for method := range methods {
			classified = append(classified, method+" "+template)
		}
	}

	sort.Strings(registered)
	sort.Strings(classified)
	assert.Equal(t, registered, classified)
}

func TestRequireControlNodeRefusesAdminOperations(t *testing.T) {
	ops := testOperations(t)

	e := echo.New()
	e.Use(RequireControlNode(ops))

	for template, methods := range ops {
		for method := range methods {
			e.Add(method, template, func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})
		}
	}

	for template, methods := range ops {
		for method, role := range methods {
			t.Run(method+" "+template, func(t *testing.T) {
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, httptest.NewRequest(method, concreteURL(template), nil))

				if role == RoleAdmin {
					assert.Equal(t, http.StatusForbidden, rec.Code)
					assert.Contains(t, rec.Body.String(), "admin_operation_not_accepted")
					assert.Contains(t, rec.Body.String(), "control node")

					return
				}

				assert.Equal(t, http.StatusOK, rec.Code, "%s must be served on every node", role)
			})
		}
	}
}

func TestClassifyOperationsRejectsUnplacedOperations(t *testing.T) {
	for _, tc := range []struct {
		name       string
		extensions map[string]any
	}{
		{name: "missing"},
		{name: "unknown value", extensions: map[string]any{nodeRoleExtension: "primary"}},
		{name: "empty value", extensions: map[string]any{nodeRoleExtension: ""}},
		{name: "wrong type", extensions: map[string]any{nodeRoleExtension: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			swagger := &openapi3.T{Paths: openapi3.NewPaths(
				openapi3.WithPath("/api/thing", &openapi3.PathItem{
					Post: &openapi3.Operation{Extensions: tc.extensions},
				}))}

			_, err := ClassifyOperations(swagger)
			require.Error(t, err)
			assert.Contains(t, err.Error(), nodeRoleExtension)
			assert.Contains(t, err.Error(), "POST /api/thing")
		})
	}
}

func TestEchoTemplate(t *testing.T) {
	assert.Equal(t, "/health", echoTemplate("/health"))
	assert.Equal(t, "/api/apps", echoTemplate("/api/apps"))
	assert.Equal(t, "/api/apps/:appIDorProductID", echoTemplate("/api/apps/{appIDorProductID}"))
	assert.Equal(t, "/api/apps/:appIDorProductID/groups/:groupID/updates_override",
		echoTemplate("/api/apps/{appIDorProductID}/groups/{groupID}/updates_override"))
}

// concreteURL fills a route template in with values, so that a request reaches
// the route and the guard sees the template again through echo.Context.Path.
func concreteURL(template string) string {
	segments := strings.Split(template, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = "x"
		}
	}

	return strings.Join(segments, "/")
}
