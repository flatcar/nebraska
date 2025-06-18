package server

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
	"github.com/pkg/errors"

	db "github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/auth"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/handler"
	custommiddleware "github.com/kinvolk/nebraska/backend/pkg/middleware"
	"github.com/kinvolk/nebraska/backend/pkg/sessions"
	echosessions "github.com/kinvolk/nebraska/backend/pkg/sessions/echo"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/memcache"
	memcachegob "github.com/kinvolk/nebraska/backend/pkg/sessions/memcache/gob"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/securecookie"
	"github.com/kinvolk/nebraska/backend/pkg/util"
)

const serviceName = "nebraska"

var (
	logger            = util.NewLogger("nebraska")
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
		e.Logger.SetLevel(log.DEBUG)
		e.Debug = conf.Debug
	}

	swagger, err := codegen.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("Swagger config error: %w", err)
	}

	p := prometheus.NewPrometheus(serviceName, nil)
	p.Use(e)

	// setup authenticator
	defaultTeam, err := db.GetTeam()
	if err != nil {
		return nil, fmt.Errorf("Cannot fetch the default teamID: %w", err)
	}

	// setup session store
	sessionStore := setupSessionStore(*conf)

	authenticator, err := setupAuthenticator(*conf, sessionStore, defaultTeam.ID)
	if err != nil {
		return nil, fmt.Errorf("Authenticator setup error: %w", err)
	}
	if authenticator == nil {
		return nil, fmt.Errorf("Invalid auth mode %s", conf.AuthMode)
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
		return nil, fmt.Errorf("Error setting up handlers: %w", err)
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
					logger.Err(fileErr).Msg("Error serving index.html")
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
			logger.Err(err).Msg("Error updating instance stats")
		}
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			err := db.UpdateInstanceStats(nil, nil)
			if err != nil {
				logger.Err(err).Msg("Error updating instance stats")
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
		oidcAuthConfig := &auth.OIDCAuthConfig{
			DefaultTeamID: defaultTeamID,
			IssuerURL:     conf.OidcIssuerURL,
			AdminRoles:    strings.Split(conf.OidcAdminRoles, ","),
			ViewerRoles:   strings.Split(conf.OidcViewerRoles, ","),
			RolesPath:     conf.OidcRolesPath,
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
					return errors.Wrap(err, "github cookie not found")
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
		return errors.New("Bearer token not found in request")
	}
	split := strings.Split(token, " ")
	if len(split) == 2 {
		if split[0] != "Bearer" {
			return errors.New("Bearer token not found in request")
		}
	} else {
		return errors.New("Invalid Bearer token")
	}
	return nil
}
