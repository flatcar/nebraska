package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/kinvolk/nebraska/backend/cmd/nebraska/ginhelpers"
	"github.com/kinvolk/nebraska/backend/pkg/sessions"
	ginsessions "github.com/kinvolk/nebraska/backend/pkg/sessions/gin"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/memcache"
	memcachegob "github.com/kinvolk/nebraska/backend/pkg/sessions/memcache/gob"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/securecookie"
)

type OIDCAuthConfig struct {
	DefaultTeamID   string
	ClientID        string
	ClientSecret    string
	CallbackURL     string
	TokenPath       string
	IssuerURL       string
	AdminRoles      []string
	ViewerRoles     []string
	SessionAuthKey  []byte
	SessionCryptKey []byte
	RolesPath       string
}

type oidcAuth struct {
	clientID      string
	issuerURL     string
	callbackURL   string
	provider      *oidc.Provider
	verifier      *oidc.IDTokenVerifier
	oauthConfig   *oauth2.Config
	defaultTeamID string
	adminRoles    []string
	viewerRoles   []string
	stateMap      map[string]string
	rolesPath     string
	sessionStore  *sessions.Store
}

func NewOIDCAuthenticator(config *OIDCAuthConfig) Authenticator {
	ctx := context.Background()

	// setup oidc provider
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		logger.Error().AnErr("error", err).Msg("Error setting up oidc provider")
		return nil
	}

	oidcProviderConfig := &oidc.Config{
		ClientID:          config.ClientID,
		SkipClientIDCheck: true,
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config.CallbackURL,
		Scopes:       []string{oidc.ScopeOpenID},
	}

	verifier := provider.Verifier(oidcProviderConfig)
	// state map is used keep track of login and callback requests
	stateMap := map[string]string{}

	// setup session store
	cache := memcache.New(memcachegob.New())
	codec := securecookie.New(config.SessionAuthKey, config.SessionCryptKey)
	sessionStore := sessions.NewStore(cache, codec)

	return &oidcAuth{config.ClientID, config.IssuerURL, config.CallbackURL, provider, verifier, oauthConfig, config.DefaultTeamID, config.AdminRoles, config.ViewerRoles, stateMap, config.RolesPath, sessionStore}
}

func (oa *oidcAuth) SetupRouter(router ginhelpers.Router) {
	router.Use(ginsessions.SessionsMiddleware(oa.sessionStore, "oidc"))
	oidcRouter := router.Group("/login", "oidc")
	oidcRouter.GET("/cb", oa.loginCb)
	oidcRouter.GET("/", oa.login)
	oidcRouter.GET("/validate_token", oa.ValidateToken)
}

func (oa *oidcAuth) ValidateToken(c *gin.Context) {
	ctx := c.Request.Context()

	// set request id in response header
	requestID := c.Writer.Header().Get("X-Request-ID")

	// Check is the id token exists in the request
	token := tokenFromRequest(c)
	if token == "" {
		logger.Debug().Str("request_id", requestID).Msg("ValidateToken, Authorization header is empty")
		httpError(c, http.StatusUnauthorized)
		return
	}

	// If refresh token is not available in the session
	// mark the request as unauthorized so that the session
	// can be recreated with refresh_token
	session := ginsessions.GetSession(c)
	refreshToken := session.Get("refresh_token")
	if refreshToken == nil {
		logger.Debug().Str("request_id", requestID).Msg("ValidateToken, Refresh token not found in session")
		httpError(c, http.StatusUnauthorized)
		return
	}

	_, err := oa.verifier.Verify(ctx, token)
	if err != nil {
		logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("ValidateToken, Token verification error")
		httpError(c, http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, map[string]bool{"valid": true})
}

func (oa *oidcAuth) loginCb(c *gin.Context) {
	// set request id in response header
	requestID := c.Writer.Header().Get("X-Request-ID")

	// Validate State received from provider
	state := c.Request.URL.Query().Get("state")

	rurl, ok := oa.stateMap[state]
	if !ok {
		logger.Error().Str("request_id", requestID).Msg("Returned state not found in state map")
		httpError(c, http.StatusInternalServerError)
		return
	}

	// Exchange code for token from oidc provider
	token, err := oa.oauthConfig.Exchange(c.Request.Context(), c.Request.URL.Query().Get("code"))
	if err != nil {
		logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("Can't exchange code for token")
		httpError(c, http.StatusInternalServerError)
		return
	}

	// Extract id_token from claims
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		logger.Error().Str("request_id", requestID).Msg("Token doesn't contain ID Token")
		httpError(c, http.StatusInternalServerError)
		return
	}

	// Store refresh_token in session
	session := ginsessions.GetSession(c)
	session.Set("refresh_token", token.RefreshToken)
	sessionSave(c, session, "login_cb")

	// delete state from map as the callback is processed.
	// also prevents reusing the same callback URL once again.
	delete(oa.stateMap, state)

	// Add token to redirect url provided by the client
	redirectURL, err := url.Parse(rurl)
	if err != nil {
		logger.Error().Str("request_id", requestID).Msg(fmt.Sprintf("Invalid redirect URL: %s", rurl))
		httpError(c, http.StatusBadRequest)
		return
	}
	q := redirectURL.Query()
	q.Set("id_token", idToken)
	redirectURL.RawQuery = q.Encode()
	redirectTo(c, redirectURL.String())
}

