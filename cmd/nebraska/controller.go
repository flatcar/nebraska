package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kinvolk/nebraska/cmd/nebraska/auth"
	"github.com/kinvolk/nebraska/pkg/api"
	"github.com/kinvolk/nebraska/pkg/omaha"
	"github.com/kinvolk/nebraska/pkg/syncer"
	"github.com/kinvolk/nebraska/pkg/version"
)

const (
	GithubAccessManagementURL = "https://github.com/settings/apps/authorizations"
	UpdateMaxRequestSize      = 64 * 1024
)

// ClientConfig represents Nebraska's configuration of interest for the client.
type ClientConfig struct {
	AccessManagementURL string `json:"access_management_url"`
	NebraskaVersion     string `json:"nebraska_version"`
	Logo                string `json:"logo"`
	Title               string `json:"title"`
	HeaderStyle         string `json:"header_style"`
}

type controller struct {
	api          *api.API
	omahaHandler *omaha.Handler
	syncer       *syncer.Syncer
	clientConfig *ClientConfig
	auth         auth.Authenticator
}

type controllerConfig struct {
	api                 *api.API
	enableSyncer        bool
	hostFlatcarPackages bool
	flatcarPackagesPath string
	nebraskaURL         string
	noopAuthConfig      *auth.NoopAuthConfig
	githubAuthConfig    *auth.GithubAuthConfig
	flatcarUpdatesURL   string
	checkFrequency      time.Duration
}

func newController(conf *controllerConfig) (*controller, error) {
	authenticator, err := getAuthenticator(conf)
	if err != nil {
		return nil, err
	}
	c := &controller{
		api:          conf.api,
		omahaHandler: omaha.NewHandler(conf.api),
		auth:         authenticator,
	}

	if conf.enableSyncer {
		syncerConf := &syncer.Config{
			API:               conf.api,
			HostPackages:      conf.hostFlatcarPackages,
			PackagesPath:      conf.flatcarPackagesPath,
			PackagesURL:       conf.nebraskaURL + "/flatcar/",
			FlatcarUpdatesURL: conf.flatcarUpdatesURL,
			CheckFrequency:    conf.checkFrequency,
		}
		syncer, err := syncer.New(syncerConf)
		if err != nil {
			return nil, err
		}
		c.syncer = syncer
		go syncer.Start()
	}

	c.clientConfig = NewClientConfig(conf)

	return c, nil
}

func (ctl *controller) close() {
	if ctl.syncer != nil {
		ctl.syncer.Stop()
	}
	ctl.api.Close()
}

func getAuthenticator(config *controllerConfig) (auth.Authenticator, error) {
	if config.noopAuthConfig != nil {
		return auth.NewNoopAuthenticator(config.noopAuthConfig), nil
	}
	if config.githubAuthConfig != nil {
		return auth.NewGithubAuthenticator(config.githubAuthConfig), nil
	}
	return nil, fmt.Errorf("authentication method not configured")
}

func httpError(c *gin.Context, status int) {
	c.AbortWithStatus(status)
}

func NewClientConfig(conf *controllerConfig) *ClientConfig {
	config := &ClientConfig{}

	if conf.githubAuthConfig != nil {
		if conf.githubAuthConfig.EnterpriseURL != "" {
			config.AccessManagementURL = conf.githubAuthConfig.EnterpriseURL + "/settings/apps/authorizations"
		} else {
			config.AccessManagementURL = GithubAccessManagementURL
		}
	}
	config.NebraskaVersion = version.Version
	config.Title = *appTitle
	config.HeaderStyle = *appHeaderStyle
	if *appLogoPath != "" {
		svg, err := ioutil.ReadFile(*appLogoPath)
		if err != nil {
			logger.Error().Err(err).Msg("Reading svg from path in config")
			return nil
		}
		if err := xml.Unmarshal(svg, &struct{}{}); err != nil {
			logger.Error().Err(err).Msg("Invalid format for SVG")
			return nil
		}
		config.Logo = string(svg)
		return config
	}
	return config
}

// ----------------------------------------------------------------------------
// Authentication
//

// authenticate is a middleware handler in charge of authenticating requests.
func (ctl *controller) authenticate(c *gin.Context) {
	teamID, replied := ctl.auth.Authenticate(c)
	if replied {
		return
	}
	logger.Debug().Str("setting team id in context keys", teamID).Msg("authenticate")
	c.Set("team_id", teamID)
	c.Next()
}

