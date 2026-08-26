package runtime

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/handler/internal/shared"
)

func (h *Handler) PaginateGroups(ctx echo.Context, appIDorProductID string, params codegen.PaginateGroupsParams) error {
	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	appID, err := h.runtime.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	totalCount, err := h.runtime.GetGroupsCount(appID)
	if err != nil {
		l.Error().Err(err).Str("appID", appID).Msg("getGroups count - getting groups")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	groups, err := h.runtime.GetGroups(appID, uint64(*params.Page), uint64(*params.Perpage))
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

func (h *Handler) GetGroup(ctx echo.Context, _ string, groupID string) error {
	group, err := h.runtime.GetGroup(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("getGroup - getting group")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, group)
}

// ClearGroupUpdatesOverride releases the safe-mode brake this node tripped for
// itself. It writes group_local only, leaving the admin default on groups, which
// a control node replicates out, untouched.
func (h *Handler) ClearGroupUpdatesOverride(ctx echo.Context, _ string, groupID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	// A single-instance deployment presents one updates-enabled switch, so it has
	// no node-local override to clear separately from the group default.
	if h.conf.InstanceMode.IsSingle() {
		return ctx.JSON(http.StatusNotImplemented, map[string]any{
			"error":       "local_override_not_supported",
			"description": "Node-local overrides exist only in a distributed deployment. Enable updates on the group instead.",
		})
	}

	if err := h.runtime.ClearUpdatesEnabledOverride(groupID); err != nil {
		if err == types.ErrNoRowsAffected {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("clearGroupUpdatesOverride - clearing local updates-enabled override")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Str("groupID", groupID).Msg("clearGroupUpdatesOverride - cleared the local updates-enabled override")

	return ctx.NoContent(http.StatusNoContent)
}

func (h *Handler) GetGroupVersionTimeline(ctx echo.Context, _ string, groupID openapi_types.UUID, params codegen.GetGroupVersionTimelineParams) error {
	groupIDStr := groupID.String()

	versionCountTimeline, isCache, err := h.runtime.GetGroupVersionCountTimeline(groupIDStr, params.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupIDStr).Msg("getGroupVersionCountTimeline - getting version timeline")
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
	statusCountTimeline, err := h.runtime.GetGroupStatusCountTimeline(groupID, params.Duration)
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
	instancesStats, err := h.runtime.GetGroupInstancesStats(groupID, params.Duration)
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
	versionBreakdown, err := h.runtime.GetGroupVersionBreakdown(groupID)

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
	appID, err := h.runtime.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	p := types.InstancesQueryParams{
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

	groupInstances, err := h.runtime.GetInstances(p, params.Duration)
	if err != nil {
		l.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, groupInstances)
}

func (h *Handler) GetGroupInstancesCount(ctx echo.Context, appIDorProductID string, groupID string, params codegen.GetGroupInstancesCountParams) error {
	appID, err := h.runtime.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	p := types.InstancesQueryParams{
		ApplicationID: appID,
		GroupID:       groupID,
	}

	count, err := h.runtime.GetInstancesCount(p, params.Duration)
	if err != nil {
		l.Error().Err(err).Msgf("getInstances - getting instances params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, codegen.InstanceCount{Count: uint64(count)})
}

type groupsPage struct {
	TotalCount int            `json:"totalCount"`
	Count      int            `json:"count"`
	Groups     []*types.Group `json:"groups"`
}
