package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func (h *Handler) PaginateApps(ctx echo.Context, params codegen.PaginateAppsParams) error {
	teamID := getTeamID(ctx)

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	totalCount, err := h.db.GetAppsCount(teamID)
	if err != nil {
		logger.Error().Err(err).Str("teamID", teamID).Msg("getApps count - getting apps")
		return ctx.NoContent(http.StatusBadRequest)
	}

	apps, err := h.db.GetApps(teamID, *params.Page, *params.Perpage)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("teamID", teamID).Msg("getApps - getting apps")
		return ctx.NoContent(http.StatusBadRequest)
	}

	return ctx.JSON(http.StatusOK, applicationPage{totalCount, len(apps), apps})
}

func (h *Handler) CreateApp(ctx echo.Context, params codegen.CreateAppParams) error {
	logger := loggerWithUsername(logger, ctx)

	teamID := getTeamID(ctx)

	var request codegen.CreateAppInfo
	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("addApp - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	app := appFromRequest(request.Name, request.Description, "", teamID, request.ProductId)

	source := ""
	if params.CloneFrom != nil {
		source = *params.CloneFrom
	}

	app, err = h.db.AddAppCloning(app, source)
	if err != nil {
		logger.Error().Err(err).Str("sourceAppID", *params.CloneFrom).Msgf("addApp - cloning app %v", app)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	app, err = h.db.GetApp(app.ID)
	if err != nil {
		logger.Error().Err(err).Str("appID", app.ID).Msg("addApp - getting added app")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("addApp - successfully added app %+v", app)
	return ctx.JSON(http.StatusOK, app)
}

func (h *Handler) GetApp(ctx echo.Context, appID string) error {
	app, err := h.db.GetApp(appID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appID).Msg("getApp - getting app")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, app)
}

func (h *Handler) UpdateApp(ctx echo.Context, appID string) error {
	logger := loggerWithUsername(logger, ctx)

	var request codegen.UpdateAppInfo
	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("updateApp - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	oldApp, err := h.db.GetApp(appID)
	if err != nil {
		logger.Error().Err(err).Str("appID", appID).Msg("updateApp - getting old app to update")
		return ctx.NoContent(http.StatusBadRequest)
	}

	app := appFromRequest(request.Name, request.Description, appID, "", request.ProductId)

	err = h.db.UpdateApp(app)
	if err != nil {
		logger.Error().Err(err).Msgf("updatedApp - updating app %s", appID)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	app, err = h.db.GetApp(appID)
	if err != nil {
		logger.Error().Err(err).Str("appID", appID).Msg("updateApp - getting updated app")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// TODO: Confirm if old and new values should be logged
	logger.Info().Msgf("updateApp - successfully updated app %+v -> %+v", oldApp, app)

	return ctx.JSON(http.StatusOK, app)
}

func (h *Handler) DeleteApp(ctx echo.Context, appID string) error {
	logger := loggerWithUsername(logger, ctx)

	app, err := h.db.GetApp(appID)
	if err != nil {
		logger.Error().Err(err).Str("appID", appID).Msg("deleteApp - getting app to delete")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.DeleteApp(appID)
	if err != nil {
		logger.Error().Err(err).Str("appID", appID).Msg("deleteApp")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	logger.Info().Msgf("deleteApp - successfully deleted app %+v", app)

	return ctx.NoContent(http.StatusOK)
}

func appFromRequest(name string, description string, appID string, teamID string, productID *string) *api.Application {
	productIDNullString := null.String{}
	if productID != nil && *productID != "" {
		productIDNullString = null.StringFrom(*productID)
	}
	app := api.Application{
		TeamID:      teamID,
		Name:        name,
		Description: description,
		ProductID:   productIDNullString,
	}
	if teamID != "" {
		app.TeamID = teamID
	}
	if appID != "" {
		app.ID = appID
	}

	return &app
}

type applicationPage struct {
	TotalCount   int                `json:"totalCount"`
	Count        int                `json:"count"`
	Applications []*api.Application `json:"applications"`
}