// ----------------------------------------------------------------------------
// API: users
//

func (ctl *controller) updateUserPassword(c *gin.Context) {
}

// ----------------------------------------------------------------------------
// API: applications CRUD
//

func (ctl *controller) addApp(c *gin.Context) {
	sourceAppID := c.Request.URL.Query().Get("clone_from")

	app := &api.Application{}
	if err := json.NewDecoder(c.Request.Body).Decode(app); err != nil {
		logger.Error().Err(err).Msg("addApp - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}
	app.TeamID = c.GetString("team_id")

	_, err := ctl.api.AddAppCloning(app, sourceAppID)
	if err != nil {
		logger.Error().Err(err).Str("sourceAppID", sourceAppID).Msgf("addApp - cloning app %v", app)
		httpError(c, http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error().Err(err).Str("appID", app.ID).Msgf("addApp - getting added app")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(app); err != nil {
		logger.Error().Err(err).Msgf("addApp - encoding app %v", app)
	}
}

func (ctl *controller) updateApp(c *gin.Context) {
	app := &api.Application{}
	if err := json.NewDecoder(c.Request.Body).Decode(app); err != nil {
		logger.Error().Err(err).Msg("updateApp - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}
	app.ID = c.Params.ByName("app_id")
	app.TeamID = c.GetString("team_id")

	err := ctl.api.UpdateApp(app)
	if err != nil {
		logger.Error().Err(err).Msgf("updatedApp - updating app %v", app)
		httpError(c, http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error().Err(err).Str("appID", app.ID).Msg("updateApp - getting updated app")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(app); err != nil {
		logger.Error().Err(err).Str("appID", app.ID).Msg("updateApp - encoding app")
	}
}

func (ctl *controller) deleteApp(c *gin.Context) {
	appID := c.Params.ByName("app_id")

	err := ctl.api.DeleteApp(appID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error().Err(err).Str("appID", appID).Msg("deleteApp")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getApp(c *gin.Context) {
	appID := c.Params.ByName("app_id")

	app, err := ctl.api.GetApp(appID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(app); err != nil {
			logger.Error().Err(err).Str("appID", appID).Msg("getApp - encoding app")
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("appID", appID).Msg("getApp - getting app")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getApps(c *gin.Context) {
	teamID := c.GetString("team_id")
	page, _ := strconv.ParseUint(c.Query("page"), 10, 64)
	perPage, _ := strconv.ParseUint(c.Query("perpage"), 10, 64)

	apps, err := ctl.api.GetApps(teamID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(apps); err != nil {
			logger.Error().Err(err).Str("teamID", teamID).Msgf("getApps - encoding apps")
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("teamID", teamID).Msg("getApps - getting apps")
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: groups CRUD
//

func (ctl *controller) addGroup(c *gin.Context) {
	group := &api.Group{}
	if err := json.NewDecoder(c.Request.Body).Decode(group); err != nil {
		logger.Error().Err(err).Msg("addGroup - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}
	group.ApplicationID = c.Params.ByName("app_id")

	_, err := ctl.api.AddGroup(group)
	if err != nil {
		logger.Error().Err(err).Msgf("addGroup - adding group %v", group)
		httpError(c, http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error().Err(err).Str("groupID", group.ID).Msgf("addGroup - getting added group")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(group); err != nil {
		logger.Error().Err(err).Msgf("addGroup - encoding group %v", group)
	}
}

func (ctl *controller) updateGroup(c *gin.Context) {
	group := &api.Group{}
	if err := json.NewDecoder(c.Request.Body).Decode(group); err != nil {
		logger.Error().Err(err).Msg("updateGroup - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}
	group.ID = c.Params.ByName("group_id")
	group.ApplicationID = c.Params.ByName("app_id")

	err := ctl.api.UpdateGroup(group)
	if err != nil {
		logger.Error().Err(err).Msgf("updateGroup - updating group %v", group)
		httpError(c, http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error().Err(err).Str("groupID", group.ID).Msg("updateGroup - fetching updated group")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(group); err != nil {
		logger.Error().Err(err).Msgf("updateGroup - encoding group %v", group)
	}
}

func (ctl *controller) deleteGroup(c *gin.Context) {
	groupID := c.Params.ByName("group_id")

	err := ctl.api.DeleteGroup(groupID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error().Err(err).Str("groupID", groupID).Msgf("deleteGroup")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroup(c *gin.Context) {
	groupID := c.Params.ByName("group_id")

	group, err := ctl.api.GetGroup(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(group); err != nil {
			logger.Error().Err(err).Msgf("getGroup - encoding group %v", group)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("groupID", groupID).Msg("getGroup - getting group")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroups(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	page, _ := strconv.ParseUint(c.Query("page"), 10, 64)
	perPage, _ := strconv.ParseUint(c.Query("perpage"), 10, 64)

	groups, err := ctl.api.GetGroups(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(groups); err != nil {
			logger.Error().Err(err).Str("appID", appID).Msg("getGroups - encoding groups")
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("appID", appID).Msgf("getGroups - getting groups")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroupVersionCountTimeline(c *gin.Context) {
	groupID := c.Params.ByName("group_id")
	duration := c.Query("duration")
	versionCountTimeline, err := ctl.api.GetGroupVersionCountTimeline(groupID, duration)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(versionCountTimeline); err != nil {
			logger.Error().Err(err).Msgf("getGroupVersionCountTimeline - encoding group count-timeline %v", versionCountTimeline)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("groupID", groupID).Msgf("getGroupVersionCountTimeline - getting version timeline")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroupStatusCountTimeline(c *gin.Context) {
	groupID := c.Params.ByName("group_id")
	duration := c.Query("duration")
	statusCountTimeline, err := ctl.api.GetGroupStatusCountTimeline(groupID, duration)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(statusCountTimeline); err != nil {
			logger.Error().Err(err).Msgf("getGroupStatusCountTimeline - encoding group count-timeline %v", statusCountTimeline)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("groupID", groupID).Msgf("getGroupStatusCountTimeline - getting status timeline")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroupInstancesStats(c *gin.Context) {
	groupID := c.Params.ByName("group_id")
	duration := c.Query("duration")
	instancesStats, err := ctl.api.GetGroupInstancesStats(groupID, duration)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(instancesStats); err != nil {
			logger.Error().Err(err).Msgf("getGroupInstancesStats - encoding group instancesStats %v", instancesStats)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("groupID", groupID).Msgf("getGroupInstancesStats - getting instances stats groupID")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroupVersionBreakdown(c *gin.Context) {
	groupID := c.Params.ByName("group_id")

	versionBreakdown, err := ctl.api.GetGroupVersionBreakdown(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(versionBreakdown); err != nil {
			logger.Error().Err(err).Msgf("getVersionBreakdown - encoding group version_breakdown %v", versionBreakdown)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("groupID", groupID).Msg("getVersionBreakdown - getting version breakdown")
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: channels CRUD
//

func (ctl *controller) addChannel(c *gin.Context) {
	channel := &api.Channel{}
	if err := json.NewDecoder(c.Request.Body).Decode(channel); err != nil {
		logger.Error().Err(err).Msgf("addChannel")
		httpError(c, http.StatusBadRequest)
		return
	}
	channel.ApplicationID = c.Params.ByName("app_id")

	_, err := ctl.api.AddChannel(channel)
	if err != nil {
		logger.Error().Err(err).Msgf("addChannel channel %v", channel)
		httpError(c, http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error().Err(err).Str("channelID", channel.ID).Msg("addChannel")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(channel); err != nil {
		logger.Error().Err(err).Str("channelID", channel.ID).Msg("addChannel - encoding channel")
	}
}

func (ctl *controller) updateChannel(c *gin.Context) {
	channel := &api.Channel{}
	if err := json.NewDecoder(c.Request.Body).Decode(channel); err != nil {
		logger.Error().Err(err).Msg("updateChannel - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}
	channel.ID = c.Params.ByName("channel_id")
	channel.ApplicationID = c.Params.ByName("app_id")

	err := ctl.api.UpdateChannel(channel)
	if err != nil {
		logger.Error().Err(err).Msgf("updateChannel - updating channel %v", channel)
		httpError(c, http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error().Err(err).Str("channelID", channel.ID).Msg("updateChannel - getting channel updated")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(channel); err != nil {
		logger.Error().Err(err).Str("channelID", channel.ID).Msgf("updateChannel - encoding channel")
	}
}

func (ctl *controller) deleteChannel(c *gin.Context) {
	channelID := c.Params.ByName("channel_id")

	err := ctl.api.DeleteChannel(channelID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error().Err(err).Str("channelID", channelID).Msg("deleteChannel")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getChannel(c *gin.Context) {
	channelID := c.Params.ByName("channel_id")

	channel, err := ctl.api.GetChannel(channelID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(channel); err != nil {
			logger.Error().Err(err).Str("channelID", channel.ID).Msg("getChannel - encoding channel")
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("channelID", channel.ID).Msg("getChannel - getting updated channel")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getChannels(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	page, _ := strconv.ParseUint(c.Query("page"), 10, 64)
	perPage, _ := strconv.ParseUint(c.Query("perpage"), 10, 64)

	channels, err := ctl.api.GetChannels(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(channels); err != nil {
			logger.Error().Err(err).Str("appID", appID).Msg("getChannels - encoding channel")
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("appID", appID).Msg("getChannels - getting channels")
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: packages CRUD
//

func (ctl *controller) addPackage(c *gin.Context) {
	pkg := &api.Package{}
	if err := json.NewDecoder(c.Request.Body).Decode(pkg); err != nil {
		logger.Error().Err(err).Msg("addPackage - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}
	pkg.ApplicationID = c.Params.ByName("app_id")

	_, err := ctl.api.AddPackage(pkg)
	if err != nil {
		logger.Error().Err(err).Msgf("addPackage - adding package %v", pkg)
		httpError(c, http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error().Err(err).Str("packageID", pkg.ID).Msg("addPackage - getting added package")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(pkg); err != nil {
		logger.Error().Err(err).Str("packageID", pkg.ID).Msgf("addPackage - encoding package")
	}
}

func (ctl *controller) updatePackage(c *gin.Context) {
	pkg := &api.Package{}
	if err := json.NewDecoder(c.Request.Body).Decode(pkg); err != nil {
		logger.Error().Err(err).Msg("updatePackage - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}
	pkg.ID = c.Params.ByName("package_id")
	pkg.ApplicationID = c.Params.ByName("app_id")

	err := ctl.api.UpdatePackage(pkg)
	if err != nil {
		logger.Error().Err(err).Msgf("updatePackage - updating package %v", pkg)
		httpError(c, http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error().Err(err).Str("packageID", pkg.ID).Msg("addPackage - getting updated package")
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(pkg); err != nil {
		logger.Error().Err(err).Str("packageID", pkg.ID).Msg("updatePackage - encoding package")
	}
}

func (ctl *controller) deletePackage(c *gin.Context) {
	packageID := c.Params.ByName("package_id")

	err := ctl.api.DeletePackage(packageID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error().Err(err).Str("packageID", packageID).Msgf("deletePackage")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getPackage(c *gin.Context) {
	packageID := c.Params.ByName("package_id")

	pkg, err := ctl.api.GetPackage(packageID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(pkg); err != nil {
			logger.Error().Err(err).Str("packageID", packageID).Msg("getPackage - encoding package")
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("packageID", packageID).Msgf("getPackage - getting package")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getPackages(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	page, _ := strconv.ParseUint(c.Query("page"), 10, 64)
	perPage, _ := strconv.ParseUint(c.Query("perpage"), 10, 64)

	pkgs, err := ctl.api.GetPackages(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(pkgs); err != nil {
			logger.Error().Err(err).Str("appID", appID).Msg("getPackages - encoding packages")
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("appID", appID).Msg("getPackages - getting packages")
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: instances
//

func (ctl *controller) getInstanceStatusHistory(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	groupID := c.Params.ByName("group_id")
	instanceID := c.Params.ByName("instance_id")
	limit, _ := strconv.ParseUint(c.Query("limit"), 10, 64)

	instanceStatusHistory, err := ctl.api.GetInstanceStatusHistory(instanceID, appID, groupID, limit)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(instanceStatusHistory); err != nil {
			logger.Error().Err(err).Str("appID", appID).Str("groupID", groupID).Str("instanceID", instanceID).Msgf("getInstanceStatusHistory - encoding status history limit %d", limit)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("appID", appID).Str("groupID", groupID).Str("instanceID", instanceID).Msgf("getInstanceStatusHistory - getting status history limit %d", limit)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getInstances(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	groupID := c.Params.ByName("group_id")

	p := api.InstancesQueryParams{
		ApplicationID: appID,
		GroupID:       groupID,
		Version:       c.Query("version"),
	}
	p.Status, _ = strconv.Atoi(c.Query("status"))
	p.Page, _ = strconv.ParseUint(c.Query("page"), 10, 64)
	p.PerPage, _ = strconv.ParseUint(c.Query("perpage"), 10, 64)
	duration := c.Query("duration")
	result, err := ctl.api.GetInstances(p, duration)
	if err == nil {
		if err := json.NewEncoder(c.Writer).Encode(result); err != nil {
			logger.Error().Err(err).Msgf("getInstances - encoding instances params %v", p)
		}
	} else {
		logger.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getInstancesCount(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	groupID := c.Params.ByName("group_id")

	p := api.InstancesQueryParams{
		ApplicationID: appID,
		GroupID:       groupID,
	}
	duration := c.Query("duration")
	result, err := ctl.api.GetInstancesCount(p, duration)
	if err == nil {
		if err := json.NewEncoder(c.Writer).Encode(result); err != nil {
			logger.Error().Err(err).Msgf("getInstances - encoding instances params %v", p)
		}
	} else {
		logger.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getInstance(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	instanceID := c.Params.ByName("instance_id")
	result, err := ctl.api.GetInstance(instanceID, appID)
	if err == nil {
		if err := json.NewEncoder(c.Writer).Encode(result); err != nil {
			logger.Error().Err(err).Str("appID", appID).Str("instanceID", instanceID).Msg("getInstance - encoding instance")
		}
	} else {
		logger.Error().Err(err).Str("appID", appID).Str("instanceID", instanceID).Msg("getInstance - getting instance")
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) updateInstance(c *gin.Context) {
	instanceID := c.Params.ByName("instance_id")
	params := struct{ Alias string }{}

	if err := json.NewDecoder(c.Request.Body).Decode(&params); err != nil {
		logger.Error().Err(err).Msg("updateInstance - decoding payload")
		httpError(c, http.StatusBadRequest)
		return
	}

	instance, err := ctl.api.UpdateInstance(instanceID, params.Alias)
	if err != nil {
		logger.Error().Err(err).Str("instance", instanceID).Msgf("updateInstance - updating params %s", params)
		httpError(c, http.StatusBadRequest)
		return
	}

	if err == nil {
		if err := json.NewEncoder(c.Writer).Encode(instance); err != nil {
			logger.Error().Err(err).Str("instance", instanceID).Msgf("updateInstance - encoding params %s", params)
		}
	} else {
		logger.Error().Err(err).Str("instance", instanceID).Msgf("updateInstance - getting instance %s params", params)
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: activity
//

func (ctl *controller) getActivity(c *gin.Context) {
	teamID := c.GetString("team_id")

	p := api.ActivityQueryParams{
		AppID:      c.Query("app"),
		GroupID:    c.Query("group"),
		ChannelID:  c.Query("channel"),
		InstanceID: c.Query("instance"),
		Version:    c.Query("version"),
	}
	p.Severity, _ = strconv.Atoi(c.Query("severity"))
	p.Start, _ = time.Parse(time.RFC3339, c.Query("start"))
	p.End, _ = time.Parse(time.RFC3339, c.Query("end"))
	p.Page, _ = strconv.ParseUint(c.Query("page"), 10, 64)
	p.PerPage, _ = strconv.ParseUint(c.Query("perpage"), 10, 64)

	activityEntries, err := ctl.api.GetActivity(teamID, p)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(activityEntries); err != nil {
			logger.Error().Err(err).Msgf("getActivity - encoding activity entries params %v", p)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error().Err(err).Str("teamID", teamID).Msgf("getActivity params %v", p)
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// OMAHA server
//

func (ctl *controller) processOmahaRequest(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/xml")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, UpdateMaxRequestSize)
	if err := ctl.omahaHandler.Handle(c.Request.Body, c.Writer, getRequestIP(c.Request)); err != nil {
		logger.Error().Err(err).Msgf("process omaha request")
		if uerr := errors.Unwrap(err); uerr != nil && uerr.Error() == "http: request body too large" {
			httpError(c, http.StatusBadRequest)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers
//

func getRequestIP(r *http.Request) string {
	ips := strings.Split(r.Header.Get("X-FORWARDED-FOR"), ",")
	if ips[0] != "" && net.ParseIP(strings.TrimSpace(ips[0])) != nil {
		return ips[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// ----------------------------------------------------------------------------
// Config
//

func (ctl *controller) getConfig(c *gin.Context) {
	if err := json.NewEncoder(c.Writer).Encode(ctl.clientConfig); err != nil {
		logger.Error().Err(err).Msgf("getConfig - encoding config")
		httpError(c, http.StatusBadRequest)
	}
}
