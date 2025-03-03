package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func (h *Handler) GetInstance(ctx echo.Context, appIDorProductID string, _ string, instanceID string) error {
	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	instance, err := h.db.GetInstance(instanceID, appID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appID).Str("instanceID", instanceID).Msg("getInstance - getting instance")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, instance)
}

func (h *Handler) GetInstanceStatusHistory(ctx echo.Context, appIDorProductID string, groupID string, instanceID string, params codegen.GetInstanceStatusHistoryParams) error {
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	instanceStatusHistory, err := h.db.GetInstanceStatusHistory(instanceID, appID, groupID, uint64(limit))
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appID).Str("groupID", groupID).Str("instanceID", instanceID).Msgf("getInstanceStatusHistory - getting status history limit %d", params.Limit)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, instanceStatusHistory)
}

func (h *Handler) UpdateInstance(ctx echo.Context, instanceID string) error {
	logger := loggerWithUsername(logger, ctx)

	var request codegen.UpdateInstanceConfig

	err := ctx.Bind(&request)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	instance, err := h.db.UpdateInstance(instanceID, request.Alias)
	if err != nil {
		logger.Error().Err(err).Str("instance", instanceID).Msgf("updateInstance - updating params %s", request.Alias)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("updateInstance - successfully updated instance %q alias to %q", instanceID, instance.Alias)

	return ctx.JSON(http.StatusOK, instance)
}
