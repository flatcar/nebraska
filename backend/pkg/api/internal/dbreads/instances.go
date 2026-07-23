package dbreads

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

const (
	validityInterval     postgresDuration = "1 days"
	defaultStatsInterval time.Duration    = 24 * time.Hour
)

type instanceFilterItem int

const (
	id instanceFilterItem = iota
	ip
	lastCheckForUpdates
)

var sortFilterMap = map[instanceFilterItem]string{
	id:                  "alias",
	ip:                  "ip",
	lastCheckForUpdates: "last_check_for_updates",
}

type sortOrder int

const (
	sortOrderAsc sortOrder = iota
	sortOrderDesc
)

func sortOrderFromString(str string) sortOrder {
	val, err := strconv.Atoi(str)

	/*
		In case value is other than 0 or 1 or there is a wrong type of sortOrder passed
		fallback to sortOrderDesc
	*/
	if (val != 0 && val != 1) || err != nil {
		return sortOrderDesc
	}
	return sortOrder(val)
}

func sanitizeSortFilterParams(sortFilter string) string {
	sortFilterNumericValue, _ := strconv.Atoi(sortFilter)
	if value, ok := sortFilterMap[instanceFilterItem(sortFilterNumericValue)]; ok {
		return value
	}
	return sortFilterMap[id]
}

