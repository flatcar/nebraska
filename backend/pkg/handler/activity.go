package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func (h *Handler) PaginateActivity(ctx echo.Context, params codegen.PaginateActivityParams) error {
	teamID := getTeamID(ctx)

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	var p api.ActivityQueryParams
	if params.AppIDorProductID != nil {
		appID, err := h.db.GetAppID(*params.AppIDorProductID)
		if err != nil {
			return appNotFoundResponse(ctx, *params.AppIDorProductID)
		}
		p.AppID = appID
	}
	if params.GroupID != nil {
		p.GroupID = *params.GroupID
	}
	if params.ChannelID != nil {
		p.ChannelID = *params.ChannelID
	}
	if params.InstanceID != nil {
		p.InstanceID = *params.InstanceID
	}
	if params.Version != nil {
		p.Version = *params.Version
	}
	if params.Severity != nil {
		p.Severity = *params.Severity
	}
	p.Start, _ = time.Parse(time.RFC3339, params.Start)
	p.End, _ = time.Parse(time.RFC3339, params.End)
	p.Page = uint64(*params.Page)
	p.PerPage = uint64(*params.Perpage)

	totalCount, err := h.db.GetActivityCount(teamID, p)
	if err != nil {
		logger.Error().Err(err).Str("teamID", teamID).Msgf("getActivity count params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	activityEntries, err := h.db.GetActivity(teamID, p)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("teamID", teamID).Msgf("getActivity params %v", p)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, activityPage{totalCount, len(activityEntries), activityEntries})
}

type activityPage struct {
	TotalCount int             `json:"totalCount"`
	Count      int             `json:"count"`
	Activities []*api.Activity `json:"activities"`
}
