package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomiddleware "github.com/oapi-codegen/echo-middleware"

	db "github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/auth"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/handler"
	"github.com/flatcar/nebraska/backend/pkg/logger"
	custommiddleware "github.com/flatcar/nebraska/backend/pkg/middleware"
	"github.com/flatcar/nebraska/backend/pkg/sessions"
	echosessions "github.com/flatcar/nebraska/backend/pkg/sessions/echo"
	"github.com/flatcar/nebraska/backend/pkg/sessions/memcache"
	memcachegob "github.com/flatcar/nebraska/backend/pkg/sessions/memcache/gob"
	"github.com/flatcar/nebraska/backend/pkg/sessions/securecookie"
)

const serviceName = "nebraska"

var (
	l                 = logger.New("nebraska")
	middlewareSkipper = func(c echo.Context) bool {
		requestPath := c.Path()
		paths := []string{"/health", "/metrics", "/config", "/v1/update", "/flatcar/*", "/*"}
		for _, path := range paths {
			if requestPath == path {
				return true
			}
		}
		return false
	}
)

// New takes the config and db connection to create the server and returns it.
// It also starts a background job to update instance stats periodically.
func New(conf *config.Config, db *db.API) (*echo.Echo, error) {
	// Setup Echo Server
	e := echo.New()

	if conf.Debug {
		// SetLevel(0) means SetLevel(DEBUG)
		// but let's avoid pulling a 'log' dependency again (different from zerolog) just for this.
		e.Logger.SetLevel(0)
		e.Debug = conf.Debug
	}

	swagger, err := codegen.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("swagger config error: %w", err)
	}

	p := prometheus.NewPrometheus(serviceName, nil)
	p.Use(e)

	// setup authenticator
	defaultTeam, err := db.GetTeam()
	if err != nil {
		return nil, fmt.Errorf("cannot fetch the default teamID: %w", err)
	}

	// setup session store
	sessionStore := setupSessionStore(*conf)

	authenticator, err := setupAuthenticator(*conf, sessionStore, defaultTeam.ID)
	if err != nil {
		return nil, fmt.Errorf("authenticator setup error: %w", err)
	}
	if authenticator == nil {
		return nil, fmt.Errorf("invalid auth mode %s", conf.AuthMode)
	}

	// setup middlewares
	e.Pre(middleware.RemoveTrailingSlash())

	// remove trailing slash from the endpoint secret
	endpointSuffix := strings.TrimSuffix(conf.APIEndpointSuffix, "/")
	if endpointSuffix != "" {
		// if endpoint secret doesn't start with slash prepend it
		if !strings.HasPrefix(endpointSuffix, "/") {
			endpointSuffix = fmt.Sprintf("/%s", endpointSuffix)
		}
		e.Pre(custommiddleware.OmahaSecret(endpointSuffix))
	}
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Skipper: middlewareSkipper}))
	if conf.HTTPLog {
		e.Use(middleware.Logger())
	}
	if sessionStore != nil {
		e.Use(echosessions.SessionsMiddleware(sessionStore, conf.AuthMode))
	}
	e.Use(echomiddleware.OapiRequestValidatorWithOptions(swagger, &echomiddleware.Options{Options: openapi3filter.Options{AuthenticationFunc: nebraskaAuthenticationFunc(conf.AuthMode)}, Skipper: middlewareSkipper}))
	e.Use(custommiddleware.Auth(authenticator, custommiddleware.AuthConfig{Skipper: custommiddleware.NewAuthSkipper(conf.AuthMode)}))

	// setup handler
	handlers, err := handler.New(db, conf, authenticator)
	if err != nil {
		return nil, fmt.Errorf("error setting up handlers: %w", err)
	}

	e.Static("/", conf.HTTPStaticDir)

	if conf.HostFlatcarPackages && conf.FlatcarPackagesPath != "" {
		e.Static("/flatcar/", conf.FlatcarPackagesPath)
	}

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusNotFound
		if he, ok := err.(*echo.HTTPError); ok {
			if code == he.Code {
				fileErr := c.File(path.Join(conf.HTTPStaticDir, "index.html"))
				if fileErr != nil {
					l.Err(fileErr).Msg("Error serving index.html")
				}
				return
			}
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
	codegen.RegisterHandlers(e, handlers)

	// setup background job for updating instance stats
	go func() {
		// update once at startup
		err = db.UpdateInstanceStats(nil, nil)
		if err != nil {
			l.Err(err).Msg("Error updating instance stats")
		}
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			err := db.UpdateInstanceStats(nil, nil)
			if err != nil {
				l.Err(err).Msg("Error updating instance stats")
			}
		}
	}()

	return e, nil
}

