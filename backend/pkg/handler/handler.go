package handler

import (
	"encoding/xml"
	"io/ioutil"
	"net/url"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/auth"
	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/omaha"
	"github.com/kinvolk/nebraska/backend/pkg/util"
	"github.com/kinvolk/nebraska/backend/pkg/version"
)

const (
	UpdateMaxRequestSize      = 64 * 1024
	GithubAccessManagementURL = "https://github.com/settings/apps/authorizations"
)

type ClientConfig struct {
	AccessManagementURL string `json:"access_management_url"`
	LogoutURL           string `json:"logout_url"`
	NebraskaVersion     string `json:"nebraska_version"`
	Logo                string `json:"logo"`
	Title               string `json:"title"`
	HeaderStyle         string `json:"header_style"`
	LoginURL            string `json:"login_url"`
	AuthMode            string `json:"auth_mode"`
}

type Handler struct {
	db           *api.API
	omahaHandler *omaha.Handler
	conf         *config.Config
	clientConf   *ClientConfig
	auth         auth.Authenticator
}

var defaultPage int = 1
var defaultPerPage int = 10

var logger = util.NewLogger("nebraska")

func New(db *api.API, conf *config.Config, auth auth.Authenticator) (*Handler, error) {
	clientConfig := &ClientConfig{
		AuthMode:        conf.AuthMode,
		NebraskaVersion: version.Version,
		Title:           conf.AppTitle,
		HeaderStyle:     conf.AppHeaderStyle,
	}

	if conf.AppLogoPath != "" {
		svg, err := ioutil.ReadFile(conf.AppLogoPath)
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

	if conf.AuthMode == "oidc" {
		url, err := url.Parse(conf.NebraskaURL)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid nebraska-url")
			return nil, err
		}
		url.Path = "/login"
		clientConfig.LoginURL = url.String()
		clientConfig.AccessManagementURL = conf.OidcManagementURL
		clientConfig.LogoutURL = conf.OidcLogutURL
	}

	return &Handler{db, omaha.NewHandler(db), conf, clientConfig, auth}, nil
}
