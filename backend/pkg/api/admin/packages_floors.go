package admin

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// AddChannelPackageFloor marks a package as a floor for a specific channel
func (s *Service) AddChannelPackageFloor(channelID, packageID string, floorReason null.String) error {
	// Verify channel and package exist and are compatible in a single query
	var channelArch, pkgArch types.Arch
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

	err = s.db.QueryRow(query).Scan(&channelArch, &channelAppID, &pkgArch, &pkgAppID)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.ErrInvalidPackage
		}
		return err
	}

	// Verify architecture and application match
	if channelArch != pkgArch {
		return types.ErrArchMismatch
	}
	if channelAppID != pkgAppID {
		return types.ErrInvalidApplicationOrGroup
	}

	// Check if package is blacklisted for this channel
	isBlacklisted, err := s.isPackageBlacklistedForChannel(packageID, channelID)
	if err != nil {
		return err
	}
	if isBlacklisted {
		return types.ErrPackageBlacklisted
	}

	query, _, err = goqu.Insert("channel_package_floors").
		Cols("channel_id", "package_id", "floor_reason").
		Vals(goqu.Vals{channelID, packageID, floorReason}).
		OnConflict(goqu.DoUpdate("channel_id, package_id", goqu.Record{"floor_reason": floorReason})).
		ToSQL()

	if err != nil {
		return err
	}

	_, err = s.db.Exec(query)
	return err
}

// isPackageBlacklistedForChannel checks if a package is blacklisted for a specific channel
func (s *Service) isPackageBlacklistedForChannel(packageID, channelID string) (bool, error) {
	query, _, err := goqu.From("package_channel_blacklist").
		Select(goqu.COUNT("*")).
		Where(goqu.And(
			goqu.C("channel_id").Eq(channelID),
			goqu.C("package_id").Eq(packageID),
		)).
		ToSQL()

	if err != nil {
		return false, err
	}

	var count int
	err = s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// isPackageFloorForChannel checks if a package is marked as a floor for a specific channel
func (s *Service) isPackageFloorForChannel(packageID, channelID string) (bool, error) {
	query, _, err := goqu.From("channel_package_floors").
		Select(goqu.COUNT("*")).
		Where(goqu.And(
			goqu.C("channel_id").Eq(channelID),
			goqu.C("package_id").Eq(packageID),
		)).
		ToSQL()

	if err != nil {
		return false, err
	}

	var count int
	err = s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// RemoveChannelPackageFloor removes a package from being a floor for a specific channel
func (s *Service) RemoveChannelPackageFloor(channelID, packageID string) error {
	query, _, err := goqu.Delete("channel_package_floors").
		Where(goqu.And(
			goqu.C("channel_id").Eq(channelID),
			goqu.C("package_id").Eq(packageID),
		)).
		ToSQL()

	if err != nil {
		return err
	}

	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return types.ErrNoRowsAffected
	}

	return nil
}
