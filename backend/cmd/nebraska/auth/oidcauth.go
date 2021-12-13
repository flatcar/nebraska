package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	glob "github.com/ryanuber/go-glob"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/kinvolk/nebraska/backend/cmd/nebraska/ginhelpers"
	"github.com/kinvolk/nebraska/backend/pkg/sessions"
	ginsessions "github.com/kinvolk/nebraska/backend/pkg/sessions/gin"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/memcache"
	memcachegob "github.com/kinvolk/nebraska/backend/pkg/sessions/memcache/gob"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/securecookie"
)

const (
	stateCleanupDuration = time.Minute * 5
	stateTimeoutDuration = time.Minute * 3
)

type stateMessage struct {
	timeout     time.Time
	redirectURL string
}

type OIDCAuthConfig struct {
	DefaultTeamID     string
	ClientID          string
	ClientSecret      string
	CallbackURL       string
	TokenPath         string
	IssuerURL         string
	LogoutURL         string
	ManagementURL     string
	ValidRedirectURLs []string
	AdminRoles        []string
	ViewerRoles       []string
	SessionAuthKey    []byte
	SessionCryptKey   []byte
	RolesPath         string
	Scopes            []string
}

type oidcAuth struct {
	provider          *oidc.Provider
	verifier          *oidc.IDTokenVerifier
	oauthConfig       *oauth2.Config
	defaultTeamID     string
	clientID          string    // OIDC Client ID
	issuerURL         string    // OIDC Issuer URL
	callbackURL       string    // OIDC Callback URL, should be configured in OIDC provider. Default value is: http://localhost:8000/login/cb
	validRedirectURLs []string  // List of valid redirect URLs that the browser can redirect to after successful login
	adminRoles        []string  // List of roles with admin access
	viewerRoles       []string  // List of roles with viewer access
	scopes            []string  // List of OIDC scopes
	stateMap          *sync.Map // Map used to store state between nebraska and oidc provider, used to prevent fake authentication response
	rolesPath         string    // Json Path in which the roles will be found in ID Token
	sessionStore      *sessions.Store
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
		Scopes:       config.Scopes,
	}

	verifier := provider.Verifier(oidcProviderConfig)

	// state map is used keep track of login and callback requests
	var stateMap sync.Map

	// setup session store
	cache := memcache.New(memcachegob.New())
	codec := securecookie.New(config.SessionAuthKey, config.SessionCryptKey)
	sessionStore := sessions.NewStore(cache, codec)

	oidcAuthenticator := &oidcAuth{
		provider:          provider,
		verifier:          verifier,
		oauthConfig:       oauthConfig,
		defaultTeamID:     config.DefaultTeamID,
		clientID:          config.ClientID,
		issuerURL:         config.IssuerURL,
		callbackURL:       config.CallbackURL,
		validRedirectURLs: config.ValidRedirectURLs,
		adminRoles:        config.AdminRoles,
		viewerRoles:       config.ViewerRoles,
		scopes:            config.Scopes,
		stateMap:          &stateMap,
		rolesPath:         config.RolesPath,
		sessionStore:      sessionStore,
	}

	stateTicker := time.NewTicker(stateCleanupDuration)

	go func() {
		for {
			<-stateTicker.C
			oidcAuthenticator.cleanState()
		}
	}()

	return oidcAuthenticator
}

func (oa *oidcAuth) SetupRouter(router ginhelpers.Router) {
	router.Use(ginsessions.SessionsMiddleware(oa.sessionStore, "oidc"))
	oidcRouter := router.Group("/login", "oidc")
	oidcRouter.GET("/cb", oa.loginCb)
	oidcRouter.GET("/", oa.login)
	oidcRouter.GET("/validate_token", oa.validateToken)
}

func (oa *oidcAuth) validateToken(c *gin.Context) {
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
	// get request id from response header
	requestID := c.Writer.Header().Get("X-Request-ID")

	// Validate State received from provider
	state := c.Request.URL.Query().Get("state")

	rurl, ok := oa.stateMap.Load(state)
	if !ok {
		logger.Error().Str("request_id", requestID).Msg("Returned state not found in state map")
		httpError(c, http.StatusInternalServerError)
		return
	}

	// delete state from map as the callback is processed.
	// this prevents the same state from being reused again.
	defer oa.stateMap.Delete(state)

	message, ok := rurl.(stateMessage)
	if !ok {
		logger.Error().Str("request_id", requestID).Msg("Cannot get stateMessage from state value")
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
	oidcToken, err := oa.verifier.Verify(c.Request.Context(), idToken)
	if err != nil {
		logger.Error().Str("request_id", requestID).AnErr("error", err).Msg("Can't verify the token")
		httpError(c, http.StatusInternalServerError)
		return
	}

	// Store refresh_token in session
	session := ginsessions.GetSession(c)
	session.Set("refresh_token", token.RefreshToken)
	session.Set("username", oidcToken.Subject)
	sessionSave(c, session, "login_cb")

	// Add token to redirect url provided by the client
	redirectURL, err := url.Parse(message.redirectURL)
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
	isValidRedirect := false
	for _, redirectURL := range oa.validRedirectURLs {
		if glob.Glob(fmt.Sprintf("%s*", redirectURL), loginRedirectURL) {
			isValidRedirect = true
			break
		}
	}

	if !isValidRedirect {
		c.String(http.StatusBadRequest, "Invalid login_redirect_url")
		return
	}

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
	q.Set("scope", strings.Join(oa.scopes, " "))
	q.Set("response_type", "code")
	q.Set("response_mode", "query")
	q.Set("state", state)

	// Store login redirect url in State map
	msg := stateMessage{
		time.Now().Add(stateTimeoutDuration),
		loginRedirectURL,
	}
	oa.stateMap.Store(state, msg)

	authURL.RawQuery = q.Encode()

	// Redirect to generated AuthURL
	redirectTo(c, authURL.String())
}

// tokenFromRequest extracts token from request header. If Authorization header is not present returns id_token query param .
func tokenFromRequest(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	split := strings.Split(token, " ")
	if len(split) == 2 && split[0] == "Bearer" {
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

	// get request id from response header
	requestID := c.Writer.Header().Get("X-Request-ID")

	// Check if the id token exists in the request
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
				logger.Warn().Str("request_id", requestID).AnErr("error", err).Msg("Failed to use refresh token, reauthenticating")
				httpError(c, http.StatusUnauthorized)
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

	// If access level is empty or doesn't match role scope then return an error
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

func (oa *oidcAuth) cleanState() {
	now := time.Now()
	oa.stateMap.Range(func(key, value interface{}) bool {
		val, ok := value.(stateMessage)
		if !ok {
			oa.stateMap.Delete(key)
			return true
		}
		if now.After(val.timeout) {
			logger.Debug().Str("message", fmt.Sprintf("Deleting expired key %s from state map", key))
			oa.stateMap.Delete(key)
		}
		return true
	})
}
