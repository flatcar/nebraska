package runtime

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/handler/internal/shared"
)

func (h *Handler) PaginateChannels(ctx echo.Context, appIDorProductID string, params codegen.PaginateChannelsParams) error {
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

	totalCount, err := h.runtime.GetChannelsCount(appID)
	if err != nil {
		l.Error().Err(err).Str("appID", appID).Msg("getChannels count - getting channels")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channels, err := h.runtime.GetChannels(appID, uint64(*params.Page), uint64(*params.Perpage))
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("appID", appID).Msg("getChannels - getting channels")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, channelsPage{totalCount, len(channels), channels})
}

func (h *Handler) GetChannel(ctx echo.Context, _ string, channelID string) error {
	channel, err := h.runtime.GetChannel(channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("channelID", channelID).Msg("getChannel - getting updated channel")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, channel)
}

type channelsPage struct {
	TotalCount int              `json:"totalCount"`
	Count      int              `json:"count"`
	Channels   []*types.Channel `json:"channels"`
}
