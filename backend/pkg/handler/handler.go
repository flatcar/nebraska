package handler

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/auth"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/logger"
	"github.com/flatcar/nebraska/backend/pkg/omaha"
	"github.com/flatcar/nebraska/backend/pkg/version"
)

const (
	UpdateMaxRequestSize      = 64 * 1024
	GithubAccessManagementURL = "https://github.com/settings/apps/authorizations"
)

type Handler struct {
	db           *api.API
	omahaHandler *omaha.Handler
	conf         *config.Config
	clientConf   *codegen.Config
	auth         auth.Authenticator
}

var defaultPage = 1
var defaultPerPage = 10

var l = logger.New("nebraska")

func New(db *api.API, conf *config.Config, auth auth.Authenticator) (*Handler, error) {
	clientConfig := &codegen.Config{
		AuthMode:        conf.AuthMode,
		NebraskaVersion: version.Version,
		Title:           conf.AppTitle,
		HeaderStyle:     conf.AppHeaderStyle,
	}

	if conf.AppLogoPath != "" {
		svg, err := os.ReadFile(conf.AppLogoPath)
		if err != nil {
			l.Error().Err(err).Msg("Reading svg from path in config")
			return nil, err
		}
		if err := xml.Unmarshal(svg, &struct{}{}); err != nil {
			l.Error().Err(err).Msg("Invalid format for SVG")
			return nil, err
		}
		clientConfig.Logo = string(svg)
	}

	if conf.AuthMode == "github" {
		clientConfig.AccessManagementUrl = "https://github.com/settings/connections/applications/" + conf.GhClientID
	}

	if conf.AuthMode == "oidc" {
		url, err := url.Parse(conf.NebraskaURL)
		if err != nil {
			l.Error().Err(err).Msg("Invalid nebraska-url")
			return nil, err
		}
		url.Path = "/login"
		clientConfig.LoginUrl = url.String()
		clientConfig.AccessManagementUrl = conf.OidcManagementURL
		clientConfig.LogoutUrl = conf.OidcLogutURL
	}

	return &Handler{db, omaha.NewHandler(db), conf, clientConfig, auth}, nil
}

func (h *Handler) Health(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "OK")
}
