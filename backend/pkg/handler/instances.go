package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/kinvolk/nebraska/backend/pkg/metrics"
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

func (h *Handler) GetLatestInstanceStats(ctx echo.Context) error {
	metrics.InstanceMetricsHandler.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}

func (h *Handler) GetInstanceStats(ctx echo.Context, params codegen.GetInstanceStatsParams) error {
	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	totalCount, err := h.db.GetInstanceStatsCount()
	if err != nil {
		logger.Error().Err(err).Msg("get instance stats count")
		return ctx.NoContent(http.StatusBadRequest)
	}

	metrics, err := h.db.GetInstanceStats(uint64(*params.Page), uint64(*params.Perpage))
	if err != nil {
		logger.Error().Err(err).Msg("getInstanceStats - getting instance stats")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	ctx.Response().Header().Set(echo.HeaderContentType, "application/x-ndjson")
	ctx.Response().WriteHeader(http.StatusOK)

	m := make([]map[string]interface{}, len(metrics))
	for i, metric := range metrics {
		formattedMetric := map[string]interface{}{
			"type":      "instance_count",
			"channel":   metric.ChannelName,
			"version":   metric.Version,
			"arch":      metric.Arch,
			"timestamp": metric.Timestamp,
			"count":     metric.Instances,
		}

		m[i] = formattedMetric
	}

	p := metricPage{
		TotalCount: totalCount,
		Count:      len(metrics),
		Metrics:    m,
	}
	if err := json.NewEncoder(ctx.Response()).Encode(p); err != nil {
		logger.Error().Err(err).Msg("getInstanceStats - encoding instance stats")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return nil
}

type metricPage struct {
	TotalCount int                      `json:"totalCount"`
	Count      int                      `json:"count"`
	Metrics    []map[string]interface{} `json:"metrics"`
}