// GetInstance returns the instance identified by the id provided.
func (q *Queries) GetInstance(instanceID, appID string) (*types.Instance, error) {
	var instance types.Instance
	query, _, err := goqu.From("instance").
		Where(goqu.C("id").Eq(instanceID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = q.db.QueryRowx(query).StructScan(&instance)
	if err != nil {
		return nil, err
	}
	/* passing "" to sortFilter while invoking getInstanceApp signifies we are not interested
	in a sort
	*/
	instanceApplication, err := q.getInstanceApp(appID, instance.ID, validityInterval, "", 0)
	switch err {
	case nil:
		instance.Application = *instanceApplication
	case sql.ErrNoRows:
		instance.Application = types.InstanceApplication{}
	default:
		return nil, err
	}

	return &instance, nil
}
func (q *Queries) getInstanceApp(appID, instanceID string, duration postgresDuration, sortFilter string, orderOfSort sortOrder) (*types.InstanceApplication, error) {
	var instanceApp types.InstanceApplication
	query, _, err := q.instanceAppQuery(appID, instanceID, duration, sortFilter, orderOfSort).ToSQL()
	if err != nil {
		return nil, err
	}
	err = q.db.QueryRowx(query).StructScan(&instanceApp)
	if err != nil {
		return nil, err
	}
	return &instanceApp, nil
}

// GetInstanceStatusHistory returns the status history of an instance in the
// context of the application/group provided.
func (q *Queries) GetInstanceStatusHistory(instanceID, appID, groupID string, limit uint64) ([]*types.InstanceStatusHistoryEntry, error) {
	var instanceStatusHistory []*types.InstanceStatusHistoryEntry
	query, _, err := q.instanceStatusHistoryQuery(instanceID, appID, groupID, limit).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var instanceStatusHistoryEntity types.InstanceStatusHistoryEntry
		err = rows.StructScan(&instanceStatusHistoryEntity)
		if err != nil {
			return nil, err
		}
		if instanceStatusHistoryEntity.Status == types.InstanceStatusError {
			instanceStatusHistoryEntity.ErrorCode, err = q.GetEvent(instanceID, appID, instanceStatusHistoryEntity.CreatedTs)
			if err != nil {
				return nil, err
			}
		}
		instanceStatusHistory = append(instanceStatusHistory, &instanceStatusHistoryEntity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return instanceStatusHistory, nil
}

func prepareGetInstancesQuery(instanceQuery *goqu.SelectDataset, instanceAppQuery *goqu.SelectDataset) *goqu.SelectDataset {
	return goqu.From(goqu.L("Instance")).With("Instance", instanceQuery).With("application", instanceAppQuery).InnerJoin(
		goqu.L("application"),
		goqu.On(goqu.L("Instance.id").Eq(goqu.L("application.instance_id"))),
	).Select(goqu.L("*"))
}

func prepareSearchQuery(finalQuery *goqu.SelectDataset, p types.InstancesQueryParams) *goqu.SelectDataset {
	searchFilter := p.SearchFilter
	searchValue := p.SearchValue
	searchExpression := "%" + searchValue + "%"
	outputQuery := finalQuery
	if searchFilter == "All" && searchValue != "" {
		// search by alias -> ids -> ip -> date
		outputQuery = finalQuery.Where(
			goqu.Or(goqu.I("alias").ILike(searchExpression),
				goqu.I("id").ILike(searchExpression),
				goqu.L("text(ip)").Like(searchExpression)))
	} else if searchFilter != "" && searchValue != "" {
		if searchFilter == "ip" {
			outputQuery = finalQuery.Where(
				goqu.L("text(ip)").Like(searchExpression))
		} else {
			outputQuery = finalQuery.Where(
				goqu.I(searchFilter).ILike(searchExpression))
		}
	}
	return outputQuery
}

// GetInstances returns all instances that match with the provided criteria.
func (q *Queries) GetInstances(p types.InstancesQueryParams, duration string) (types.InstancesWithTotal, error) {
	var instances []*types.Instance
	var err error
	totalCount, err := q.GetInstancesCount(p, duration)
	if err != nil {
		return types.InstancesWithTotal{}, err
	}
	p.Page, p.PerPage = validatePaginationParams(p.Page, p.PerPage)
	var dbDuration postgresDuration
	dbDuration, _, err = durationParamToPostgresTimings(durationParam(duration))
	if err != nil {
		return types.InstancesWithTotal{}, err
	}

	limit, offset := sqlPaginate(p.Page, p.PerPage)
	sortFilter := sanitizeSortFilterParams(p.SortFilter)
	sortOrder := sortOrderFromString(p.SortOrder)
	instancesQuery := q.instancesQuery(p, dbDuration)
	instancesQuery = instancesQuery.Select("id", "ip", "created_ts", goqu.Case().
		When(goqu.C("alias").Neq(""), goqu.C("alias")).Else(goqu.C("id")).As("alias"))

	instanceAppQuery := prepareInstanceAppQuery()
	finalQuery := prepareGetInstancesQuery(instancesQuery, instanceAppQuery)
	switch sortOrder {
	case sortOrderAsc:
		finalQuery = finalQuery.Order(goqu.I(sortFilter).Asc().NullsLast())
	case sortOrderDesc:
		finalQuery = finalQuery.Order(goqu.I(sortFilter).Desc().NullsLast())
	}

	finalQuery = prepareSearchQuery(finalQuery, p)
	query, _, err := finalQuery.
		Limit(limit).
		Offset(offset).
		ToSQL()
	if err != nil {
		return types.InstancesWithTotal{}, err
	}
	rows, err := q.db.Queryx(query)
	if err != nil {
		return types.InstancesWithTotal{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var instance types.Instance
		err = rows.Scan(&instance.ID, &instance.IP, &instance.CreatedTs, &instance.Alias,
			&instance.Application.Version, &instance.Application.Status, &instance.Application.LastCheckForUpdates,
			&instance.Application.LastUpdateVersion, &instance.Application.UpdateInProgress,
			&instance.Application.ApplicationID, &instance.Application.GroupID, &instance.Application.InstanceID)
		if err != nil {
			return types.InstancesWithTotal{}, err
		}
		instances = append(instances, &instance)
	}
	if err := rows.Err(); err != nil {
		return types.InstancesWithTotal{}, err
	}
	result := types.InstancesWithTotal{
		TotalInstances: uint64(totalCount),
		Instances:      instances,
	}
	return result, nil
}
func prepareInstanceAppQuery() *goqu.SelectDataset {
	return goqu.From("instance_application").
		Select("version", "status", "last_check_for_updates", "last_update_version", "update_in_progress", "application_id", "group_id", "instance_id")
}
func (q *Queries) GetInstancesCount(p types.InstancesQueryParams, duration string) (int, error) {
	var err error

	var dbDuration postgresDuration
	dbDuration, _, err = durationParamToPostgresTimings(durationParam(duration))
	if err != nil {
		return 0, err
	}
	instancesQuery := q.instancesQuery(p, dbDuration)
	instancesQuery = instancesQuery.Select("id", "ip", "created_ts", goqu.Case().
		When(goqu.C("alias").Neq(""), goqu.C("alias")).Else(goqu.C("id")).As("alias"))

	instanceAppQuery := prepareInstanceAppQuery()
	finalQuery := prepareGetInstancesQuery(instancesQuery, instanceAppQuery)
	finalQuery = prepareSearchQuery(finalQuery, p).Select(goqu.L("COUNT(*)"))

	return q.GetCountQuery(finalQuery)
}

// instanceAppQuery returns a SelectDataset prepared to return the app status
// of the app identified by the application id provided for a given instance.
func (q *Queries) instanceAppQuery(appID, instanceID string, duration postgresDuration, sortFilter string, orderOfSort sortOrder) *goqu.SelectDataset {
	query := prepareInstanceAppQuery().Where(goqu.C("application_id").Eq(appID)).
		Where(goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", duration))

	if instanceID != "" {
		query = query.Where(goqu.C("instance_id").Eq(instanceID))
	}

	if sortFilter != "" {
		switch orderOfSort {
		case sortOrderAsc:
			query = query.Order(goqu.I(sortFilter).Asc().NullsLast())
		case sortOrderDesc:
			query = query.Order(goqu.I(sortFilter).Desc().NullsLast())
		}
	}
	return query
}

func ignoreFakeInstanceCondition(instanceIDField string) string {
	return fmt.Sprintf(`(%[1]s IS NULL OR %[1]s NOT LIKE '{________-____-____-____-____________}')`, instanceIDField)
}

func (q *Queries) getFilterInstancesQuery(selectPart exp.LiteralExpression, p types.InstancesQueryParams, duration postgresDuration) *goqu.SelectDataset {
	query := goqu.From("instance_application").
		Select(selectPart).
		Where(goqu.C("application_id").Eq(p.ApplicationID), goqu.C("group_id").Eq(p.GroupID)).
		Where(goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", duration),
			goqu.L(ignoreFakeInstanceCondition("instance_id")))

	if p.Status == types.InstanceStatusUndefined {
		query = query.Where(goqu.L("status IS NULL"))
	} else if p.Status != 0 {
		query = query.Where(goqu.C("status").Eq(p.Status))
	}
	if p.Version != "" {
		query = query.Where(goqu.C("version").Eq(p.Version))
	}
	return query
}

// instancesQuery returns a SelectDataset prepared to return all instances
// that match the criteria provided in InstancesQueryParams.
func (q *Queries) instancesQuery(p types.InstancesQueryParams, duration postgresDuration) *goqu.SelectDataset {
	instancesSubquery := q.getFilterInstancesQuery(goqu.L("instance_id"), p, duration)

	return goqu.From("instance").
		Where(goqu.L("id IN ?", instancesSubquery))
}

// instanceStatusHistoryQuery returns a SelectDataset prepared to return the
// status history of a given instance in the context of an application/group.
func (q *Queries) instanceStatusHistoryQuery(instanceID, appID, groupID string, limit uint64) *goqu.SelectDataset {
	if limit == 0 {
		limit = 20
	}
	return goqu.From("instance_status_history").Where(goqu.C("instance_id").Eq(instanceID)).
		Where(goqu.C("application_id").Eq(appID)).
		Where(goqu.C("group_id").Eq(groupID)).
		Order(goqu.C("created_ts").Desc()).
		Limit(uint(limit))
}

// GetDefaultInterval returns the default interval used for instance stats queries.
func (q *Queries) GetDefaultInterval() time.Duration {
	return defaultStatsInterval
}

// InstanceStatsQuery returns a SelectDataset to estimate the active fleet size at a given point in time.
// It answers "how many instances were part of the active fleet on day X",
// not "how many instances specifically checked in on day X".
//
// Since last_check_for_updates gets overwritten on every check-in, we cannot determine
// whether an instance was active on a specific past day. Instead, we count instances that:
//  1. existed at the time (created_ts <= timestamp)
//  2. are still alive (last_check_for_updates > timestamp - duration)
func (q *Queries) InstanceStatsQuery(t *time.Time, duration *time.Duration) *goqu.SelectDataset {
	if t == nil {
		now := time.Now().UTC()
		t = &now
	}

	if duration == nil {
		d := defaultStatsInterval
		duration = &d
	}

	// Helper function to convert duration to PostgreSQL interval string
	durationToInterval := func(d time.Duration) string {
		if d <= 0 {
			d = time.Microsecond
		}

		parts := []string{}

		hours := int(d.Hours())
		if hours != 0 {
			parts = append(parts, fmt.Sprintf("%d hours", hours))
		}

		remainder := d - time.Duration(hours)*time.Hour
		minutes := int(remainder.Minutes())
		if minutes != 0 {
			parts = append(parts, fmt.Sprintf("%d minutes", minutes))
		}

		remainder -= time.Duration(minutes) * time.Minute
		seconds := int(remainder.Seconds())
		if seconds != 0 {
			parts = append(parts, fmt.Sprintf("%d seconds", seconds))
		}

		remainder -= time.Duration(seconds) * time.Second
		microseconds := remainder.Microseconds()
		if microseconds != 0 {
			parts = append(parts, fmt.Sprintf("%d microseconds", microseconds))
		}

		return strings.Join(parts, " ")
	}

	interval := durationToInterval(*duration)
	timestamp := goqu.L("timestamp ?", goqu.V(t.Format("2006-01-02T15:04:05.999999Z07:00")))
	timestampMinusDuration := goqu.L("timestamp ? - interval ?", goqu.V(t.Format("2006-01-02T15:04:05.999999Z07:00")), interval)

	query := goqu.From(goqu.T("instance_application")).
		Select(
			timestamp,
			goqu.T("channel").Col("name").As("channel_name"),
			goqu.Case().
				When(goqu.T("channel").Col("arch").Eq(1), "AMD64").
				When(goqu.T("channel").Col("arch").Eq(2), "ARM").
				Else("").
				As("arch"),
			goqu.C("version").As("version"),
			goqu.COUNT("*").As("instances")).Distinct().
		Join(goqu.T("groups"), goqu.On(goqu.C("group_id").Eq(goqu.T("groups").Col("id")))).
		Join(goqu.T("channel"), goqu.On(goqu.T("groups").Col("channel_id").Eq(goqu.T("channel").Col("id")))).
		Join(goqu.T("instance"), goqu.On(goqu.T("instance_application").Col("instance_id").Eq(goqu.T("instance").Col("id")))).
		Where(
			goqu.C("last_check_for_updates").Gt(timestampMinusDuration),
			goqu.L(ignoreFakeInstanceCondition("instance_id")),
			goqu.T("instance").Col("created_ts").Lte(timestamp)).
		GroupBy(timestamp,
			goqu.T("channel").Col("name"),
			goqu.T("channel").Col("arch"),
			goqu.C("version")).
		Order(timestamp.Asc())

	return query
}

// GetInstanceStats returns an InstanceStats table with all instances that have
// been previously been checked in.
func (q *Queries) GetInstanceStats() ([]types.InstanceStats, error) {
	query, _, err := goqu.From("instance_stats").
		Order(goqu.C("timestamp").Asc()).ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []types.InstanceStats
	for rows.Next() {
		var instance types.InstanceStats
		err = rows.StructScan(&instance)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// GetInstanceStatsByTimestamp returns an InstanceStats array of instances matching a
// given timestamp value, ordered by version.
func (q *Queries) GetInstanceStatsByTimestamp(t time.Time) ([]types.InstanceStats, error) {
	timestamp := goqu.L("timestamp ?", goqu.V(t.Format("2006-01-02T15:04:05.999999Z07:00")))

	query, _, err := goqu.From("instance_stats").
		Where(goqu.C("timestamp").Eq(timestamp)).
		Order(goqu.C("version").Asc()).ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []types.InstanceStats
	for rows.Next() {
		var instance types.InstanceStats
		err = rows.StructScan(&instance)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, nil
}
