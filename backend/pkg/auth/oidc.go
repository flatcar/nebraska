package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
		return nil, fmt.Errorf("Error setting up oidc provider: %w", err)
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

func (oa *oidcAuth) SetupRouter(router *echo.Echo) {
	// No setup needed for stateless token validation
}

func (oa *oidcAuth) ValidateToken(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Writer.Header().Get("X-Request-ID")

	token := tokenFromRequest(c)
	if token == "" {
		logger.Debug().Str("request_id", requestID).Msg("ValidateToken, Authorization header is empty")
		httpError(c, http.StatusUnauthorized)
		return nil
	}

	// Verify JWT access token
	_, err := oa.verifier.Verify(ctx, token)
	if err != nil {
		logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("ValidateToken, Access token verification failed")
		httpError(c, http.StatusUnauthorized)
		return nil
	}

	return c.JSON(http.StatusOK, map[string]bool{"valid": true})
}

func (oa *oidcAuth) LoginCb(c echo.Context) error {
	// OAuth flow is handled by frontend directly with OIDC provider
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"error": "OAuth flow is handled by frontend. Use direct OIDC provider authorization.",
	})
}

func (oa *oidcAuth) Login(c echo.Context) error {
	// OAuth flow is handled by frontend directly with OIDC provider
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"error": "OAuth flow is handled by frontend. Use direct OIDC provider authorization.",
	})
}

func (oa *oidcAuth) LoginToken(c echo.Context) error {
	// Password grant type is deprecated and removed for security reasons
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"error": "Password grant type is not supported. Please use the authorization code flow.",
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

	var claims map[string]interface{}
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
checkloop:
	for _, role := range roles {
		if accessLevel != "viewer" {
			for _, roRole := range oa.viewerRoles {
				if roRole == role {
					accessLevel = "viewer"
					break
				}
			}
		}

		for _, rwRole := range oa.adminRoles {
			if rwRole == role {
				accessLevel = "admin"
				break checkloop
			}
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
		logger.Debug().Str("request_id", requestID).Msg("Authorization header is empty")
		httpError(c, http.StatusUnauthorized)
		return "", true
	}

	// Verify JWT access token
	accessToken, err := oa.verifier.Verify(ctx, token)
	if err != nil {
		logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("Access token verification failed")
		httpError(c, http.StatusUnauthorized)
		return "", true
	}

	// Extract roles from access token claims
	roles, err := rolesFromToken(accessToken, oa.rolesPath)
	if err != nil {
		logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("Can't extract roles from access token")
		httpError(c, http.StatusInternalServerError)
		return "", true
	}

	accessLevel := oa.determineAccessLevel(roles)

	// If access level is empty or doesn't match role scope then return an error
	if accessLevel == "" {
		logger.Debug().Msg("Misconfigured Roles, Can't get access level from access token")
		httpError(c, http.StatusForbidden)
		return "", true
	} else if accessLevel != "admin" {
		if c.Request().Method != "HEAD" && c.Request().Method != "GET" {
			logger.Error().Str("request_id", requestID).Msg("User doesn't have admin access")
			httpError(c, http.StatusForbidden)
			return "", true
		}
	}

	return oa.defaultTeamID, false
}


func (oa *oidcAuth) LoginWebhook(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func httpError(c echo.Context, status int) {
	//nolint:errcheck
	c.NoContent(status)
}

func redirectTo(c echo.Context, where string) {
	//nolint:errcheck
	c.Redirect(http.StatusTemporaryRedirect, where)
}

