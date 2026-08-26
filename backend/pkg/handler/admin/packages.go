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

func (h *Handler) CreatePackage(ctx echo.Context, appIDorProductID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	appID, err := h.admin.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.PackageConfig

	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("addPackage - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	pkg := packageFromRequest(appID, request.Arch, request.ChannelsBlacklist, request.Description, request.Filename, request.Hash, request.Size, request.Url, request.Version, request.Type, request.FlatcarAction, "", request.ExtraFiles)

	pkg, err = h.admin.AddPackage(pkg)
	if err != nil {
		l.Error().Err(err).Msgf("addPackage - adding package %v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	pkg, err = h.admin.GetPackage(pkg.ID)
	if err != nil {
		l.Error().Err(err).Str("packageID", pkg.ID).Msg("addPackage - getting added package")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("addPackage - successfully added package %+v", pkg)

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) UpdatePackage(ctx echo.Context, appIDorProductID string, packageID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	appID, err := h.admin.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.PackageConfig

	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("updatePackage - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	pkg := packageFromRequest(appID, request.Arch, request.ChannelsBlacklist, request.Description, request.Filename, request.Hash, request.Size, request.Url, request.Version, request.Type, request.FlatcarAction, packageID, request.ExtraFiles)

	oldPkg, err := h.admin.GetPackage(packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("packageID", packageID).Msg("updatePackage - getting old package to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.admin.UpdatePackage(pkg)
	if err != nil {
		l.Error().Err(err).Msgf("updatePackage - updating package %+v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	pkg, err = h.admin.GetPackage(packageID)
	if err != nil {
		l.Error().Err(err).Str("packageID", packageID).Msg("updatePackage - getting old package to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("updatePackage - successfully updated package %+v -> %+v", oldPkg, pkg)

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) DeletePackage(ctx echo.Context, _ string, packageID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	pkg, err := h.admin.GetPackage(packageID)
	if err != nil {
		l.Error().Err(err).Str("packageID", packageID).Msg("deletePackage - getting package to delete")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.admin.DeletePackage(packageID)
	if err != nil {
		l.Error().Err(err).Str("packageID", packageID).Msg("deletePackage")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("deletePackage - successfully deleted package %+v", pkg)

	return ctx.NoContent(http.StatusNoContent)
}

func packageFromRequest(appID string, arch int, ChannelsBlacklist []string, description string, filename string, hash string, size string, url string, version string, packageType int, flAction *codegen.FlatcarActionPackage, ID string, extraFiles *codegen.ExtraFiles) *types.Package {
	var flatcarAction *types.FlatcarAction

	if flAction != nil {
		if flAction.Id != nil {
			if flatcarAction == nil {
				flatcarAction = &types.FlatcarAction{}
			}
			flatcarAction.ID = *flAction.Id
		}
		if flAction.Sha256 != nil {
			if flatcarAction == nil {
				flatcarAction = &types.FlatcarAction{}
			}
			flatcarAction.Sha256 = *flAction.Sha256
		}
	}

	if ID != "" {
		if flatcarAction == nil {
			flatcarAction = &types.FlatcarAction{}
		}
		flatcarAction.PackageID = ID
	}

	var extraFilesArray []types.File
	if extraFiles != nil {
		for _, file := range *extraFiles {
			f := types.File{
				Name:    null.StringFrom(*file.Name),
				Hash:    null.StringFrom(*file.Hash),
				Hash256: null.StringFrom(*file.Hash256),
				Size:    null.StringFrom(*file.Size),
			}
			if file.Id != nil {
				f.ID = int64(*file.Id)
			}
			extraFilesArray = append(extraFilesArray, f)
		}
	}

	pkg := types.Package{
		ApplicationID: appID,
		Arch:          types.Arch(arch),
		Description:   null.StringFrom(description),
		Filename:      null.StringFrom(filename),
		Hash:          null.StringFrom(hash),
		Size:          null.StringFrom(size),
		Type:          packageType,
		URL:           url,
		Version:       version,
		FlatcarAction: flatcarAction,
		ExtraFiles:    extraFilesArray,
	}
	if ChannelsBlacklist != nil {
		pkg.ChannelsBlacklist = ChannelsBlacklist
	}

	if ID != "" {
		pkg.ID = ID
	}
	return &pkg
}

// SetChannelFloor handles creating or updating a floor package relationship (idempotent)
func (h *Handler) SetChannelFloor(ctx echo.Context, channelID string, packageID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	var request codegen.SetChannelFloorJSONRequestBody
	if err := ctx.Bind(&request); err != nil {
		l.Error().Err(err).Msg("SetChannelFloor - binding request")
		return ctx.NoContent(http.StatusBadRequest)
	}

	// Convert floor reason to null.String
	var floorReason null.String
	if request.FloorReason != nil {
		floorReason = null.StringFrom(*request.FloorReason)
	}

	// Use AddChannelPackageFloor which already handles upsert via ON CONFLICT
	if err := h.admin.AddChannelPackageFloor(channelID, packageID, floorReason); err != nil {
		switch err {
		case types.ErrInvalidPackage:
			return ctx.NoContent(http.StatusNotFound)
		case types.ErrArchMismatch:
			return ctx.String(http.StatusBadRequest, "Architecture mismatch between channel and package")
		case types.ErrInvalidApplicationOrGroup:
			return ctx.String(http.StatusBadRequest, "Package does not belong to the same application as the channel")
		default:
			l.Error().Err(err).Str("channelID", channelID).Str("packageID", packageID).Msg("SetChannelFloor - setting floor")
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	l.Info().Str("channelID", channelID).Str("packageID", packageID).Msg("SetChannelFloor - successfully set floor")

	// Return JSON response
	response := map[string]interface{}{
		"channel_id": channelID,
		"package_id": packageID,
		"status":     "ok",
	}
	return ctx.JSON(http.StatusOK, response)
}

// RemoveChannelFloor handles removing a package as a floor for a channel
func (h *Handler) RemoveChannelFloor(ctx echo.Context, channelID string, packageID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	if err := h.admin.RemoveChannelPackageFloor(channelID, packageID); err != nil {
		if err == types.ErrNoRowsAffected {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("channelID", channelID).Str("packageID", packageID).Msg("RemoveChannelFloor - removing floor")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Str("channelID", channelID).Str("packageID", packageID).Msg("RemoveChannelFloor - successfully removed floor")
	return ctx.NoContent(http.StatusNoContent)
}
