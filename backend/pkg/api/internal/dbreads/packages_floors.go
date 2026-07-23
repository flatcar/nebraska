package dbreads

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
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

// IsPackageBlacklistedForChannel checks if a package is blacklisted for a specific channel
func (q *Queries) IsPackageBlacklistedForChannel(packageID, channelID string) (bool, error) {
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
	err = q.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsPackageFloorForChannel checks if a package is marked as a floor for a specific channel
func (q *Queries) IsPackageFloorForChannel(packageID, channelID string) (bool, error) {
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
	err = q.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetChannelFloorPackages returns all floor packages for a specific channel
func (q *Queries) GetChannelFloorPackages(channelID string) ([]*types.Package, error) {
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

	return q.getPackagesFromQuery(query)
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

// GetRequiredChannelFloors returns floor packages between instance and target versions for a channel
func (q *Queries) GetRequiredChannelFloors(channel *types.Channel, instanceVersion string) ([]*types.Package, error) {
	if channel == nil || channel.Package == nil {
		return nil, ErrNoPackageFound
	}
	if instanceVersion == "" {
		return nil, fmt.Errorf("instance version cannot be empty")
	}

	targetVersion := channel.Package.Version

	maxFloorsPerResponse := DefaultMaxFloorsPerResponse
	if q.maxFloorsPerResponse > 0 {
		maxFloorsPerResponse = q.maxFloorsPerResponse
	}

	// No blacklist check needed for floors
	gtExpr, err := versionCompareExpr("p.version", ">", instanceVersion)
	if err != nil {
		return nil, err
	}

	lteExpr, err := versionCompareExpr("p.version", "<=", targetVersion)
	if err != nil {
		return nil, err
	}

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
		Where(goqu.And(
			goqu.C("channel_id").Table("cpf").Eq(channel.ID),
			gtExpr,
			lteExpr,
		)).
		Order(goqu.L(semverExpr).Asc()).
		Limit(uint(maxFloorsPerResponse)).
		ToSQL()

	if err != nil {
		return nil, err
	}

	return q.getPackagesFromQuery(query)
}

// GetChannelFloorPackagesCount returns the count of floor packages for a channel
func (q *Queries) GetChannelFloorPackagesCount(channelID string) (int, error) {
	query := goqu.From("channel_package_floors").
		Where(goqu.C("channel_id").Eq(channelID)).
		Select(goqu.L("count(*)"))
	return q.GetCountQuery(query)
}

// GetChannelFloorPackagesPaginated returns paginated floor packages for a channel
func (q *Queries) GetChannelFloorPackagesPaginated(channelID string, page, perPage uint64) ([]*types.Package, error) {
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

	return q.getPackagesFromQuery(query)
}

// GetPackageFloorChannels returns all channels where a package is marked as a floor
func (q *Queries) GetPackageFloorChannels(packageID string) ([]types.ChannelFloorInfo, error) {
	// Use a temporary struct that embeds Channel and adds floor_reason
	type channelWithFloor struct {
		types.Channel
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

	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.ChannelFloorInfo
	for rows.Next() {
		var chWithFloor channelWithFloor
		if err := rows.StructScan(&chWithFloor); err != nil {
			return nil, err
		}

		// Load the package that the channel points to (if any) - same as getChannelsFromQuery
		if chWithFloor.PackageID.Valid {
			pkg, err := q.getPackage(chWithFloor.PackageID)
			switch err {
			case nil:
				chWithFloor.Package = pkg
			case sql.ErrNoRows:
				chWithFloor.Package = nil
			default:
				return nil, err
			}
		}

		result = append(result, types.ChannelFloorInfo{
			Channel:     &chWithFloor.Channel,
			FloorReason: chWithFloor.FloorReason,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
