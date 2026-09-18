package middleware

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// nodeRoleExtension names the OpenAPI extension that places an operation on the
// nodes allowed to serve it. It is mandatory, so an endpoint added without a
// decision about where it belongs fails ClassifyOperations instead of picking up
// a default.
const nodeRoleExtension = "x-nebraska-node-role"

const (
	// RoleAdmin operations write admin-owned tables, which only a control node does.
	RoleAdmin = "admin"
	// RoleRuntime operations read, or write runtime-owned tables, on every node.
	RoleRuntime = "runtime"
	// RoleNone operations touch no database at all.
	RoleNone = "none"
)

// Operations holds the node role of every operation, keyed by the route template
// Echo matches on and then by HTTP method.
type Operations map[string]map[string]string

// ClassifyOperations reads the node role of every operation in the spec.
func ClassifyOperations(swagger *openapi3.T) (Operations, error) {
	ops := make(Operations, swagger.Paths.Len())

	var unclassified []string

	for path, item := range swagger.Paths.Map() {
		for method, operation := range item.Operations() {
			role, ok := operation.Extensions[nodeRoleExtension].(string)
			if !ok || (role != RoleAdmin && role != RoleRuntime && role != RoleNone) {
				unclassified = append(unclassified, method+" "+path)
				continue
			}

			template := echoTemplate(path)
			if ops[template] == nil {
				ops[template] = make(map[string]string)
			}

			ops[template][method] = role
		}
	}

	if len(unclassified) > 0 {
		sort.Strings(unclassified)
		return nil, fmt.Errorf("%s must be %s, %s or %s on every operation, and is missing or invalid on: %s",
			nodeRoleExtension, RoleAdmin, RoleRuntime, RoleNone, strings.Join(unclassified, ", "))
	}

	return ops, nil
}

// RequireControlNode refuses the operations that write admin-owned tables, for a
// node that does not accept admin writes. Reads, the Omaha endpoint and the
// runtime endpoints are served as usual.
func RequireControlNode(ops Operations) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ops[c.Path()][c.Request().Method] == RoleAdmin {
				return c.JSON(http.StatusForbidden, map[string]any{
					"error":       "admin_operation_not_accepted",
					"description": "Admin operations are only accepted on the control node.",
				})
			}

			return next(c)
		}
	}
}

// echoTemplate rewrites an OpenAPI path into the route template registered for
// it, which is what echo.Context.Path reports once a request has matched.
func echoTemplate(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			segments[i] = ":" + segment[1:len(segment)-1]
		}
	}

	return strings.Join(segments, "/")
}
