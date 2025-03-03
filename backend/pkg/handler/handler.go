package handler

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/auth"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/omaha"
	"github.com/kinvolk/nebraska/backend/pkg/util"
	"github.com/kinvolk/nebraska/backend/pkg/version"
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

var logger = util.NewLogger("nebraska")

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
			logger.Error().Err(err).Msg("Reading svg from path in config")
			return nil, err
		}
		if err := xml.Unmarshal(svg, &struct{}{}); err != nil {
			logger.Error().Err(err).Msg("Invalid format for SVG")
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
			logger.Error().Err(err).Msg("Invalid nebraska-url")
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
