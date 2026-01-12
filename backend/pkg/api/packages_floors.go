package api

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"
)

var (
	// ErrPackageBlacklisted indicates that the package is blacklisted for this channel
	ErrPackageBlacklisted = errors.New("nebraska: cannot mark blacklisted package as floor")
)

// semverToIntArray returns a PostgreSQL expression that converts a semantic version
// to an integer array for proper version comparison.
// Handles versions like "1.2.3", "1.2.3-beta", "1.2.3+build"
// The column parameter must be a safe SQL identifier (no user input!)
func semverToIntArray(column string) (string, error) {
	if column != "?" && !regexp.MustCompile(`^[a-z_]+(\.[a-z_]+)?$`).MatchString(column) {
		return "", fmt.Errorf("semverToIntArray: invalid column name %q - potential SQL injection", column)
	}
	return fmt.Sprintf("string_to_array((regexp_split_to_array(%s, '[+-]'))[1], '.')::int[]", column), nil
}

// versionCompareExpr creates a version comparison expression
func versionCompareExpr(column, operator, value string) (goqu.Expression, error) {
	// Validate operator to prevent SQL injection
	validOperators := map[string]bool{
		">": true, ">=": true, "<": true, "<=": true, "=": true, "!=": true,
	}
	if !validOperators[operator] {
		return nil, fmt.Errorf("versionCompareExpr: invalid operator %q - potential SQL injection", operator)
	}

	colArray, err := semverToIntArray(column)
	if err != nil {
		return nil, err
	}

	valArray, err := semverToIntArray("?")
	if err != nil {
		return nil, err
	}

	return goqu.L(fmt.Sprintf("%s %s %s", colArray, operator, valArray), value), nil
}

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
	isBlacklisted, err := api.isPackageBlacklistedForChannel(packageID, channelID)
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

