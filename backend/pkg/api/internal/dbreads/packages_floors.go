package dbreads

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"

	"github.com/blang/semver/v4"
	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
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

// GetRequiredChannelFloors returns the floor packages between the instance version and
// the channel target (instance < floor <= target), sorted ascending by semantic version.
func (q *Queries) GetRequiredChannelFloors(channel *types.Channel, instanceVersion string) ([]*types.Package, error) {
	if channel == nil || channel.Package == nil {
		return nil, types.ErrNoPackageFound
	}
	if instanceVersion == "" {
		return nil, fmt.Errorf("instance version cannot be empty")
	}

	allFloors, err := q.GetChannelFloorPackages(channel.ID)
	if err != nil {
		return nil, err
	}

	return selectFloorsInRange(allFloors, instanceVersion, channel.Package.Version)
}

// selectFloorsInRange returns the floor packages with instanceVersion < v <= targetVersion,
// sorted ascending by semantic version.
func selectFloorsInRange(floors []*types.Package, instanceVersion, targetVersion string) ([]*types.Package, error) {
	inst, err := semver.Make(instanceVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid instance version %q: %w", instanceVersion, err)
	}
	target, err := semver.Make(targetVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid target version %q: %w", targetVersion, err)
	}

	inRange := make([]*types.Package, 0, len(floors))
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

	return inRange, nil
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
