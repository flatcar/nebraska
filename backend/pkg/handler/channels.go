package handler

import (
	"database/sql"
	"net/http"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"
)

func (h *handler) PaginateChannels(ctx echo.Context, appId string, params codegen.PaginateChannelsParams) error {

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	totalCount, err := h.db.GetChannelsCount(appId)
	if err != nil {
		logger.Error().Err(err).Str("appID", appId).Msg("getChannels count - getting channels")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channels, err := h.db.GetChannels(appId, *params.Page, *params.Perpage)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appId).Msg("getChannels - getting channels")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, channelsPage{totalCount, len(channels), channels})

}

func (h *handler) CreateChannel(ctx echo.Context, appId string) error {
	logger := loggerWithUsername(logger, ctx)

	var request codegen.CreateChannelInfo
	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("addChannel")
		return ctx.NoContent(http.StatusBadRequest)
	}

	channel := newChannel(appId, request.Arch, request.Color, request.Name, request.PackageId)
	_, err = h.db.AddChannel(channel)
	if err != nil {
		logger.Error().Err(err).Msgf("addChannel channel %v", channel)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel, err = h.db.GetChannel(channel.ID)
	if err != nil {
		logger.Error().Err(err).Str("channelID", channel.ID).Msg("addChannel")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("addChannel - successfully added channel %+v", channel)
	return ctx.JSON(http.StatusOK, channel)
}

func (h *handler) GetChannel(ctx echo.Context, appId string, channelId string) error {
	channel, err := h.db.GetChannel(channelId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("channelID", channelId).Msg("getChannel - getting updated channel")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	return ctx.JSON(http.StatusOK, channel)
}

func (h *handler) UpdateChannel(ctx echo.Context, appId string, channelId string) error {
	logger := loggerWithUsername(logger, ctx)

	var request codegen.UpdateChannelInfo

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("updateChannel")
		return ctx.NoContent(http.StatusBadRequest)
	}

	oldChannel, err := h.db.GetChannel(channelId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("channelID", channelId).Msg("updateChannel - getting old channel to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel := newChannel(appId, request.Arch, request.Color, request.Name, request.PackageId)
	channel.ID = channelId

	err = h.db.UpdateChannel(channel)
	if err != nil {
		logger.Error().Err(err).Msgf("updateChannel - updating channel %+v", channel)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel, err = h.db.GetChannel(channelId)
	if err != nil {
		logger.Error().Err(err).Str("channelID", channel.ID).Msg("updateChannel - getting channel updated")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("updateChannel - successfully updated channel %+v (PACKAGE: %+v) -> %+v (PACKAGE: %+v)", oldChannel, oldChannel.Package, channel, channel.Package)

	return ctx.JSON(http.StatusOK, channel)
}

func (h *handler) DeleteChannel(ctx echo.Context, appId string, channelId string) error {

	logger := loggerWithUsername(logger, ctx)

	channel, err := h.db.GetChannel(channelId)
	if err != nil {
		logger.Error().Err(err).Str("channelID", channel.ID).Msg("updateChannel - getting channel to be deleted")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.DeleteChannel(channelId)
	if err != nil {
		logger.Error().Err(err).Str("channelID", channelId).Msg("deleteChannel")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("deleteChannel - successfully deleted channel %+v (PACKAGE: %+v)", channel, channel.Package)

	return ctx.NoContent(http.StatusOK)
}

func newChannel(appID string, arch uint, color string, name string, packageId *string) *api.Channel {
	channel := &api.Channel{
		ApplicationID: appID,
		Name:          name,
		Color:         color,
		Arch:          api.Arch(arch),
		PackageID:     null.StringFromPtr(packageId),
	}
	return channel
}

type channelsPage struct {
	TotalCount int            `json:"totalCount"`
	Count      int            `json:"count"`
	Channels   []*api.Channel `json:"channels"`
}
