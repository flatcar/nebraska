package api

import (
	"database/sql"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

var (
	// ErrPackageBlacklisted indicates that the package is blacklisted for this channel
	ErrPackageBlacklisted = errors.New("nebraska: cannot mark blacklisted package as floor")
)

// AddChannelPackageFloor marks a package as a floor for a specific channel
func (api *API) AddChannelPackageFloor(channelID, packageID string, floorReason null.String) error {
	// Verify channel and package exist and are compatible in a single query
	var channelArch, pkgArch Arch
	var channelAppID, pkgAppID string

	query, _, err := goqu.From(goqu.T("channel").As("c")).
		CrossJoin(goqu.T("package").As("p")).
		Select(
			goqu.C("arch").Table("c"),
			goqu.C("application_id").Table("c"),
			goqu.C("arch").Table("p"),
			goqu.C("application_id").Table("p"),
		).
		Where(goqu.And(
			goqu.C("id").Table("c").Eq(channelID),
			goqu.C("id").Table("p").Eq(packageID),
		)).
		ToSQL()

	if err != nil {
		return err
	}

	err = api.db.QueryRow(query).Scan(&channelArch, &channelAppID, &pkgArch, &pkgAppID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidPackage
		}
		return err
	}

	// Verify architecture and application match
	if channelArch != pkgArch {
		return ErrArchMismatch
	}
	if channelAppID != pkgAppID {
		return ErrInvalidApplicationOrGroup
	}

	// Check if package is blacklisted for this channel
	isBlacklisted, err := api.IsPackageBlacklistedForChannel(packageID, channelID)
	if err != nil {
		return err
	}
	if isBlacklisted {
		return ErrPackageBlacklisted
	}

	query, _, err = goqu.Insert("channel_package_floors").
		Cols("channel_id", "package_id", "floor_reason").
		Vals(goqu.Vals{channelID, packageID, floorReason}).
		OnConflict(goqu.DoUpdate("channel_id, package_id", goqu.Record{"floor_reason": floorReason})).
		ToSQL()

	if err != nil {
		return err
	}

	_, err = api.db.Exec(query)
	return err
}

// RemoveChannelPackageFloor removes a package from being a floor for a specific channel
func (api *API) RemoveChannelPackageFloor(channelID, packageID string) error {
	query, _, err := goqu.Delete("channel_package_floors").
		Where(goqu.And(
			goqu.C("channel_id").Eq(channelID),
			goqu.C("package_id").Eq(packageID),
		)).
		ToSQL()

	if err != nil {
		return err
	}

	result, err := api.db.Exec(query)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return nil
}

type ChannelFloorInfo = types.ChannelFloorInfo
