package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

func (h *Handler) PaginateGroups(ctx echo.Context, appIDorProductID string, params codegen.PaginateGroupsParams) error {
	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	totalCount, err := h.db.GetGroupsCount(appID)
	if err != nil {
		l.Error().Err(err).Str("appID", appID).Msg("getGroups count - getting groups")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	groups, err := h.db.GetGroups(appID, uint64(*params.Page), uint64(*params.Perpage))
	if err != nil {
		if err == sql.ErrNoRows {
			l.Error().Err(err).Msg("getGroups - getting groups not found error")
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("appID", appID).Msg("getGroups - getting groups")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, groupsPage{totalCount, len(groups), groups})
}

func (h *Handler) CreateGroup(ctx echo.Context, appIDorProductID string) error {
	l := loggerWithUsername(l, ctx)

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.GroupConfig
	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("addGroup - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	group := groupFromRequest(request.Name, request.Description, request.PolicyMaxUpdatesPerPeriod, request.PolicyOfficeHours, request.PolicyPeriodInterval, request.PolicySafeMode, request.PolicyTimezone, request.PolicyUpdateTimeout, request.PolicyUpdatesEnabled, request.ChannelId, request.Track, "", appID)

	group, err = h.db.AddGroup(group)
	if err != nil {
		l.Error().Err(err).Msgf("addGroup - adding group %v", group)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group, err = h.db.GetGroup(group.ID)
	if err != nil {
		l.Error().Err(err).Msgf("addGroup - adding group %v", group)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	l.Info().Msgf("addGroup - successfully added group %+v", group)

	return ctx.JSON(http.StatusOK, group)
}

func (h *Handler) GetGroup(ctx echo.Context, _ string, groupID string) error {
	group, err := h.db.GetGroup(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("getGroup - getting group")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, group)
}

func (h *Handler) UpdateGroup(ctx echo.Context, appIDorProductID string, groupID string) error {
	l := loggerWithUsername(l, ctx)

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.GroupConfig
	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("updateGroup - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	oldGroup, err := h.db.GetGroup(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("updateGroup - getting old group to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group := groupFromRequest(request.Name, request.Description, request.PolicyMaxUpdatesPerPeriod, request.PolicyOfficeHours, request.PolicyPeriodInterval, request.PolicySafeMode, request.PolicyTimezone, request.PolicyUpdateTimeout, request.PolicyUpdatesEnabled, request.ChannelId, request.Track, groupID, appID)

	err = h.db.UpdateGroup(group)
	if err != nil {
		l.Error().Err(err).Msgf("updateGroup - updating group %+v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group, err = h.db.GetGroup(groupID)
	if err != nil {
		l.Error().Err(err).Str("groupID", groupID).Msg("getGroup - getting group")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("updateGroup - successfully updated group %+v -> %+v", oldGroup, group)

	return ctx.JSON(http.StatusOK, group)
}

func (h *Handler) DeleteGroup(ctx echo.Context, _ string, groupID string) error {
	l := loggerWithUsername(l, ctx)

	group, err := h.db.GetGroup(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("updateGroup - getting old group to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.DeleteGroup(groupID)
	if err != nil {
		l.Error().Err(err).Str("groupID", groupID).Msg("deleteGroup")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("deleteGroup - successfully deleted group %+v", group)

	return ctx.NoContent(http.StatusNoContent)
}

func (h *Handler) GetGroupVersionTimeline(ctx echo.Context, _ string, groupID string, params codegen.GetGroupVersionTimelineParams) error {
	versionCountTimeline, isCache, err := h.db.GetGroupVersionCountTimeline(groupID, params.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("getGroupVersionCountTimeline - getting version timeline")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if isCache {
		ctx.Response().Header().Set("X-Cache", "HIT")
	} else {
		ctx.Response().Header().Set("X-Cache", "MISS")
	}

	return ctx.JSON(http.StatusOK, versionCountTimeline)
}

func (h *Handler) GetGroupStatusTimeline(ctx echo.Context, _ string, groupID string, params codegen.GetGroupStatusTimelineParams) error {
	statusCountTimeline, err := h.db.GetGroupStatusCountTimeline(groupID, params.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("getGroupStatusCountTimeline - getting status timeline")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, statusCountTimeline)
}

func (h *Handler) GetGroupInstanceStats(ctx echo.Context, _ string, groupID string, params codegen.GetGroupInstanceStatsParams) error {
	instancesStats, err := h.db.GetGroupInstancesStats(groupID, params.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("getGroupInstancesStats - getting instances stats groupID")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, instancesStats)
}

func (h *Handler) GetGroupVersionBreakdown(ctx echo.Context, _ string, groupID string) error {
	versionBreakdown, err := h.db.GetGroupVersionBreakdown(groupID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("getVersionBreakdown - getting version breakdown")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if len(versionBreakdown) == 0 {
		// WAT?: because otherwise it serializes to null not []
		return ctx.JSON(http.StatusOK, []string{})
	}
	return ctx.JSON(http.StatusOK, versionBreakdown)
}

func (h *Handler) GetGroupInstances(ctx echo.Context, appIDorProductID string, groupID string, params codegen.GetGroupInstancesParams) error {
	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	p := api.InstancesQueryParams{
		ApplicationID: appID,
		GroupID:       groupID,
		Status:        params.Status,
		Page:          uint64(*params.Page),
		PerPage:       uint64(*params.Perpage),
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
	if params.SearchFilter != nil {
		p.SearchFilter = *params.SearchFilter
	}
	if params.SearchValue != nil {
		p.SearchValue = *params.SearchValue
	}

	groupInstances, err := h.db.GetInstances(p, params.Duration)
	if err != nil {
		l.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, groupInstances)
}

func (h *Handler) GetGroupInstancesCount(ctx echo.Context, appIDorProductID string, groupID string, params codegen.GetGroupInstancesCountParams) error {
	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	p := api.InstancesQueryParams{
		ApplicationID: appID,
		GroupID:       groupID,
	}

	count, err := h.db.GetInstancesCount(p, params.Duration)
	if err != nil {
		l.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, codegen.InstanceCount{Count: uint64(count)})
}

func groupFromRequest(name string, description *string, policyMaxUpdatesPerPeriod int, policyOfficeHours *bool, policyPeriodInterval string, policySafeMode *bool, policyTimezone string, policyUpdateTimeout string, policyUpdatesEnabled *bool, channelID *string, track *string, groupID string, appID string) *api.Group {
	group := &api.Group{
		Name:                      name,
		PolicyMaxUpdatesPerPeriod: policyMaxUpdatesPerPeriod,
		PolicyPeriodInterval:      policyPeriodInterval,
		PolicyUpdateTimeout:       policyUpdateTimeout,
	}
	if channelID != nil && *channelID != "" {
		group.ChannelID = null.StringFromPtr(channelID)
	}
	if policyTimezone != "" {
		group.PolicyTimezone = null.StringFrom(policyTimezone)
	}
	if groupID != "" {
		group.ID = groupID
	}
	if appID != "" {
		group.ApplicationID = appID
	}
	if track != nil {
		group.Track = *track
	}
	if description != nil {
		group.Description = *description
	}
	if policyOfficeHours != nil {
		group.PolicyOfficeHours = *policyOfficeHours
	}
	if policySafeMode != nil {
		group.PolicySafeMode = *policySafeMode
	}
	if policyUpdatesEnabled != nil {
		group.PolicyUpdatesEnabled = *policyUpdatesEnabled
	}

	return group
}

type groupsPage struct {
	TotalCount int          `json:"totalCount"`
	Count      int          `json:"count"`
	Groups     []*api.Group `json:"groups"`
}
