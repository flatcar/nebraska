package handler

import (
	"database/sql"
	"net/http"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/labstack/echo/v4"
)

func (h *handler) GetInstance(ctx echo.Context, appId string, groupId string, instanceId string) error {

	instance, err := h.db.GetInstance(instanceId, appId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appId).Str("instanceID", instanceId).Msg("getInstance - getting instance")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, instance)
}

func (h *handler) GetInstanceStatusHistory(ctx echo.Context, appId string, groupId string, instanceId string, params codegen.GetInstanceStatusHistoryParams) error {

	instanceStatusHistory, err := h.db.GetInstanceStatusHistory(instanceId, appId, groupId, params.Limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appId).Str("groupID", groupId).Str("instanceID", instanceId).Msgf("getInstanceStatusHistory - getting status history limit %d", params.Limit)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, instanceStatusHistory)
}

func (h *handler) UpdateInstance(ctx echo.Context, instanceId string) error {

	logger := loggerWithUsername(logger, ctx)

	var request codegen.UpdateInstanceInfo

	err := ctx.Bind(&request)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	instance, err := h.db.UpdateInstance(instanceId, request.Alias)
	if err != nil {
		logger.Error().Err(err).Str("instance", instanceId).Msgf("updateInstance - updating params %s", request.Alias)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("updateInstance - successfully updated instance %q alias to %q", instanceId, instance.Alias)

	return ctx.JSON(http.StatusOK, instance)
}
