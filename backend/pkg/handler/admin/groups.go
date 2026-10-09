package admin

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/handler/internal/shared"
)

func (h *Handler) CreateGroup(ctx echo.Context, appIDorProductID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	appID, err := h.admin.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.GroupConfig
	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("addGroup - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	group := groupFromRequest(request.Name, request.Description, request.PolicyMaxUpdatesPerPeriod, request.PolicyOfficeHours, request.PolicyPeriodInterval, request.PolicySafeMode, request.PolicyTimezone, request.PolicyUpdateTimeout, request.PolicyUpdatesEnabled, request.ChannelId, request.Track, "", appID)

	group, err = h.admin.AddGroup(group)
	if err != nil {
		l.Error().Err(err).Msgf("addGroup - adding group %v", group)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group, err = h.admin.GetGroup(group.ID)
	if err != nil {
		l.Error().Err(err).Msgf("addGroup - adding group %v", group)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	l.Info().Msgf("addGroup - successfully added group %+v", group)

	return ctx.JSON(http.StatusOK, group)
}

func (h *Handler) UpdateGroup(ctx echo.Context, appIDorProductID string, groupID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	appID, err := h.admin.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.GroupConfig
	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("updateGroup - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	oldGroup, err := h.admin.GetGroup(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("updateGroup - getting old group to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	group := groupFromRequest(request.Name, request.Description, request.PolicyMaxUpdatesPerPeriod, request.PolicyOfficeHours, request.PolicyPeriodInterval, request.PolicySafeMode, request.PolicyTimezone, request.PolicyUpdateTimeout, request.PolicyUpdatesEnabled, request.ChannelId, request.Track, groupID, appID)

	err = h.admin.UpdateGroup(group)
	if err != nil {
		l.Error().Err(err).Msgf("updateGroup - updating group %+v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// A single-instance deployment has one updates-enabled switch, so re-enabling
	// updates here also releases a brake this node tripped for itself.
	if h.conf.InstanceMode.IsSingle() &&
		!oldGroup.PolicyUpdatesEnabled && group.PolicyUpdatesEnabled {
		if err := h.brake.ClearUpdatesEnabledOverride(groupID); err != nil {
			l.Error().Err(err).Str("groupID", groupID).Msg("updateGroup - clearing local updates-enabled override")
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	group, err = h.admin.GetGroup(groupID)
	if err != nil {
		l.Error().Err(err).Str("groupID", groupID).Msg("getGroup - getting group")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("updateGroup - successfully updated group %+v -> %+v", oldGroup, group)

	return ctx.JSON(http.StatusOK, group)
}

func (h *Handler) DeleteGroup(ctx echo.Context, _ string, groupID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	group, err := h.admin.GetGroup(groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("groupID", groupID).Msg("updateGroup - getting old group to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.admin.DeleteGroup(groupID)
	if err != nil {
		l.Error().Err(err).Str("groupID", groupID).Msg("deleteGroup")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("deleteGroup - successfully deleted group %+v", group)

	return ctx.NoContent(http.StatusNoContent)
}

func groupFromRequest(name string, description *string, policyMaxUpdatesPerPeriod int, policyOfficeHours *bool, policyPeriodInterval string, policySafeMode *bool, policyTimezone string, policyUpdateTimeout string, policyUpdatesEnabled *bool, channelID *string, track *string, groupID string, appID string) *types.Group {
	group := &types.Group{
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
