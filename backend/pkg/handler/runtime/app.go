package runtime

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/handler/internal/shared"
)

func (h *Handler) PaginateApps(ctx echo.Context, params codegen.PaginateAppsParams) error {
	teamID := shared.GetTeamID(ctx)

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	totalCount, err := h.runtime.GetAppsCount(teamID)
	if err != nil {
		l.Error().Err(err).Str("teamID", teamID).Msg("getApps count - getting apps")
		return ctx.NoContent(http.StatusBadRequest)
	}

	apps, err := h.runtime.GetApps(teamID, uint64(*params.Page), uint64(*params.Perpage))
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("teamID", teamID).Msg("getApps - getting apps")
		return ctx.NoContent(http.StatusBadRequest)
	}

	return ctx.JSON(http.StatusOK, applicationPage{totalCount, len(apps), apps})
}

func (h *Handler) GetApp(ctx echo.Context, appIDorProductID string) error {
	appID, err := h.runtime.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	app, err := h.runtime.GetApp(appID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("appID", appID).Msg("getApp - getting app")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, app)
}

type applicationPage struct {
	TotalCount   int                  `json:"totalCount"`
	Count        int                  `json:"count"`
	Applications []*types.Application `json:"applications"`
}
