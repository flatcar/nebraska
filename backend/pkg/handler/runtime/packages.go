package runtime

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/handler/internal/shared"
)

func (h *Handler) PaginatePackages(ctx echo.Context, appIDorProductID string, params codegen.PaginatePackagesParams) error {
	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		params.Perpage = &defaultPerPage
	}

	appID, err := h.runtime.GetAppID(appIDorProductID)
	if err != nil {
		return shared.AppNotFoundResponse(ctx, appIDorProductID)
	}

	totalCount, err := h.runtime.GetPackagesCount(appID, params.SearchVersion)
	if err != nil {
		l.Error().Err(err).Str("appID", appID).Msg("getPackages count - encoding packages")
		return ctx.NoContent(http.StatusInternalServerError)
	}
	pkgs, err := h.runtime.GetPackages(appID, uint64(*params.Page), uint64(*params.Perpage), params.SearchVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("appID", appID).Msg("getPackages - encoding packages")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, packagePage{totalCount, len(pkgs), pkgs})
}

func (h *Handler) GetPackage(ctx echo.Context, _ string, packageID string) error {
	pkg, err := h.runtime.GetPackage(packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("packageID", packageID).Msg("getPackage - getting package")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, pkg)
}

type packagePage struct {
	TotalCount int              `json:"totalCount"`
	Count      int              `json:"count"`
	Packages   []*types.Package `json:"packages"`
}

// Floor handlers following existing patterns

// Define the response structure following existing pattern
type floorPackagesPage struct {
	TotalCount int              `json:"totalCount"`
	Count      int              `json:"count"`
	Packages   []*types.Package `json:"packages"`
}

// PaginateChannelFloors handles paginated requests for channel floor packages
func (h *Handler) PaginateChannelFloors(ctx echo.Context, channelID string, params codegen.PaginateChannelFloorsParams) error {
	l := shared.LoggerWithUsername(l, ctx)

	if params.Page == nil {
		params.Page = &defaultPage
	}

	if params.Perpage == nil {
		// Use a larger default for floor packages since they're typically a small set
		// and we want to show all of them in the UI
		defaultFloorPerPage := 100
		params.Perpage = &defaultFloorPerPage
	}

	totalCount, err := h.runtime.GetChannelFloorPackagesCount(channelID)
	if err != nil {
		l.Error().Err(err).Str("channelID", channelID).Msg("PaginateChannelFloors - getting floor packages count")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// If no floors, return empty result immediately
	if totalCount == 0 {
		return ctx.JSON(http.StatusOK, floorPackagesPage{0, 0, []*types.Package{}})
	}

	// Get paginated floor packages
	packages, err := h.runtime.GetChannelFloorPackagesPaginated(channelID, uint64(*params.Page), uint64(*params.Perpage))
	if err != nil {
		if err == sql.ErrNoRows {
			// This shouldn't happen if count > 0, but handle gracefully
			return ctx.JSON(http.StatusOK, floorPackagesPage{totalCount, 0, []*types.Package{}})
		}
		l.Error().Err(err).Str("channelID", channelID).Msg("PaginateChannelFloors - getting floor packages")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, floorPackagesPage{totalCount, len(packages), packages})
}

// GetPackageFloorChannels handles requests for channels where a package is a floor
func (h *Handler) GetPackageFloorChannels(ctx echo.Context, _ string, packageID string) error {
	l := shared.LoggerWithUsername(l, ctx)

	// First verify the package exists
	_, err := h.runtime.GetPackage(packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.NoContent(http.StatusNotFound)
		}
		l.Error().Err(err).Str("packageID", packageID).Msg("GetPackageFloorChannels - getting package")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	channelInfos, err := h.runtime.GetPackageFloorChannels(packageID)
	if err != nil {
		l.Error().Err(err).Str("packageID", packageID).Msg("GetPackageFloorChannels - getting floor channels")
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// Response structure matching the frontend expectations
	type response struct {
		Channels []types.ChannelFloorInfo `json:"channels"`
		Count    int                      `json:"count"`
	}

	return ctx.JSON(http.StatusOK, response{
		Channels: channelInfos,
		Count:    len(channelInfos),
	})
}
