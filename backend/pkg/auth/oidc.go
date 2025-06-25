package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

type OIDCAuthConfig struct {
	DefaultTeamID string
	IssuerURL     string
	AdminRoles    []string
	ViewerRoles   []string
	RolesPath     string
}

type oidcAuth struct {
	provider      *oidc.Provider
	verifier      *oidc.IDTokenVerifier
	defaultTeamID string
	issuerURL     string
	adminRoles    []string
	viewerRoles   []string
	rolesPath     string
}

func NewOIDCAuthenticator(config *OIDCAuthConfig) (Authenticator, error) {
	ctx := context.Background()

	// setup oidc provider
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("error setting up oidc provider: %w", err)
	}

	// Configure verifier for JWT access tokens (not ID tokens)
	oidcProviderConfig := &oidc.Config{
		SkipClientIDCheck: true, // Access tokens don't have client_id claim
		SkipExpiryCheck:   false,
		SkipIssuerCheck:   false,
	}

	verifier := provider.Verifier(oidcProviderConfig)

	oidcAuthenticator := &oidcAuth{
		provider:      provider,
		verifier:      verifier,
		defaultTeamID: config.DefaultTeamID,
		issuerURL:     config.IssuerURL,
		adminRoles:    config.AdminRoles,
		viewerRoles:   config.ViewerRoles,
		rolesPath:     config.RolesPath,
	}

	return oidcAuthenticator, nil
}

func (oa *oidcAuth) SetupRouter(_ *echo.Echo) {
	// No setup needed for stateless token validation
}

func (oa *oidcAuth) ValidateToken(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Writer.Header().Get("X-Request-ID")

	token := tokenFromRequest(c)
	if token == "" {
		l.Debug().Str("request_id", requestID).Msg("ValidateToken, Authorization header is empty")
		httpError(c, http.StatusUnauthorized)
		return nil
	}

	// Verify JWT access token
	_, err := oa.verifier.Verify(ctx, token)
	if err != nil {
		l.Error().Str("request_id", requestID).AnErr("error", err).Msg("ValidateToken, Access token verification failed")
		httpError(c, http.StatusUnauthorized)
		return nil
	}

	return c.JSON(http.StatusOK, map[string]bool{"valid": true})
}

// LoginCb is not used in the new OIDC architecture
// Frontend handles OAuth flow directly with OIDC provider
func (oa *oidcAuth) LoginCb(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]any{
		"error":       "login_callback_not_supported",
		"description": "OAuth callback flow is not supported in OIDC mode. The frontend handles OIDC authorization flow directly with the identity provider.",
		"docs":        "See OIDC migration guide for proper configuration.",
	})
}

// Login is not used in the new OIDC architecture
// Frontend handles OAuth flow directly with OIDC provider
func (oa *oidcAuth) Login(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]any{
		"error":       "login_endpoint_not_supported",
		"description": "Server-side login flow is not supported in OIDC mode. The frontend handles OIDC authorization flow directly with the identity provider.",
		"docs":        "See OIDC migration guide for proper configuration.",
	})
}

// tokenFromRequest extracts token from request header.
func tokenFromRequest(c echo.Context) string {
	token := c.Request().Header.Get("Authorization")
	split := strings.Split(token, " ")
	if len(split) == 2 && split[0] == "Bearer" {
		return split[1]
	}
	return ""
}

// rolesFromToken extracts roles from JWT access token claims
func rolesFromToken(token *oidc.IDToken, rolesPath string) ([]string, error) {
	roles := []string{}
	if rolesPath == "" {
		return roles, nil
	}

	var claims map[string]any
	if err := token.Claims(&claims); err != nil {
		return roles, err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return roles, err
	}

	result := gjson.Get(string(claimsJSON), rolesPath)
	result.ForEach(func(_, value gjson.Result) bool {
		roles = append(roles, value.String())
		return true
	})

	return roles, nil
}

// determineAccessLevel determines user access level based on roles
func (oa *oidcAuth) determineAccessLevel(roles []string) string {
	accessLevel := ""

	// Check and set access level
	for _, role := range roles {
		if slices.Contains(oa.adminRoles, role) {
			return "admin"
		}
		if accessLevel != "viewer" && slices.Contains(oa.viewerRoles, role) {
			accessLevel = "viewer"
		}
	}

	return accessLevel
}

func (oa *oidcAuth) Authorize(c echo.Context) (teamID string, replied bool) {
	ctx := c.Request().Context()
	requestID := c.Response().Writer.Header().Get("X-Request-ID")

	// Check if the JWT access token exists in the request
	token := tokenFromRequest(c)
	if token == "" {
		l.Debug().Str("request_id", requestID).Msg("Authorization header is empty")
		httpError(c, http.StatusUnauthorized)
		return "", true
	}

	// Verify JWT access token
	accessToken, err := oa.verifier.Verify(ctx, token)
	if err != nil {
		l.Error().Str("request_id", requestID).AnErr("error", err).Msg("Access token verification failed")
		httpError(c, http.StatusUnauthorized)
		return "", true
	}

	// Extract roles from access token claims
	roles, err := rolesFromToken(accessToken, oa.rolesPath)
	if err != nil {
		l.Error().Str("request_id", requestID).AnErr("error", err).Msg("Can't extract roles from access token")
		httpError(c, http.StatusInternalServerError)
		return "", true
	}

	accessLevel := oa.determineAccessLevel(roles)

	// If access level is empty or doesn't match role scope then return an error
	if accessLevel == "" {
		l.Debug().Msg("Misconfigured Roles, Can't get access level from access token")
		httpError(c, http.StatusForbidden)
		return "", true
	} else if accessLevel != "admin" {
		if c.Request().Method != "HEAD" && c.Request().Method != "GET" {
			l.Error().Str("request_id", requestID).Msg("User doesn't have admin access")
			httpError(c, http.StatusForbidden)
			return "", true
		}
	}

	return oa.defaultTeamID, false
}

func (oa *oidcAuth) LoginWebhook(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]any{
		"error":       "webhook_not_supported",
		"description": "Webhooks are not supported in OIDC mode. Webhooks are only used with GitHub authentication.",
	})
}

func httpError(c echo.Context, status int) {
	//nolint:errcheck
	c.NoContent(status)
}

func redirectTo(c echo.Context, where string) {
	//nolint:errcheck
	c.Redirect(http.StatusTemporaryRedirect, where)
}