// isPackageBlacklistedForChannel checks if a package is blacklisted for a specific channel
func (api *API) isPackageBlacklistedForChannel(packageID, channelID string) (bool, error) {
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
	err = api.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// isPackageFloorForChannel checks if a package is marked as a floor for a specific channel
func (api *API) isPackageFloorForChannel(packageID, channelID string) (bool, error) {
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
	err = api.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
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

// GetChannelFloorPackages returns all floor packages for a specific channel
func (api *API) GetChannelFloorPackages(channelID string) ([]*Package, error) {
	// No blacklist check needed for floors
	semverExpr, err := semverToIntArray("p.version")
	if err != nil {
		return nil, err
	}

	query, _, err := goqu.From(goqu.L(`
		package p
		JOIN channel_package_floors cpf ON p.id = cpf.package_id
	`)).
		Select(goqu.L(`
			p.*,
			true as is_floor,
			cpf.floor_reason
		`)).
		Where(goqu.C("channel_id").Table("cpf").Eq(channelID)).
		Order(goqu.L(semverExpr).Asc()).
		ToSQL()

	if err != nil {
		return nil, err
	}

	return api.getPackagesFromQuery(query)
}

const (
	// DefaultMaxFloorsPerResponse is the default maximum number of floor versions
	// to return in a single update response. This limit prevents:
	// - Timeouts during sequential syncer updates
	// - Very large Omaha responses
	// - Overwhelming syncers with too many intermediate versions
	// Can be overridden via NEBRASKA_MAX_FLOORS_PER_RESPONSE env var
	DefaultMaxFloorsPerResponse = 5
)

// GetRequiredChannelFloorsWithLimit returns floor packages between instance and target versions,
// along with a boolean indicating if more floors remain beyond the limit.
// This uses LIMIT+1 approach: query for limit+1 rows, if we get more than limit, there are more.
// This is more efficient than a separate COUNT query.
func (api *API) GetRequiredChannelFloorsWithLimit(channel *Channel, instanceVersion string) ([]*Package, bool, error) {
	if channel == nil || channel.Package == nil {
		return nil, false, ErrNoPackageFound
	}
	if instanceVersion == "" {
		return nil, false, fmt.Errorf("instance version cannot be empty")
	}

	targetVersion := channel.Package.Version

	maxFloorsPerResponse := DefaultMaxFloorsPerResponse
	if api.maxFloorsPerResponse > 0 {
		maxFloorsPerResponse = api.maxFloorsPerResponse
	}

	// No blacklist check needed for floors
	gtExpr, err := versionCompareExpr("p.version", ">", instanceVersion)
	if err != nil {
		return nil, false, err
	}

	lteExpr, err := versionCompareExpr("p.version", "<=", targetVersion)
	if err != nil {
		return nil, false, err
	}

	semverExpr, err := semverToIntArray("p.version")
	if err != nil {
		return nil, false, err
	}

	// Query for LIMIT+1 to detect if more floors exist
	query, _, err := goqu.From(goqu.L(`
		package p
		JOIN channel_package_floors cpf ON p.id = cpf.package_id
	`)).
		Select(goqu.L(`
			p.*,
			true as is_floor,
			cpf.floor_reason
		`)).
		Where(goqu.And(
			goqu.C("channel_id").Table("cpf").Eq(channel.ID),
			gtExpr,
			lteExpr,
		)).
		Order(goqu.L(semverExpr).Asc()).
		Limit(uint(maxFloorsPerResponse + 1)).
		ToSQL()

	if err != nil {
		return nil, false, err
	}

	floors, err := api.getPackagesFromQuery(query)
	if err != nil {
		return nil, false, err
	}

	// If we got more than the limit, there are more floors remaining
	if len(floors) > maxFloorsPerResponse {
		// Return only up to the limit, indicate more remain
		return floors[:maxFloorsPerResponse], true, nil
	}

	// All floors returned
	return floors, false, nil
}

// GetChannelFloorPackagesCount returns the count of floor packages for a channel
func (api *API) GetChannelFloorPackagesCount(channelID string) (int, error) {
	query := goqu.From("channel_package_floors").
		Where(goqu.C("channel_id").Eq(channelID)).
		Select(goqu.L("count(*)"))
	return api.GetCountQuery(query)
}

// GetChannelFloorPackagesPaginated returns paginated floor packages for a channel
func (api *API) GetChannelFloorPackagesPaginated(channelID string, page, perPage uint64) ([]*Package, error) {
	page, perPage = validatePaginationParams(page, perPage)
	limit, offset := sqlPaginate(page, perPage)

	// No blacklist check needed for floors
	semverExpr, err := semverToIntArray("p.version")
	if err != nil {
		return nil, err
	}

	query, _, err := goqu.From(goqu.L(`
		package p
		JOIN channel_package_floors cpf ON p.id = cpf.package_id
	`)).
		Select(goqu.L(`
			p.*,
			true as is_floor,
			cpf.floor_reason
		`)).
		Where(goqu.C("channel_id").Table("cpf").Eq(channelID)).
		Order(goqu.L(semverExpr).Asc()).
		Limit(limit).
		Offset(offset).
		ToSQL()

	if err != nil {
		return nil, err
	}

	return api.getPackagesFromQuery(query)
}

// ChannelFloorInfo contains a channel and its floor reason for a specific package
type ChannelFloorInfo struct {
	Channel     *Channel    `json:"channel"`
	FloorReason null.String `json:"floor_reason"`
}

// GetPackageFloorChannels returns all channels where a package is marked as a floor
func (api *API) GetPackageFloorChannels(packageID string) ([]ChannelFloorInfo, error) {
	// Use a temporary struct that embeds Channel and adds floor_reason
	type channelWithFloor struct {
		Channel
		FloorReason null.String `db:"floor_reason"`
	}

	query, _, err := goqu.From(goqu.T("channel").As("c")).
		Join(goqu.T("channel_package_floors").As("cpf"), goqu.On(
			goqu.C("id").Table("c").Eq(goqu.C("channel_id").Table("cpf")),
		)).
		Select(
			goqu.C("id").Table("c"),
			goqu.C("name").Table("c"),
			goqu.C("color").Table("c"),
			goqu.C("created_ts").Table("c"),
			goqu.C("application_id").Table("c"),
			goqu.C("package_id").Table("c"),
			goqu.C("arch").Table("c"),
			goqu.C("floor_reason").Table("cpf"),
		).
		Where(goqu.C("package_id").Table("cpf").Eq(packageID)).
		Order(goqu.C("name").Table("c").Asc()).
		ToSQL()

	if err != nil {
		return nil, err
	}

	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ChannelFloorInfo
	for rows.Next() {
		var chWithFloor channelWithFloor
		if err := rows.StructScan(&chWithFloor); err != nil {
			return nil, err
		}

		// Load the package that the channel points to (if any) - same as getChannelsFromQuery
		if chWithFloor.PackageID.Valid {
			pkg, err := api.getPackage(chWithFloor.PackageID)
			switch err {
			case nil:
				chWithFloor.Package = pkg
			case sql.ErrNoRows:
				chWithFloor.Package = nil
			default:
				return nil, err
			}
		}

		result = append(result, ChannelFloorInfo{
			Channel:     &chWithFloor.Channel,
			FloorReason: chWithFloor.FloorReason,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
