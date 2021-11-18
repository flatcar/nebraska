package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	db "github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/auth"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/handler"
	"github.com/kinvolk/nebraska/backend/pkg/metrics"
	custommiddleware "github.com/kinvolk/nebraska/backend/pkg/middleware"
	"github.com/kinvolk/nebraska/backend/pkg/sessions"
	echosessions "github.com/kinvolk/nebraska/backend/pkg/sessions/echo"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/memcache"
	memcachegob "github.com/kinvolk/nebraska/backend/pkg/sessions/memcache/gob"
	"github.com/kinvolk/nebraska/backend/pkg/sessions/securecookie"
	"github.com/kinvolk/nebraska/backend/pkg/syncer"
	"github.com/kinvolk/nebraska/backend/pkg/util"
)

const serviceName = "nebraska"

var (
	logger            = util.NewLogger("nebraska")
	middlewareSkipper = func(c echo.Context) bool {
		requestPath := c.Path()

		paths := []string{"/health", "/metrics", "/config", "/v1/update", "/*"}
		for _, path := range paths {
			if requestPath == path {
				return true
			}
		}
		return false
	}
)

func main() {
	// config parse

	conf, err := config.Parse()
	if err != nil {
		log.Fatal("Error parsing config, err:", err)
	}

	err = conf.Validate()
	if err != nil {
		log.Fatal("Config is invaliad, err:", err)
	}

	// create new DB
	db, err := db.New()
	if err != nil {
		log.Fatal("Api err:", err)
	}

	// Setup Echo Server
	e := echo.New()

	// setup logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if conf.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		e.Logger.SetLevel(log.DEBUG)
		e.Debug = conf.Debug
	}

	swagger, err := codegen.GetSwagger()
	if err != nil {
		log.Fatal("Swagger config error", err)
	}

	// setup and instrument metrics
	err = metrics.RegisterAndInstrument(db)
	if err != nil {
		log.Fatal("Metrics register error", err)
	}

	p := prometheus.NewPrometheus(serviceName, nil)
	p.Use(e)

	// setup authenticator
	defaultTeam, err := db.GetTeam()
	if err != nil {
		logger.Fatal().Err(err).Msg("Cannot fetch the default teamID")
	}

	// setup session store
	sessionStore := setupSessionStore(*conf)

	authenticator, err := setupAuthenticator(*conf, sessionStore, defaultTeam.ID)
	if err != nil {
		logger.Fatal().Err(err).Msg("Authenticator error")
	}
	if authenticator == nil {
		logger.Fatal().Msgf("Invalid auth mode %s", conf.AuthMode)
	}

	// setup middlewares
	if conf.APIEndpointSuffix != "" {
		e.Pre(custommiddleware.OmahaSecret(conf.APIEndpointSuffix))
	}
	e.Pre(middleware.RemoveTrailingSlash())
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
	e.Use(custommiddleware.Auth(authenticator, custommiddleware.AuthConfig{Skipper: custommiddleware.NewAuthSkipper(conf.AuthMode)}))
	e.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, &oapimiddleware.Options{Options: openapi3filter.Options{AuthenticationFunc: nebraskaAuthenticationFunc(conf.AuthMode)}, Skipper: middlewareSkipper}))

	// setup syncer
	checkFrequency, err := time.ParseDuration(conf.CheckFrequencyVal)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid Check Frequency value")
	}

	if conf.SyncerPkgsURL == "" && conf.HostFlatcarPackages {
		conf.SyncerPkgsURL = conf.NebraskaURL + "/flatcar/"
	}

	if conf.EnableSyncer {
		syncer, err := syncer.New(&syncer.Config{
			API:               db,
			HostPackages:      conf.HostFlatcarPackages,
			PackagesPath:      conf.FlatcarPackagesPath,
			PackagesURL:       conf.SyncerPkgsURL,
			FlatcarUpdatesURL: conf.FlatcarUpdatesURL,
			CheckFrequency:    checkFrequency,
		})
		if err != nil {
			logger.Fatal().Err(err).Msg("Error setting up syncer")
		}

		go syncer.Start()
		defer syncer.Stop()
	}

	// setup handler
	handlers, err := handler.New(db, conf, authenticator)
	if err != nil {
		log.Fatal("Error setting up handlers, err:", err)
	}

	e.Static("/", conf.HTTPStaticDir)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusNotFound
		if he, ok := err.(*echo.HTTPError); ok {
			if code == he.Code {
				fileErr := c.File(path.Join(conf.HTTPStaticDir, "index.html"))
				logger.Err(fileErr).Msg("Error serving index.html")
				return
			}
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
	codegen.RegisterHandlers(e, handlers)

	// run server
	log.Fatal(e.Start(fmt.Sprintf(":%d", conf.ServerPort)))
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
			return nil, errors.Wrap(err, "nebraska-url is invalid, can't generate oidc callback URL")
		}

		url.Path = "/login/cb"

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
		return auth.NewOIDCAuthenticator(oidcAuthConfig), nil
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
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		switch authMode {
		case "noop":
			return nil
		case "oidc":
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
		return nil
	}
}
