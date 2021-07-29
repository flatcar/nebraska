package handler

import (
	"database/sql"
	"net/http"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"
)

func (h *handler) PaginateGroups(ctx echo.Context, appId string, params codegen.PaginateGroupsParams) error {

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	totalCount, err := h.db.GetGroupsCount(appId)
	if err != nil {
		logger.Error().Err(err).Str("appID", appId).Msg("getGroups count - getting groups")
		return ctx.NoContent(http.StatusInternalServerError)

	}

	groups, err := h.db.GetGroups(appId, *params.Page, *params.Perpage)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error().Err(err).Msg("getGroups - getting groups not found error")
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appId).Msg("getGroups - getting groups")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, groupsPage{totalCount, len(groups), groups})
}

func (h *handler) CreateGroup(ctx echo.Context, appId string) error {

	logger := loggerWithUsername(logger, ctx)

	var request codegen.CreateGroupInfo
	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("addGroup - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	group := groupFromRequest(request.Name, request.Description, request.PolicyMaxUpdatesPerPeriod, request.PolicyOfficeHours, request.PolicyPeriodInterval, request.PolicySafeMode, request.PolicyTimezone, request.PolicyUpdateTimeout, request.PolicyUpdatesEnabled, request.ChannelId, request.Track, "", appId)

	group, err = h.db.AddGroup(group)
	if err != nil {
		logger.Error().Err(err).Msgf("addGroup - adding group %v", group)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group, err = h.db.GetGroup(group.ID)
	if err != nil {
		logger.Error().Err(err).Msgf("addGroup - adding group %v", group)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	logger.Info().Msgf("addGroup - successfully added group %+v", group)

	return ctx.JSON(http.StatusOK, group)
}

func (h *handler) GetGroup(ctx echo.Context, appId string, groupId string) error {

	group, err := h.db.GetGroup(groupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("groupID", groupId).Msg("getGroup - getting group")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, group)
}

func (h *handler) UpdateGroup(ctx echo.Context, appId string, groupId string) error {
	logger := loggerWithUsername(logger, ctx)

	var request codegen.UpdateGroupInfo
	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("updateGroup - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	oldGroup, err := h.db.GetGroup(groupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("groupID", groupId).Msg("updateGroup - getting old group to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group := groupFromRequest(request.Name, request.Description, request.PolicyMaxUpdatesPerPeriod, request.PolicyOfficeHours, request.PolicyPeriodInterval, request.PolicySafeMode, request.PolicyTimezone, request.PolicyUpdateTimeout, request.PolicyUpdatesEnabled, request.ChannelId, request.Track, groupId, appId)

	err = h.db.UpdateGroup(group)
	if err != nil {
		logger.Error().Err(err).Msgf("updateGroup - updating group %+v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group, err = h.db.GetGroup(groupId)
	if err != nil {
		logger.Error().Err(err).Str("groupID", groupId).Msg("getGroup - getting group")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("updateGroup - successfully updated group %+v -> %+v", oldGroup, group)

	return ctx.JSON(http.StatusOK, group)

}

func (h *handler) DeleteGroup(ctx echo.Context, appId string, groupId string) error {

	logger := loggerWithUsername(logger, ctx)

	group, err := h.db.GetGroup(groupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("groupID", groupId).Msg("updateGroup - getting old group to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.DeleteGroup(groupId)
	if err != nil {
		logger.Error().Err(err).Str("groupID", groupId).Msg("deleteGroup")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("deleteGroup - successfully deleted group %+v", group)

	return ctx.NoContent(http.StatusOK)
}

func (h *handler) GetGroupVersionTimeline(ctx echo.Context, appId string, groupId string, params codegen.GetGroupVersionTimelineParams) error {

	versionCountTimeline, isCache, err := h.db.GetGroupVersionCountTimeline(groupId, params.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("groupID", groupId).Msg("getGroupVersionCountTimeline - getting version timeline")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if isCache {
		ctx.Response().Header().Set("X-Cache", "HIT")
	} else {
		ctx.Response().Header().Set("X-Cache", "MISS")
	}

	return ctx.JSON(http.StatusOK, versionCountTimeline)

}

func (h *handler) GetGroupStatusTimeline(ctx echo.Context, appId string, groupId string, params codegen.GetGroupStatusTimelineParams) error {

	statusCountTimeline, err := h.db.GetGroupStatusCountTimeline(groupId, params.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("groupID", groupId).Msg("getGroupStatusCountTimeline - getting status timeline")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, statusCountTimeline)
}

func (h *handler) GetGroupInstanceStats(ctx echo.Context, appId string, groupId string, params codegen.GetGroupInstanceStatsParams) error {

	instancesStats, err := h.db.GetGroupInstancesStats(groupId, params.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("groupID", groupId).Msg("getGroupInstancesStats - getting instances stats groupID")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, instancesStats)
}

func (h *handler) GetGroupVersionBreakdown(ctx echo.Context, appId string, groupId string) error {

	versionBreakdown, err := h.db.GetGroupVersionBreakdown(groupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("groupID", groupId).Msg("getVersionBreakdown - getting version breakdown")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, versionBreakdown)
}

func (h *handler) GetGroupInstances(ctx echo.Context, appId string, groupId string, params codegen.GetGroupInstancesParams) error {

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	p := api.InstancesQueryParams{
		ApplicationID: appId,
		GroupID:       groupId,
		Status:        params.Status,
		Page:          *params.Page,
		PerPage:       *params.Perpage,
	}
	if params.Version != nil {
		p.Version = *params.Version
	}
	if params.SortFilter != nil {
		p.SortFilter = *params.SortFilter
	}
	if params.SortOrder != nil {
		p.SortOrder = *params.SortOrder
	}

	groupInstances, err := h.db.GetInstances(p, params.Duration)
	if err != nil {
		logger.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, groupInstances)
}

func (h *handler) GetGroupInstancesCount(ctx echo.Context, appId string, groupId string, params codegen.GetGroupInstancesCountParams) error {

	p := api.InstancesQueryParams{
		ApplicationID: appId,
		GroupID:       groupId,
	}

	count, err := h.db.GetInstancesCount(p, params.Duration)
	if err != nil {
		logger.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, codegen.InstanceCount{Count: count})

}

func groupFromRequest(name string, description string, policyMaxUpdatesPerPeriod int, policyOfficeHours bool, policyPeriodInterval string, policySafeMode bool, policyTimezone string, policyUpdateTimeout string, policyUpdatesEnabled bool, channelID *string, track string, groupID string, appID string) *api.Group {

	group := &api.Group{
		Name:                      name,
		Description:               description,
		PolicyMaxUpdatesPerPeriod: policyMaxUpdatesPerPeriod,
		PolicyOfficeHours:         policyOfficeHours,
		PolicyPeriodInterval:      policyPeriodInterval,
		PolicySafeMode:            policySafeMode,
		PolicyTimezone:            null.StringFrom(policyTimezone),
		PolicyUpdateTimeout:       policyUpdateTimeout,
		PolicyUpdatesEnabled:      policyUpdatesEnabled,
		ChannelID:                 null.StringFromPtr(channelID),
		Track:                     track,
	}
	if groupID != "" {
		group.ID = groupID
	}
	if appID != "" {
		group.ApplicationID = appID
	}

	return group
}

type groupsPage struct {
	TotalCount int          `json:"totalCount"`
	Count      int          `json:"count"`
	Groups     []*api.Group `json:"groups"`
}
