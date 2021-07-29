package handler

import (
	"database/sql"
	"net/http"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/labstack/echo/v4"
)

func (h *handler) PaginateActivity(ctx echo.Context, params codegen.PaginateActivityParams) error {

	teamID := getTeamID(ctx)

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	var p api.ActivityQueryParams
	if params.AppId != nil {
		p.AppID = *params.AppId
	}
	if params.GroupId != nil {
		p.GroupID = *params.GroupId
	}
	if params.ChannelId != nil {
		p.ChannelID = *params.ChannelId
	}
	if params.InstanceId != nil {
		p.InstanceID = *params.InstanceId
	}
	if params.Version != nil {
		p.Version = *params.Version
	}
	if params.Severity != nil {
		p.Severity = *params.Severity
	}
	p.Start = params.Start
	p.End = params.End
	p.Page = *params.Page
	p.PerPage = *params.Perpage

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
