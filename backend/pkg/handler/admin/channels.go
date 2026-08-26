package admin

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/handler/internal/shared"
)

func (h *Handler) CreateChannel(ctx echo.Context, appIDorProductID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	var request codegen.ChannelConfig
	err := ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("addChannel")
		return ctx.NoContent(http.StatusBadRequest)
	}

	appID, err := h.admin.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}
	channel := newChannel(appID, request.Arch, request.Color, request.Name, request.PackageId)
	_, err = h.admin.AddChannel(channel)
	if err != nil {
		l.Error().Err(err).Msgf("addChannel channel %v", channel)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel, err = h.admin.GetChannel(channel.ID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channel.ID).Msg("addChannel")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("addChannel - successfully added channel %+v", channel)
	return ctx.JSON(http.StatusOK, channel)
}

func (h *Handler) UpdateChannel(ctx echo.Context, appIDorProductID string, channelID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	appID, err := h.admin.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.ChannelConfig

	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("updateChannel")
		return ctx.NoContent(http.StatusBadRequest)
	}

	oldChannel, err := h.admin.GetChannel(channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("channelID", channelID).Msg("updateChannel - getting old channel to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel := newChannel(appID, request.Arch, request.Color, request.Name, request.PackageId)
	channel.ID = channelID

	err = h.admin.UpdateChannel(channel)
	if err != nil {
		l.Error().Err(err).Msgf("updateChannel - updating channel %+v", channel)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channel, err = h.admin.GetChannel(channelID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channel.ID).Msg("updateChannel - getting channel updated")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("updateChannel - successfully updated channel %+v (PACKAGE: %+v) -> %+v (PACKAGE: %+v)", oldChannel, oldChannel.Package, channel, channel.Package)

	return ctx.JSON(http.StatusOK, channel)
}

func (h *Handler) DeleteChannel(ctx echo.Context, _ string, channelID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	channel, err := h.admin.GetChannel(channelID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channel.ID).Msg("updateChannel - getting channel to be deleted")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.admin.DeleteChannel(channelID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channelID).Msg("deleteChannel")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("deleteChannel - successfully deleted channel %+v (PACKAGE: %+v)", channel, channel.Package)

	return ctx.NoContent(http.StatusNoContent)
}

func newChannel(appID string, arch uint, color string, name string, packageID *string) *types.Channel {
	channel := &types.Channel{
		ApplicationID: appID,
		Name:          name,
		Color:         color,
		Arch:          types.Arch(arch),
	}
	if packageID != nil && *packageID != "" {
		channel.PackageID = null.StringFromPtr(packageID)
	}
	return channel
}
