package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

func (h *Handler) PaginateChannels(ctx echo.Context, appIDorProductID string, params codegen.PaginateChannelsParams) error {
	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	totalCount, err := h.db.GetChannelsCount(appID)
	if err != nil {
		l.Error().Err(err).Str("appID", appID).Msg("getChannels count - getting channels")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channels, err := h.db.GetChannels(appID, uint64(*params.Page), uint64(*params.Perpage))
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("appID", appID).Msg("getChannels - getting channels")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, channelsPage{totalCount, len(channels), channels})
}

func (h *Handler) CreateChannel(ctx echo.Context, appIDorProductID string) error {
	l := loggerWithUsername(l, ctx)

	var request codegen.ChannelConfig
	err := ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("addChannel")
		return ctx.NoContent(http.StatusBadRequest)
	}

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}
	channel := newChannel(appID, request.Arch, request.Color, request.Name, request.PackageId)
	_, err = h.db.AddChannel(channel)
	if err != nil {
		l.Error().Err(err).Msgf("addChannel channel %v", channel)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel, err = h.db.GetChannel(channel.ID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channel.ID).Msg("addChannel")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("addChannel - successfully added channel %+v", channel)
	return ctx.JSON(http.StatusOK, channel)
}

func (h *Handler) GetChannel(ctx echo.Context, _ string, channelID string) error {
	channel, err := h.db.GetChannel(channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("channelID", channelID).Msg("getChannel - getting updated channel")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, channel)
}

func (h *Handler) UpdateChannel(ctx echo.Context, appIDorProductID string, channelID string) error {
	l := loggerWithUsername(l, ctx)

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.ChannelConfig

	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("updateChannel")
		return ctx.NoContent(http.StatusBadRequest)
	}

	oldChannel, err := h.db.GetChannel(channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("channelID", channelID).Msg("updateChannel - getting old channel to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel := newChannel(appID, request.Arch, request.Color, request.Name, request.PackageId)
	channel.ID = channelID

	err = h.db.UpdateChannel(channel)
	if err != nil {
		l.Error().Err(err).Msgf("updateChannel - updating channel %+v", channel)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel, err = h.db.GetChannel(channelID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channel.ID).Msg("updateChannel - getting channel updated")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("updateChannel - successfully updated channel %+v (PACKAGE: %+v) -> %+v (PACKAGE: %+v)", oldChannel, oldChannel.Package, channel, channel.Package)

	return ctx.JSON(http.StatusOK, channel)
}

func (h *Handler) DeleteChannel(ctx echo.Context, _ string, channelID string) error {
	l := loggerWithUsername(l, ctx)

	channel, err := h.db.GetChannel(channelID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channel.ID).Msg("updateChannel - getting channel to be deleted")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.DeleteChannel(channelID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channelID).Msg("deleteChannel")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("deleteChannel - successfully deleted channel %+v (PACKAGE: %+v)", channel, channel.Package)

	return ctx.NoContent(http.StatusNoContent)
}

func newChannel(appID string, arch uint, color string, name string, packageID *string) *api.Channel {
	channel := &api.Channel{
		ApplicationID: appID,
		Name:          name,
		Color:         color,
		Arch:          api.Arch(arch),
	}
	if packageID != nil && *packageID != "" {
		channel.PackageID = null.StringFromPtr(packageID)
	}
	return channel
}

type channelsPage struct {
	TotalCount int            `json:"totalCount"`
	Count      int            `json:"count"`
	Channels   []*api.Channel `json:"channels"`
}
