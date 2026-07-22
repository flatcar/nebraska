package api

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/blang/semver/v4"
	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
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

// GetRequiredChannelFloors returns the floor packages between the instance version and
// the channel target (instance < floor <= target), sorted ascending by semantic version
// and capped at the configured limit.
func (api *API) GetRequiredChannelFloors(channel *Channel, instanceVersion string) ([]*Package, error) {
	if channel == nil || channel.Package == nil {
		return nil, ErrNoPackageFound
	}
	if instanceVersion == "" {
		return nil, fmt.Errorf("instance version cannot be empty")
	}

	limit := DefaultMaxFloorsPerResponse
	if api.maxFloorsPerResponse > 0 {
		limit = api.maxFloorsPerResponse
	}

	allFloors, err := api.GetChannelFloorPackages(channel.ID)
	if err != nil {
		return nil, err
	}

	floors, _, err := selectFloorsInRange(allFloors, instanceVersion, channel.Package.Version, limit)
	return floors, err
}

// selectFloorsInRange returns the floor packages with instanceVersion < v <= targetVersion,
// sorted ascending by semantic version and capped at limit. hasMore is true when more
// floors fall in range than the limit.
func selectFloorsInRange(floors []*Package, instanceVersion, targetVersion string, limit int) ([]*Package, bool, error) {
	inst, err := semver.Make(instanceVersion)
	if err != nil {
		return nil, false, fmt.Errorf("invalid instance version %q: %w", instanceVersion, err)
	}
	target, err := semver.Make(targetVersion)
	if err != nil {
		return nil, false, fmt.Errorf("invalid target version %q: %w", targetVersion, err)
	}

	inRange := make([]*Package, 0, len(floors))
	for _, p := range floors {
		v, err := semver.Make(p.Version)
		if err != nil {
			continue // versions are validated on insert; skip defensively
		}
		if v.GT(inst) && v.LTE(target) {
			inRange = append(inRange, p)
		}
	}

	sort.SliceStable(inRange, func(i, j int) bool {
		vi, _ := semver.Make(inRange[i].Version)
		vj, _ := semver.Make(inRange[j].Version)
		return vi.LT(vj)
	})

	hasMore := len(inRange) > limit
	if hasMore {
		inRange = inRange[:limit]
	}
	return inRange, hasMore, nil
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

type ChannelFloorInfo = types.ChannelFloorInfo

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
