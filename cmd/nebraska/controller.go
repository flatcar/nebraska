package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
)

// ClientConfig represents Nebraska's configuration of interest for the client.
type ClientConfig struct {
	AccessManagementURL string `json:"access_management_url"`
	NebraskaVersion     string `json:"nebraska_version"`
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
			API:          conf.api,
			HostPackages: conf.hostFlatcarPackages,
			PackagesPath: conf.flatcarPackagesPath,
			PackagesURL:  conf.nebraskaURL + "/flatcar/",
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
	logger.Debug("authenticate", "setting team id in context keys", teamID)
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
		logger.Error("addApp - decoding payload", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	app.TeamID = c.GetString("team_id")

	_, err := ctl.api.AddAppCloning(app, sourceAppID)
	if err != nil {
		logger.Error("addApp - cloning app", "error", err.Error(), "app", app, "sourceAppID", sourceAppID)
		httpError(c, http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("addApp - getting added app", "error", err.Error(), "appID", app.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(app); err != nil {
		logger.Error("addApp - encoding app", "error", err.Error(), "app", app)
	}
}

func (ctl *controller) updateApp(c *gin.Context) {
	app := &api.Application{}
	if err := json.NewDecoder(c.Request.Body).Decode(app); err != nil {
		logger.Error("updateApp - decoding payload", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	app.ID = c.Params.ByName("app_id")
	app.TeamID = c.GetString("team_id")

	err := ctl.api.UpdateApp(app)
	if err != nil {
		logger.Error("updatedApp - updating app", "error", err.Error(), "app", app)
		httpError(c, http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("updateApp - getting updated app", "error", err.Error(), "appID", app.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(app); err != nil {
		logger.Error("updateApp - encoding app", "error", err.Error(), "appID", app.ID)
	}
}

func (ctl *controller) deleteApp(c *gin.Context) {
	appID := c.Params.ByName("app_id")

	err := ctl.api.DeleteApp(appID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error("deleteApp", "error", err.Error(), "appID", appID)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getApp(c *gin.Context) {
	appID := c.Params.ByName("app_id")

	app, err := ctl.api.GetApp(appID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(app); err != nil {
			logger.Error("getApp - encoding app", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getApp - getting app", "error", err.Error(), "appID", appID)
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
			logger.Error("getApps - encoding apps", "error", err.Error(), "teamID", teamID)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getApps - getting apps", "error", err.Error(), "teamID", teamID)
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: groups CRUD
//

func (ctl *controller) addGroup(c *gin.Context) {
	group := &api.Group{}
	if err := json.NewDecoder(c.Request.Body).Decode(group); err != nil {
		logger.Error("addGroup - decoding payload", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	group.ApplicationID = c.Params.ByName("app_id")

	_, err := ctl.api.AddGroup(group)
	if err != nil {
		logger.Error("addGroup - adding group", "error", err.Error(), "group", group)
		httpError(c, http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("addGroup - getting added group", "error", err.Error(), "groupID", group.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(group); err != nil {
		logger.Error("addGroup - encoding group", "error", err.Error(), "group", group)
	}
}

func (ctl *controller) updateGroup(c *gin.Context) {
	group := &api.Group{}
	if err := json.NewDecoder(c.Request.Body).Decode(group); err != nil {
		logger.Error("updateGroup - decoding payload", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	group.ID = c.Params.ByName("group_id")
	group.ApplicationID = c.Params.ByName("app_id")

	err := ctl.api.UpdateGroup(group)
	if err != nil {
		logger.Error("updateGroup - updating group", "error", err.Error(), "group", group)
		httpError(c, http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("updateGroup - fetching updated group", "error", err.Error(), "groupID", group.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(group); err != nil {
		logger.Error("updateGroup - encoding group", "error", err.Error(), "group", group)
	}
}

func (ctl *controller) deleteGroup(c *gin.Context) {
	groupID := c.Params.ByName("group_id")

	err := ctl.api.DeleteGroup(groupID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error("deleteGroup", "error", err.Error(), "groupID", groupID)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroup(c *gin.Context) {
	groupID := c.Params.ByName("group_id")

	group, err := ctl.api.GetGroup(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(group); err != nil {
			logger.Error("getGroup - encoding group", "error", err.Error(), "group", group)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getGroup - getting group", "error", err.Error(), "groupID", groupID)
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
			logger.Error("getGroups - encoding groups", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getGroups - getting groups", "error", err.Error(), "appID", appID)
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
			logger.Error("getGroupVersionCountTimeline - encoding group", "error", err.Error(), "count-timeline", versionCountTimeline)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getGroupVersionCountTimeline - getting version timeline", "error", err.Error(), "groupID", groupID)
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
			logger.Error("getGroupStatusCountTimeline - encoding group", "error", err.Error(), "count-timeline", statusCountTimeline)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getGroupStatusCountTimeline - getting status timeline", "error", err.Error(), "groupID", groupID)
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
			logger.Error("getGroupInstancesStats - encoding group", "error", err.Error(), "instancesStats", instancesStats)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getGroupInstancesStats - getting instances stats", "error", err.Error(), "groupID", groupID)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroupVersionBreakdown(c *gin.Context) {
	groupID := c.Params.ByName("group_id")

	versionBreakdown, err := ctl.api.GetGroupVersionBreakdown(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(versionBreakdown); err != nil {
			logger.Error("getVersionBreakdown - encoding group", "error", err.Error(), "version_breakdown", versionBreakdown)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getVersionBreakdown - getting version breakdown", "error", err.Error(), "groupID", groupID)
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: channels CRUD
//

func (ctl *controller) addChannel(c *gin.Context) {
	channel := &api.Channel{}
	if err := json.NewDecoder(c.Request.Body).Decode(channel); err != nil {
		logger.Error("addChannel", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	channel.ApplicationID = c.Params.ByName("app_id")

	_, err := ctl.api.AddChannel(channel)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channel", channel)
		httpError(c, http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channelID", channel.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(channel); err != nil {
		logger.Error("addChannel - encoding channel", "error", err.Error(), "channelID", channel.ID)
	}
}

func (ctl *controller) updateChannel(c *gin.Context) {
	channel := &api.Channel{}
	if err := json.NewDecoder(c.Request.Body).Decode(channel); err != nil {
		logger.Error("updateChannel - decoding payload", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	channel.ID = c.Params.ByName("channel_id")
	channel.ApplicationID = c.Params.ByName("app_id")

	err := ctl.api.UpdateChannel(channel)
	if err != nil {
		logger.Error("updateChannel - updating channel", "error", err.Error(), "channel", channel)
		httpError(c, http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("updateChannel - getting channel updated", "error", err.Error(), "channelID", channel.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(channel); err != nil {
		logger.Error("updateChannel - encoding channel", "error", err.Error(), "channelID", channel.ID)
	}
}

func (ctl *controller) deleteChannel(c *gin.Context) {
	channelID := c.Params.ByName("channel_id")

	err := ctl.api.DeleteChannel(channelID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error("deleteChannel", "error", err.Error(), "channelID", channelID)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getChannel(c *gin.Context) {
	channelID := c.Params.ByName("channel_id")

	channel, err := ctl.api.GetChannel(channelID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(channel); err != nil {
			logger.Error("getChannel - encoding channel", "error", err.Error(), "channelID", channelID)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getChannel - getting updated channel", "error", err.Error(), "channelID", channelID)
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
			logger.Error("getChannels - encoding channel", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getChannels - getting channels", "error", err.Error(), "appID", appID)
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: packages CRUD
//

func (ctl *controller) addPackage(c *gin.Context) {
	pkg := &api.Package{}
	if err := json.NewDecoder(c.Request.Body).Decode(pkg); err != nil {
		logger.Error("addPackage - decoding payload", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	pkg.ApplicationID = c.Params.ByName("app_id")

	_, err := ctl.api.AddPackage(pkg)
	if err != nil {
		logger.Error("addPackage - adding package", "error", err.Error(), "package", pkg)
		httpError(c, http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting added package", "error", err.Error(), "packageID", pkg.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(pkg); err != nil {
		logger.Error("addPackage - encoding package", "error", err.Error(), "packageID", pkg.ID)
	}
}

func (ctl *controller) updatePackage(c *gin.Context) {
	pkg := &api.Package{}
	if err := json.NewDecoder(c.Request.Body).Decode(pkg); err != nil {
		logger.Error("updatePackage - decoding payload", "error", err.Error())
		httpError(c, http.StatusBadRequest)
		return
	}
	pkg.ID = c.Params.ByName("package_id")
	pkg.ApplicationID = c.Params.ByName("app_id")

	err := ctl.api.UpdatePackage(pkg)
	if err != nil {
		logger.Error("updatePackage - updating package", "error", err.Error(), "package", pkg)
		httpError(c, http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting updated package", "error", err.Error(), "packageID", pkg.ID)
		httpError(c, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(c.Writer).Encode(pkg); err != nil {
		logger.Error("updatePackage - encoding package", "error", err.Error(), "packageID", pkg.ID)
	}
}

func (ctl *controller) deletePackage(c *gin.Context) {
	packageID := c.Params.ByName("package_id")

	err := ctl.api.DeletePackage(packageID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		logger.Error("deletePackage", "error", err.Error(), "packageID", packageID)
		httpError(c, http.StatusBadRequest)
	}
}

func (ctl *controller) getPackage(c *gin.Context) {
	packageID := c.Params.ByName("package_id")

	pkg, err := ctl.api.GetPackage(packageID)
	switch err {
	case nil:
		if err := json.NewEncoder(c.Writer).Encode(pkg); err != nil {
			logger.Error("getPackage - encoding package", "error", err.Error(), "packageID", packageID)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getPackage - getting package", "error", err.Error(), "packageID", packageID)
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
			logger.Error("getPackages - encoding packages", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getPackages - getting packages", "error", err.Error(), "appID", appID)
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
			logger.Error("getInstanceStatusHistory - encoding status history", "error", err.Error(), "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getInstanceStatusHistory - getting status history", "error", err.Error(), "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
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
			logger.Error("getInstances - encoding instances", "error", err.Error(), "params", p)
		}
	} else {
		logger.Error("getInstances - getting instances", "error", err.Error(), "params", p)
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
			logger.Error("getInstances - encoding instances", "error", err.Error(), "params", p)
		}
	} else {
		logger.Error("getInstances - getting instances", "error", err.Error(), "params", p)
		httpError(c, http.StatusBadRequest)
	}
}
func (ctl *controller) getInstance(c *gin.Context) {
	appID := c.Params.ByName("app_id")
	instanceID := c.Params.ByName("instance_id")
	result, err := ctl.api.GetInstance(instanceID, appID)
	if err == nil {
		if err := json.NewEncoder(c.Writer).Encode(result); err != nil {
			logger.Error("getInstance - encoding instance", "error", err.Error(), "appID", appID, "instanceID", instanceID)
		}
	} else {
		logger.Error("getInstance - getting instance", "error", err.Error(), "appID", appID, "instanceID", instanceID)
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
			logger.Error("getActivity - encoding activity entries", "error", err.Error(), "params", p)
		}
	case sql.ErrNoRows:
		httpError(c, http.StatusNotFound)
	default:
		logger.Error("getActivity", "error", err, "teamID", teamID, "params", p)
		httpError(c, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// Metrics
//

const (
	appInstancesPerChannelMetricsProlog = `# HELP nebraska_application_instances_per_channel A number of applications from specific channel running on instances
# TYPE nebraska_application_instances_per_channel gauge`
	failedUpdatesMetricsProlog = `# HELP nebraska_failed_updates A number of failed updates of an application
# TYPE nebraska_failed_updates gauge`
)

func escapeMetricString(str string) string {
	str = strings.Replace(str, `\`, `\\`, -1)
	str = strings.Replace(str, `"`, `\"`, -1)
	str = strings.Replace(str, "\n", `\n`, -1)
	return str
}

func (ctl *controller) getMetrics(c *gin.Context) {
	teamID := c.GetString("team_id")

	nowUnixMillis := time.Now().Unix() * 1000
	aipcMetrics, err := ctl.api.GetAppInstancesPerChannelMetrics(teamID)
	if err != nil {
		logger.Error("getMetrics - getting app instances per channel metrics", "error", err.Error(), "teamID", teamID)
		httpError(c, http.StatusBadRequest)
		return
	}
	fuMetrics, err := ctl.api.GetFailedUpdatesMetrics(teamID)
	if err != nil {
		logger.Error("getMetrics - getting failed updates metrics", "error", err.Error(), "teamID", teamID)
		httpError(c, http.StatusBadRequest)
		return
	}

	// "version" specifies a version of prometheus text file
	// format. For details see:
	//
	// https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md#basic-info
	c.Writer.Header().Set("Content-Type", "text/plain; version=0.0.4")
	c.Writer.WriteHeader(http.StatusOK)
	needEmptyLine := false
	if len(aipcMetrics) > 0 {
		if needEmptyLine {
			fmt.Fprintf(c.Writer, "\n")
		}
		fmt.Fprintf(c.Writer, "%s\n", appInstancesPerChannelMetricsProlog)
		for _, metric := range aipcMetrics {
			fmt.Fprintf(c.Writer, `nebraska_application_instances_per_channel{application="%s",version="%s",channel="%s"} %d %d%s`, escapeMetricString(metric.ApplicationName), escapeMetricString(metric.Version), escapeMetricString(metric.ChannelName), metric.InstancesCount, nowUnixMillis, "\n")
		}
		needEmptyLine = true
	}
	if len(fuMetrics) > 0 {
		if needEmptyLine {
			fmt.Fprintf(c.Writer, "\n")
		}
		fmt.Fprintf(c.Writer, "%s\n", failedUpdatesMetricsProlog)
		for _, metric := range fuMetrics {
			fmt.Fprintf(c.Writer, `nebraska_failed_updates{application="%s"} %d %d%s`, escapeMetricString(metric.ApplicationName), metric.FailureCount, nowUnixMillis, "\n")
		}
	}
}

// ----------------------------------------------------------------------------
// OMAHA server
//

func (ctl *controller) processOmahaRequest(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/xml")
	if err := ctl.omahaHandler.Handle(c.Request.Body, c.Writer, getRequestIP(c.Request)); err != nil {
		logger.Error("process omaha request", "error", err)
	}
}

// ----------------------------------------------------------------------------
// Helpers
//

func getRequestIP(r *http.Request) string {
	ips := strings.Split(r.Header.Get("X-FORWARDED-FOR"), ", ")
	if ips[0] != "" && net.ParseIP(ips[0]) != nil {
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
		logger.Error("getConfig - encoding config", "error", err.Error())
		httpError(c, http.StatusBadRequest)
	}
}
