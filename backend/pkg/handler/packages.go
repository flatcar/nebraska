package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

func (h *Handler) PaginatePackages(ctx echo.Context, appIDorProductID string, params codegen.PaginatePackagesParams) error {
	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	totalCount, err := h.db.GetPackagesCount(appID, params.SearchVersion)
	if err != nil {
		l.Error().Err(err).Str("appID", appID).Msg("getPackages count - encoding packages")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	pkgs, err := h.db.GetPackages(appID, uint64(*params.Page), uint64(*params.Perpage), params.SearchVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("appID", appID).Msg("getPackages - encoding packages")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, packagePage{totalCount, len(pkgs), pkgs})
}

func (h *Handler) CreatePackage(ctx echo.Context, appIDorProductID string) error {
	l := loggerWithUsername(l, ctx)

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.PackageConfig

	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("addPackage - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	pkg := packageFromRequest(appID, request.Arch, request.ChannelsBlacklist, request.Description, request.Filename, request.Hash, request.Size, request.Url, request.Version, request.Type, request.FlatcarAction, "", request.ExtraFiles)

	pkg, err = h.db.AddPackage(pkg)
	if err != nil {
		l.Error().Err(err).Msgf("addPackage - adding package %v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	pkg, err = h.db.GetPackage(pkg.ID)
	if err != nil {
		l.Error().Err(err).Str("packageID", pkg.ID).Msg("addPackage - getting added package")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("addPackage - successfully added package %+v", pkg)

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) GetPackage(ctx echo.Context, _ string, packageID string) error {
	pkg, err := h.db.GetPackage(packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("packageID", packageID).Msg("getPackage - getting package")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) UpdatePackage(ctx echo.Context, appIDorProductID string, packageID string) error {
	l := loggerWithUsername(l, ctx)

	appID, err := h.db.GetAppID(appIDorProductID)
	if err != nil {
		return appNotFoundResponse(ctx, appIDorProductID)
	}

	var request codegen.PackageConfig

	err = ctx.Bind(&request)
	if err != nil {
		l.Error().Err(err).Msg("updatePackage - decoding payload")
		return ctx.NoContent(http.StatusBadRequest)
	}

	pkg := packageFromRequest(appID, request.Arch, request.ChannelsBlacklist, request.Description, request.Filename, request.Hash, request.Size, request.Url, request.Version, request.Type, request.FlatcarAction, packageID, request.ExtraFiles)

	oldPkg, err := h.db.GetPackage(packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("packageID", packageID).Msg("updatePackage - getting old package to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.UpdatePackage(pkg)
	if err != nil {
		l.Error().Err(err).Msgf("updatePackage - updating package %+v", request)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	pkg, err = h.db.GetPackage(packageID)
	if err != nil {
		l.Error().Err(err).Str("packageID", packageID).Msg("updatePackage - getting old package to update")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("updatePackage - successfully updated package %+v -> %+v", oldPkg, pkg)

	return ctx.JSON(http.StatusOK, pkg)
}

func (h *Handler) DeletePackage(ctx echo.Context, _ string, packageID string) error {
	l := loggerWithUsername(l, ctx)

	pkg, err := h.db.GetPackage(packageID)
	if err != nil {
		l.Error().Err(err).Str("packageID", packageID).Msg("deletePackage - getting package to delete")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.db.DeletePackage(packageID)
	if err != nil {
		l.Error().Err(err).Str("packageID", packageID).Msg("deletePackage")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	l.Info().Msgf("deletePackage - successfully deleted package %+v", pkg)

	return ctx.NoContent(http.StatusNoContent)
}

func packageFromRequest(appID string, arch int, ChannelsBlacklist []string, description string, filename string, hash string, size string, url string, version string, packageType int, flAction *codegen.FlatcarActionPackage, ID string, extraFiles *codegen.ExtraFiles) *api.Package {
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

	var extraFilesArray []api.File
	if extraFiles != nil {
		for _, file := range *extraFiles {
			f := api.File{
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

	pkg := api.Package{
		ApplicationID: appID,
		Arch:          api.Arch(arch),
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

type packagePage struct {
	TotalCount int            `json:"totalCount"`
	Count      int            `json:"count"`
	Packages   []*api.Package `json:"packages"`
}
