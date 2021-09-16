package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func (h *Handler) PaginatePackages(ctx echo.Context, appID string, params codegen.PaginatePackagesParams) error {
	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	totalCount, err := h.db.GetPackagesCount(appID)
	if err != nil {
		logger.Error().Err(err).Str("appID", appID).Msg("getPackages count - encoding packages")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	pkgs, err := h.db.GetPackages(appID, *params.Page, *params.Perpage)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("appID", appID).Msg("getPackages - encoding packages")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, packagePage{totalCount, len(pkgs), pkgs})
}

func (h *Handler) CreatePackage(ctx echo.Context, appID string) error {
	logger := loggerWithUsername(logger, ctx)

	var request codegen.CreatePackageInfo

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("addPackage - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	if request.Type == 1 && (request.Filename == nil || request.Hash == nil || request.Size == nil || request.Url == nil) {
		// For the Flatcar type of package these fields are required.

		err := errors.New("required field missing")
		logger.Error().Err(err).Msg("addPackage - required field missing")
		return ctx.NoContent(http.StatusBadRequest)
	}

	pkg := packageFromRequest(appID, request.Arch, request.ChannelsBlacklist, request.Description, request.Filename, request.Hash, request.Size, request.Url, request.Version, request.Type, request.FlatcarAction, "", request.MetadataType, request.MetadataContent)

	pkg, err = h.db.AddPackage(pkg)
	if err != nil {
		logger.Error().Err(err).Msgf("addPackage - adding package %v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	pkg, err = h.db.GetPackage(pkg.ID)
	if err != nil {
		logger.Error().Err(err).Str("packageID", pkg.ID).Msg("addPackage - getting added package")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("addPackage - successfully added package %+v", pkg)

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) GetPackage(ctx echo.Context, appID string, packageID string) error {
	pkg, err := h.db.GetPackage(packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("packageID", packageID).Msg("getPackage - getting package")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) UpdatePackage(ctx echo.Context, appID string, packageID string) error {
	logger := loggerWithUsername(logger, ctx)

	var request codegen.UpdatePackageInfo

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("updatePackage - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	if request.Type == 1 && (request.Filename == nil || request.Hash == nil || request.Size == nil || request.Url == nil) {
		// For the Flatcar type of package these fields are required.
		err := errors.New("required field missing")
		logger.Error().Err(err).Msg("updatePackage - required field missing")
		return ctx.NoContent(http.StatusBadRequest)
	}

	pkg := packageFromRequest(appID, request.Arch, request.ChannelsBlacklist, request.Description, request.Filename, request.Hash, request.Size, request.Url, request.Version, request.Type, request.FlatcarAction, packageID, request.MetadataType, request.MetadataContent)

	oldPkg, err := h.db.GetPackage(packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		logger.Error().Err(err).Str("packageID", packageID).Msg("updatePackage - getting old package to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.UpdatePackage(pkg)
	if err != nil {
		logger.Error().Err(err).Msgf("updatePackage - updating package %+v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	pkg, err = h.db.GetPackage(packageID)
	if err != nil {
		logger.Error().Err(err).Str("packageID", packageID).Msg("updatePackage - getting old package to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("updatePackage - successfully updated package %+v -> %+v", oldPkg, pkg)

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) DeletePackage(ctx echo.Context, appID string, packageID string) error {
	logger := loggerWithUsername(logger, ctx)

	pkg, err := h.db.GetPackage(packageID)
	if err != nil {
		logger.Error().Err(err).Str("packageID", packageID).Msg("deletePackage - getting package to delete")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.DeletePackage(packageID)
	if err != nil {
		logger.Error().Err(err).Str("packageID", packageID).Msg("deletePackage")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logger.Info().Msgf("deletePackage - successfully deleted package %+v", pkg)

	return ctx.NoContent(http.StatusOK)
}

func packageFromRequest(appID string, arch int, ChannelsBlacklist []string, description string, filename *string, hash *string, size *string, url *string, version string, packageType int, flAction *codegen.FlatcarActionPackage, ID string, metadataType *string, metadataContent *string) *api.Package {
	var flatcarAction *api.FlatcarAction

	if flAction != nil {
		if flAction.Id != nil {
			if flatcarAction == nil {
				flatcarAction = &api.FlatcarAction{}
			}
			flatcarAction.ID = *flAction.Id
		}
		if flAction.Sha256 != nil {
			if flatcarAction == nil {
				flatcarAction = &api.FlatcarAction{}
			}
			flatcarAction.Sha256 = *flAction.Sha256
		}
	}

	if ID != "" {
		if flatcarAction == nil {
			flatcarAction = &api.FlatcarAction{}
		}
		flatcarAction.PackageID = ID
	}

	pkg := api.Package{
		ApplicationID:   appID,
		Arch:            api.Arch(arch),
		Description:     null.StringFrom(description),
		Filename:        null.StringFromPtr(filename),
		Hash:            null.StringFromPtr(hash),
		Size:            null.StringFromPtr(size),
		Type:            packageType,
		URL:             null.StringFromPtr(url),
		Version:         version,
		FlatcarAction:   flatcarAction,
		MetadataType:    null.StringFromPtr(metadataType),
		MetadataContent: null.StringFromPtr(metadataContent),
	}
	if ChannelsBlacklist != nil {
		pkg.ChannelsBlacklist = ChannelsBlacklist
	}

	if ID != "" {
		pkg.ID = ID
	}
	return &pkg
}

type packagePage struct {
	TotalCount int            `json:"totalCount"`
	Count      int            `json:"count"`
	Packages   []*api.Package `json:"packages"`
}