func setupAuthenticator(conf config.Config, sessionStore *sessions.Store, defaultTeamID string) (auth.Authenticator, error) {
	switch conf.AuthMode {
	case "noop":
		noopAuthConfig := &auth.NoopAuthConfig{
			DefaultTeamID: defaultTeamID,
		}
		return auth.NewNoopAuthenticator(noopAuthConfig), nil
	case "github":
		gituhbAuthConfig := &auth.GithubAuthConfig{
			EnterpriseURL:     conf.GhEnterpriseURL,
			SessionStore:      sessionStore,
			OAuthClientID:     conf.GhClientID,
			OAuthClientSecret: conf.GhClientSecret,
			WebhookSecret:     conf.GhWebhookSecret,
			ReadWriteTeams:    strings.Split(conf.GhReadWriteTeams, ","),
			ReadOnlyTeams:     strings.Split(conf.GhReadOnlyTeams, ","),
			DefaultTeamID:     defaultTeamID,
		}
		return auth.NewGithubAuthenticator(gituhbAuthConfig), nil
	case "oidc":

		url, err := url.Parse(conf.NebraskaURL)
		if err != nil {
			return nil, fmt.Errorf("nebraska-url is invalid, can't generate oidc callback URL: %w", err)
		}

		url.Path = "/login/cb"
		if conf.OidcValidRedirectURLs == "" {
			url, err := url.Parse(conf.NebraskaURL)
			if err != nil {
				return nil, fmt.Errorf("nebraska-url is invalid, can't generate valid redirect URL, Err: %w", err)
			}
			url.Path = strings.TrimSuffix(url.Path, "/")
			generatedValidRedirectURLs := fmt.Sprintf("%s/*", url.String())
			conf.OidcValidRedirectURLs = generatedValidRedirectURLs
		}
		oidcAuthConfig := &auth.OIDCAuthConfig{
			DefaultTeamID:     defaultTeamID,
			ClientID:          conf.OidcClientID,
			ClientSecret:      conf.OidcClientSecret,
			IssuerURL:         conf.OidcIssuerURL,
			CallbackURL:       url.String(),
			ValidRedirectURLs: strings.Split(conf.OidcValidRedirectURLs, ","),
			ManagementURL:     conf.OidcManagementURL,
			LogoutURL:         conf.OidcLogutURL,
			AdminRoles:        strings.Split(conf.OidcAdminRoles, ","),
			ViewerRoles:       strings.Split(conf.OidcViewerRoles, ","),
			Scopes:            strings.Split(conf.OidcScopes, ","),
			SessionStore:      sessionStore,
			RolesPath:         conf.OidcRolesPath,
		}
		return auth.NewOIDCAuthenticator(oidcAuthConfig)
	}
	return nil, nil
}

func setupSessionStore(conf config.Config) *sessions.Store {
	switch conf.AuthMode {
	case "noop":
		return nil
	case "oidc":
		cache := memcache.New(memcachegob.New())
		codec := securecookie.New([]byte(conf.OidcSessionAuthKey), []byte(conf.OidcSessionCryptKey))
		return sessions.NewStore(cache, codec)
	case "github":
		cache := memcache.New(memcachegob.New())
		codec := securecookie.New([]byte(conf.GhSessionAuthKey), []byte(conf.GhSessionCryptKey))
		return sessions.NewStore(cache, codec)
	}
	return nil
}

func nebraskaAuthenticationFunc(authMode string) func(context.Context, *openapi3filter.AuthenticationInput) error {
	return func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
		switch authMode {
		case "noop":
			return nil
		case "oidc":
			// check if token is present in query params
			if input.RequestValidationInput.Request.URL.Query().Get("id_token") != "" {
				return nil
			}
			return validateAuthorizationToken(input)
		case "github":
			err := validateAuthorizationToken(input)
			if err != nil {
				_, err := input.RequestValidationInput.Request.Cookie("github")
				if err != nil {
					return fmt.Errorf("github cookie not found: %w", err)
				}
			}
			return nil
		}
		return nil
	}
}

// check if Authorization Header is present and valid
func validateAuthorizationToken(input *openapi3filter.AuthenticationInput) error {
	token := input.RequestValidationInput.Request.Header.Get("Authorization")
	if token == "" {
		return errors.New("bearer token not found in request")
	}
	split := strings.Split(token, " ")
	if len(split) == 2 {
		if split[0] != "Bearer" {
			return errors.New("bearer token not found in request")
		}
	} else {
		return errors.New("invalid Bearer token")
	}
	return nil
}