func (oa *oidcAuth) login(c *gin.Context) {
	// check if login_redirect_url is present in query params
	loginRedirectURL := c.Request.URL.Query().Get("login_redirect_url")

	// Generate AuthURL of the OIDC provider
	authURL, err := url.Parse(oa.provider.Endpoint().AuthURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Invalid Auth Endpoint by Provider")
		return
	}
	// Generate state to validate callback from provider
	state := uuid.New().String()
	q := authURL.Query()
	q.Set("client_id", oa.clientID)
	q.Set("redirect_uri", oa.callbackURL)
	q.Set("scope", "openid")
	q.Set("response_type", "code")
	q.Set("response_mode", "query")
	q.Set("state", state)

	// Store login redirect url in State map
	if loginRedirectURL == "" {
		oa.stateMap[state] = c.Request.URL.String()
	} else {
		oa.stateMap[state] = loginRedirectURL
	}

	authURL.RawQuery = q.Encode()

	// Redirect to generated AuthURL
	redirectTo(c, authURL.String())
}

// tokenFromRequest extracts token from request. Returns empty string if not present.
func tokenFromRequest(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	split := strings.Split(token, " ")
	if len(split) == 2 {
		return split[1]
	}
	return c.Request.URL.Query().Get("id_token")
}

// rolesFromToken extracts roles from a token. Returns empty array if not present.
func rolesFromToken(token *oidc.IDToken, rolesPath string) ([]string, error) {
	roles := []string{}
	var claimsString interface{}
	err := token.Claims(&claimsString)
	if err != nil {
		return roles, err
	}
	out, err := json.Marshal(claimsString)
	if err != nil {
		return roles, err
	}

	result := gjson.Get(string(out), rolesPath)
	result.ForEach(func(key, value gjson.Result) bool {
		roles = append(roles, value.String())
		return true
	})
	return roles, nil
}

func (oa *oidcAuth) Authenticate(c *gin.Context) (teamID string, replied bool) {
	ctx := c.Request.Context()

	// set request id in response header
	requestID := c.Writer.Header().Get("X-Request-ID")

	// Check is the id token exists in the request
	token := tokenFromRequest(c)
	if token == "" {
		logger.Debug().Str("request_id", requestID).Msg("Authorization header is empty")
		httpError(c, http.StatusUnauthorized)
		return "", true
	}

	// If refresh token is not available in the session
	// mark the request as unauthorized so that the session
	// can be recreated with refresh_token
	session := ginsessions.GetSession(c)
	refreshToken := session.Get("refresh_token")
	if refreshToken == nil {
		logger.Debug().Str("request_id", requestID).Msg("Refresh token not found in session")
		httpError(c, http.StatusUnauthorized)
		return "", true
	}

	// Verify Token
	tk, err := oa.verifier.Verify(ctx, token)
	if err != nil {
		// If token is expired, use the refresh_token to fetch a new token
		// and set the new id_token in response header
		if strings.Contains(err.Error(), "token is expired") {
			ts := oa.oauthConfig.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken.(string)})
			newToken, err := ts.Token()
			if err != nil {
				logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("ID Token refresh error")
				httpError(c, http.StatusInternalServerError)
				return "", true
			}

			idToken, ok := newToken.Extra("id_token").(string)
			if !ok {
				logger.Debug().Str("request_id", requestID).Msg("New Token doesn't contain ID Token")
				httpError(c, http.StatusInternalServerError)
				return "", true
			}
			c.Writer.Header().Set("id_token", idToken)
			session.Set("refresh_token", newToken.RefreshToken)
			tk, err = oa.verifier.Verify(ctx, idToken)
			if err != nil {
				logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("Can't Verify New ID Token")
				httpError(c, http.StatusInternalServerError)
				return "", true
			}
		} else {
			logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("Token verification error")
			httpError(c, http.StatusUnauthorized)
			return "", true
		}
	}

	roles, err := rolesFromToken(tk, oa.rolesPath)
	if err != nil {
		logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("Can't extract roles from token")
		httpError(c, http.StatusInternalServerError)
		return "", true
	}

	accessLevel := ""

	// Check and set access level
checkloop:
	for _, role := range roles {
		if accessLevel != "viewer" {
			for _, roRole := range oa.viewerRoles {
				if roRole == role {
					accessLevel = "viewer"
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

	// If access level is empty or and doesn't match role scope return error
	if accessLevel == "" {
		logger.Debug().Msg("Misconfigured Roles, Can't get access level from token")
		httpError(c, http.StatusForbidden)
		return "", true
	} else if accessLevel != "admin" {
		if c.Request.Method != "HEAD" && c.Request.Method != "GET" {
			logger.Error().Str("request_id", requestID).Msg("User doesn't have admin access")
			httpError(c, http.StatusForbidden)
			return "", true
		}
	}

	return oa.defaultTeamID, false
}
