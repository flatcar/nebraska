package api

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
)

const (
	// InstanceStatusUndefined indicates that the instance hasn't sent yet an
	// event to Nebraska so it doesn't know in which state it is.
	InstanceStatusUndefined int = 1 + iota

	// InstanceStatusUpdateGranted indicates that the instance has been granted
	// an update (it should be reporting soon through events how is it going).
	InstanceStatusUpdateGranted

	// InstanceStatusError indicates that the instance reported an error while
	// processing the update.
	InstanceStatusError

	// InstanceStatusComplete indicates that the instance completed the update
	// process successfully.
	InstanceStatusComplete

	// InstanceStatusInstalled indicates that the instance has installed the
	// downloaded packages, but it hasn't applied it or restarted yet.
	InstanceStatusInstalled

	// InstanceStatusDownloaded indicates that the instance downloaded
	// successfully the update package.
	InstanceStatusDownloaded

	// InstanceStatusDownloading indicates that the instance started
	// downloading the update package.
	InstanceStatusDownloading

	// InstanceStatusOnHold indicates that the instance hasn't been granted an
	// update because one of the rollout policy limits has been reached.
	InstanceStatusOnHold
)

const (
	validityInterval postgresDuration = "1 days"
	defaultInterval  time.Duration    = 2 * time.Hour
)

// Instance represents an instance running one or more applications for which
// Nebraska can provide updates.
type Instance struct {
	ID          string              `db:"id" json:"id"`
	IP          string              `db:"ip" json:"ip"`
	CreatedTs   time.Time           `db:"created_ts" json:"created_ts"`
	Application InstanceApplication `db:"application" json:"application,omitempty"`
	Alias       string              `db:"alias" json:"alias,omitempty"`
}
type InstancesWithTotal struct {
	TotalInstances uint64      `json:"total"`
	Instances      []*Instance `json:"instances"`
}

// InstanceApplication represents some details about an application running on
// a given instance: current version of the app, last time the instance checked
// for updates for this app, etc.
type InstanceApplication struct {
	InstanceID          string      `db:"instance_id" json:"instance_id,omitempty"`
	ApplicationID       string      `db:"application_id" json:"application_id"`
	GroupID             null.String `db:"group_id" json:"group_id"`
	Version             string      `db:"version" json:"version"`
	CreatedTs           time.Time   `db:"created_ts" json:"created_ts"`
	Status              null.Int    `db:"status" json:"status"`
	LastCheckForUpdates time.Time   `db:"last_check_for_updates" json:"last_check_for_updates"`
	LastUpdateGrantedTs null.Time   `db:"last_update_granted_ts" json:"last_update_granted_ts"`
	LastUpdateVersion   null.String `db:"last_update_version" json:"last_update_version"`
	UpdateInProgress    bool        `db:"update_in_progress" json:"update_in_progress"`
}

// InstanceStatusHistoryEntry represents an entry in the instance status
// history.
type InstanceStatusHistoryEntry struct {
	ID            int         `db:"id" json:"-"`
	Status        int         `db:"status" json:"status"`
	Version       string      `db:"version" json:"version"`
	CreatedTs     time.Time   `db:"created_ts" json:"created_ts"`
	InstanceID    string      `db:"instance_id" json:"-"`
	ApplicationID string      `db:"application_id" json:"-"`
	GroupID       string      `db:"group_id" json:"-"`
	ErrorCode     null.String `db:"error_code" json:"error_code"`
}

// InstancesQueryParams represents a helper structure used to pass a set of
// parameters when querying instances.
type InstancesQueryParams struct {
	ApplicationID string `json:"application_id"`
	GroupID       string `json:"group_id"`
	Status        int    `json:"status"`
	Version       string `json:"version"`
	Page          uint64 `json:"page"`
	PerPage       uint64 `json:"perpage"`
	SortFilter    string `json:"sort_filter"`
	SortOrder     string `json:"sort_order"`
	SearchFilter  string `json:"search_filter"`
	SearchValue   string `json:"search_value"`
}

type InstanceStats struct {
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
	ChannelName string    `db:"channel_name" json:"channel_name"`
	Arch        string    `db:"arch" json:"arch"`
	Version     string    `db:"version" json:"version"`
	Instances   int       `db:"instances" json:"instances"`
}

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

// RegisterInstance registers an instance into Nebraska.
func (api *API) RegisterInstance(instanceID, instanceAlias, instanceIP, instanceVersion, appID, groupID string) (*Instance, error) {
	if !isValidSemver(instanceVersion) {
		return nil, ErrInvalidSemver
	}
	var err error
	if appID, groupID, err = api.validateApplicationAndGroup(appID, groupID); err != nil {
		return nil, err
	}

	// We want to avoid having to create an unneeded DB transaction, so we check whether it
	// is necessary (we need it when writing into the two tables, instance and
	// instance_application).

	updateInstance := true
	updateInstanceApplication := true

	instance, err := api.GetInstance(instanceID, appID)
	if err == nil {
		// Give precedence to an existing alias over an omitted or empty alias field
		if instanceAlias == "" {
			instanceAlias = instance.Alias
		}
		// The instance exists, so we just update it if its IP or Alias changed
		updateInstance = instance.IP != instanceIP || instance.Alias != instanceAlias

		recent := nowUTC().Add(-5 * time.Minute)

		// And we only update the instance_application if the latest registry is outdated or
		// older than what we establish as recent.
		updateInstanceApplication = instance.Application.LastCheckForUpdates.UTC().Before(recent) ||
			instance.Application.Version != instanceVersion || instance.Application.GroupID.String != groupID

		// Skip updating anything unnecessary
		if !updateInstance && !updateInstanceApplication {
			return instance, nil
		}
	}

	upsertInstance, _, err := goqu.Insert("instance").
		Cols("id", "ip", "alias").
		Vals(goqu.Vals{instanceID, instanceIP, instanceAlias}).
		OnConflict(goqu.DoUpdate("id", goqu.Record{"id": instanceID, "ip": instanceIP, "alias": instanceAlias})).
		ToSQL()
	if err != nil {
		return nil, err
	}

	upsertInstanceApplication, _, err := goqu.Insert("instance_application").
		Cols("instance_id", "application_id", "group_id", "version", "last_check_for_updates").
		Vals(goqu.Vals{instanceID, appID, groupID, instanceVersion, nowUTC()}).
		OnConflict(goqu.DoUpdate("ON CONSTRAINT instance_application_pkey", goqu.Record{"group_id": groupID, "version": instanceVersion, "last_check_for_updates": nowUTC()})).
		ToSQL()
	if err != nil {
		return nil, err
	}

	// If we only have one table to update, then we just do that here directly and avoid the
	// transaction below.
	if updateInstance != updateInstanceApplication {
		queryToExec := upsertInstance
		if updateInstanceApplication {
			queryToExec = upsertInstanceApplication
		}
		_, err := api.db.Exec(queryToExec)
		if err != nil {
			return nil, err
		}

		return instance, nil
	}

	// If this is an instance we haven't seen yet, then we write into instance + instance_application
	tx, err := api.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logger.Error().Err(err).Msg("RegisterInstance - could not roll back")
		}
	}()

	result, err := tx.Exec(upsertInstance)
	if err != nil {
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("RegisterInstance instance insert failed")
	}

	result, err = tx.Exec(upsertInstanceApplication)
	if err != nil {
		return nil, err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("RegisterInstance upsert for instance_application failed")
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return api.GetInstance(instanceID, appID)
}

// GetInstance returns the instance identified by the id provided.
func (api *API) GetInstance(instanceID, appID string) (*Instance, error) {
	var instance Instance
	query, _, err := goqu.From("instance").
		Where(goqu.C("id").Eq(instanceID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&instance)
	if err != nil {
		return nil, err
	}
	/* passing "" to sortFilter while invoking getInstanceApp signifies we are not interested
	in a sort
	*/
	instanceApplication, err := api.getInstanceApp(appID, instance.ID, validityInterval, "", 0)
	switch err {
	case nil:
		instance.Application = *instanceApplication
	case sql.ErrNoRows:
		instance.Application = InstanceApplication{}
	default:
		return nil, err
	}

	return &instance, nil
}
func (api *API) getInstanceApp(appID, instanceID string, duration postgresDuration, sortFilter string, orderOfSort sortOrder) (*InstanceApplication, error) {
	var instanceApp InstanceApplication
	query, _, err := api.instanceAppQuery(appID, instanceID, duration, sortFilter, orderOfSort).ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&instanceApp)
	if err != nil {
		return nil, err
	}
	return &instanceApp, nil
}

// GetInstanceStatusHistory returns the status history of an instance in the
// context of the application/group provided.
func (api *API) GetInstanceStatusHistory(instanceID, appID, groupID string, limit uint64) ([]*InstanceStatusHistoryEntry, error) {
	var instanceStatusHistory []*InstanceStatusHistoryEntry
	query, _, err := api.instanceStatusHistoryQuery(instanceID, appID, groupID, limit).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var instanceStatusHistoryEntity InstanceStatusHistoryEntry
		err = rows.StructScan(&instanceStatusHistoryEntity)
		if err != nil {
			return nil, err
		}
		if instanceStatusHistoryEntity.Status == InstanceStatusError {
			instanceStatusHistoryEntity.ErrorCode, err = api.GetEvent(instanceID, appID, instanceStatusHistoryEntity.CreatedTs)
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

func prepareSearchQuery(finalQuery *goqu.SelectDataset, p InstancesQueryParams) *goqu.SelectDataset {
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
func (api *API) GetInstances(p InstancesQueryParams, duration string) (InstancesWithTotal, error) {
	var instances []*Instance
	var err error
	totalCount, err := api.GetInstancesCount(p, duration)
	if err != nil {
		return InstancesWithTotal{}, err
	}
	p.Page, p.PerPage = validatePaginationParams(p.Page, p.PerPage)
	var dbDuration postgresDuration
	dbDuration, _, err = durationParamToPostgresTimings(durationParam(duration))
	if err != nil {
		return InstancesWithTotal{}, err
	}

	limit, offset := sqlPaginate(p.Page, p.PerPage)
	sortFilter := sanitizeSortFilterParams(p.SortFilter)
	sortOrder := sortOrderFromString(p.SortOrder)
	instancesQuery := api.instancesQuery(p, dbDuration)
	instancesQuery = instancesQuery.Select("id", "ip", "created_ts", goqu.Case().
		When(goqu.C("alias").Neq(""), goqu.C("alias")).Else(goqu.C("id")).As("alias"))

	instanceAppQuery := prepareInstanceAppQuery()
	finalQuery := prepareGetInstancesQuery(instancesQuery, instanceAppQuery)
	if sortOrder == sortOrderAsc {
		finalQuery = finalQuery.Order(goqu.I(sortFilter).Asc().NullsLast())
	} else if sortOrder == sortOrderDesc {
		finalQuery = finalQuery.Order(goqu.I(sortFilter).Desc().NullsLast())
	}

	finalQuery = prepareSearchQuery(finalQuery, p)
	query, _, err := finalQuery.
		Limit(limit).
		Offset(offset).
		ToSQL()
	if err != nil {
		return InstancesWithTotal{}, err
	}
	rows, err := api.db.Queryx(query)
	if err != nil {
		return InstancesWithTotal{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var instance Instance
		err = rows.Scan(&instance.ID, &instance.IP, &instance.CreatedTs, &instance.Alias,
			&instance.Application.Version, &instance.Application.Status, &instance.Application.LastCheckForUpdates,
			&instance.Application.LastUpdateVersion, &instance.Application.UpdateInProgress,
			&instance.Application.ApplicationID, &instance.Application.GroupID, &instance.Application.InstanceID)
		if err != nil {
			return InstancesWithTotal{}, err
		}
		instances = append(instances, &instance)
	}
	if err := rows.Err(); err != nil {
		return InstancesWithTotal{}, err
	}
	result := InstancesWithTotal{
		TotalInstances: uint64(totalCount),
		Instances:      instances,
	}
	return result, nil
}
func prepareInstanceAppQuery() *goqu.SelectDataset {
	return goqu.From("instance_application").
		Select("version", "status", "last_check_for_updates", "last_update_version", "update_in_progress", "application_id", "group_id", "instance_id")
}
func (api *API) GetInstancesCount(p InstancesQueryParams, duration string) (int, error) {
	var err error

	var dbDuration postgresDuration
	dbDuration, _, err = durationParamToPostgresTimings(durationParam(duration))
	if err != nil {
		return 0, err
	}
	instancesQuery := api.instancesQuery(p, dbDuration)
	instancesQuery = instancesQuery.Select("id", "ip", "created_ts", goqu.Case().
		When(goqu.C("alias").Neq(""), goqu.C("alias")).Else(goqu.C("id")).As("alias"))

	instanceAppQuery := prepareInstanceAppQuery()
	finalQuery := prepareGetInstancesQuery(instancesQuery, instanceAppQuery)
	finalQuery = prepareSearchQuery(finalQuery, p).Select(goqu.L("COUNT(*)"))

	return api.GetCountQuery(finalQuery)
}

func (api *API) UpdateInstance(instanceID string, alias string) (*Instance, error) {
	instance := &Instance{}
	query, _, err := goqu.Update("instance").
		Set(
			goqu.Record{
				"alias": alias,
			},
		).
		Where(goqu.C("id").Eq(instanceID)).
		Returning(goqu.T("instance").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// validateApplicationAndGroup validates if the group provided belongs to the
// provided application, returning the normalized uuid version of the appID and
// groupID provided if both are valid and the group belongs to the given
// application, or an error if something goes wrong.
func (api *API) validateApplicationAndGroup(appID, groupID string) (string, string, error) {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return "", "", err
	}
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return "", "", err
	}

	group, err := api.GetGroup(groupID)
	if err != nil {
		return "", "", err
	}

	if group.ApplicationID != appUUID.String() {
		return "", "", ErrInvalidApplicationOrGroup
	}

	return appUUID.String(), groupUUID.String(), nil
}

// updateInstanceStatus updates the status for the provided instance in the
// context of the given application, storing it as well in the instance status
// history registry.
func (api *API) updateInstanceStatus(instanceID, appID string, newStatus int) error {
	instance, err := api.GetInstance(instanceID, appID)
	if err != nil {
		return err
	}
	return api.updateInstanceObjStatus(instance, newStatus)
}

// InstanceStatusUpdateGranted for an instance and version
func (api *API) grantUpdate(instance *Instance, version string) error {
	insertData := make(map[string]interface{})
	insertData["last_update_granted_ts"] = nowUTC()
	insertData["last_update_version"] = version
	insertData["status"] = InstanceStatusUpdateGranted
	insertData["update_in_progress"] = true

	return api.updateInstanceData(instance, insertData)
}

func (api *API) updateInstanceData(instance *Instance, data map[string]interface{}) error {
	appID := instance.Application.ApplicationID

	insertData := data
	newStatus := insertData["status"].(int)

	if instance.Application.Status.Valid && instance.Application.Status.Int64 == int64(newStatus) {
		return nil
	}

	if newStatus == InstanceStatusComplete {
		insertData["version"] = goqu.L("CASE WHEN last_update_version IS NOT NULL THEN last_update_version ELSE version END")
	}

	if newStatus == InstanceStatusComplete || newStatus == InstanceStatusError || newStatus == InstanceStatusUndefined || newStatus == InstanceStatusOnHold {
		insertData["update_in_progress"] = false
	}

	// This insert is used with values returned from the update query that's executed together,
	// so we do one transaction in the DB only.
	// Note: When last_update_version is NULL this fails.
	//       There always has to be a "updateInstanceStatusUpdatedGranted" done first.
	insertQuery, _, err := goqu.Insert("instance_status_history").
		Cols("status", "version", "instance_id", "application_id", "group_id").
		With("inst_app", goqu.Update("instance_application").
			Set(insertData).
			Where(goqu.C("instance_id").Eq(instance.ID), goqu.C("application_id").Eq(appID)).
			Returning("instance_id", "application_id", "last_update_version", "group_id")).
		FromQuery(goqu.From(goqu.L("inst_app")).
			Select(goqu.V(newStatus).As("status"), goqu.C("last_update_version").As("version"), goqu.C("instance_id"), goqu.C("application_id"), goqu.C("group_id"))).
		ToSQL()

	if err != nil {
		return err
	}

	_, err = api.db.Exec(insertQuery)
	return err
}

func (api *API) updateInstanceObjStatus(instance *Instance, newStatus int) error {
	insertData := make(map[string]interface{})
	insertData["status"] = newStatus

	return api.updateInstanceData(instance, insertData)
}

// instanceAppQuery returns a SelectDataset prepared to return the app status
// of the app identified by the application id provided for a given instance.
func (api *API) instanceAppQuery(appID, instanceID string, duration postgresDuration, sortFilter string, orderOfSort sortOrder) *goqu.SelectDataset {
	query := prepareInstanceAppQuery().Where(goqu.C("application_id").Eq(appID)).
		Where(goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", duration))

	if instanceID != "" {
		query = query.Where(goqu.C("instance_id").Eq(instanceID))
	}

	if sortFilter != "" {
		if orderOfSort == sortOrderAsc {
			query = query.Order(goqu.I(sortFilter).Asc().NullsLast())
		} else if orderOfSort == sortOrderDesc {
			query = query.Order(goqu.I(sortFilter).Desc().NullsLast())
		}
	}
	return query
}

func ignoreFakeInstanceCondition(instanceIDField string) string {
	return fmt.Sprintf(`(%[1]s IS NULL OR %[1]s NOT LIKE '{________-____-____-____-____________}')`, instanceIDField)
}

func (api *API) getFilterInstancesQuery(selectPart exp.LiteralExpression, p InstancesQueryParams, duration postgresDuration) *goqu.SelectDataset {
	query := goqu.From("instance_application").
		Select(selectPart).
		Where(goqu.C("application_id").Eq(p.ApplicationID), goqu.C("group_id").Eq(p.GroupID)).
		Where(goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", duration),
			goqu.L(ignoreFakeInstanceCondition("instance_id")))

	if p.Status == InstanceStatusUndefined {
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
func (api *API) instancesQuery(p InstancesQueryParams, duration postgresDuration) *goqu.SelectDataset {
	instancesSubquery := api.getFilterInstancesQuery(goqu.L("instance_id"), p, duration)

	return goqu.From("instance").
		Where(goqu.L("id IN ?", instancesSubquery))
}

// instanceStatusHistoryQuery returns a SelectDataset prepared to return the
// status history of a given instance in the context of an application/group.
func (api *API) instanceStatusHistoryQuery(instanceID, appID, groupID string, limit uint64) *goqu.SelectDataset {
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
func (api *API) GetDefaultInterval() time.Duration {
	return defaultInterval
}

// instanceStatsQuery returns a SelectDataset prepared to return all instances
// that have been checked in during a given duration from a given time.
func (api *API) instanceStatsQuery(t *time.Time, duration *time.Duration) *goqu.SelectDataset {
	if t == nil {
		now := time.Now().UTC()
		t = &now
	}

	if duration == nil {
		d := defaultInterval
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
			goqu.COUNT("*").As("instances")).
		Join(goqu.T("groups"), goqu.On(goqu.C("group_id").Eq(goqu.T("groups").Col("id")))).
		Join(goqu.T("channel"), goqu.On(goqu.T("groups").Col("channel_id").Eq(goqu.T("channel").Col("id")))).
		Where(
			goqu.C("last_check_for_updates").Gt(timestampMinusDuration),
			goqu.C("last_check_for_updates").Lte(timestamp)).
		GroupBy(timestamp,
			goqu.T("channel").Col("name"),
			goqu.T("channel").Col("arch"),
			goqu.C("version")).
		Order(timestamp.Asc())

	return query
}

// GetInstanceStats returns an InstanceStats table with all instances that have
// been previously been checked in.
func (api *API) GetInstanceStats(s *time.Time, t *time.Time) ([]InstanceStats, error) {
	queryBuilder := goqu.From("instance_stats").
		Order(goqu.C("timestamp").Asc())

	if t != nil {
		end := goqu.L("timestamp ?", goqu.V(t.Format("2006-01-02T15:04:05.999999Z07:00")))
		queryBuilder = queryBuilder.Where(goqu.C("timestamp").Lte(end))
	}

	if s != nil {
		start := goqu.L("timestamp ?", goqu.V(s.Format("2006-01-02T15:04:05.999999Z07:00")))
		queryBuilder = queryBuilder.Where(goqu.C("timestamp").Gt(start))
	}

	query, _, err := queryBuilder.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []InstanceStats
	for rows.Next() {
		var instance InstanceStats
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
func (api *API) GetInstanceStatsByTimestamp(t time.Time) ([]InstanceStats, error) {
	timestamp := goqu.L("timestamp ?", goqu.V(t.Format("2006-01-02T15:04:05.999999Z07:00")))

	query, _, err := goqu.From("instance_stats").
		Where(goqu.C("timestamp").Eq(timestamp)).
		Order(goqu.C("version").Asc()).ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []InstanceStats
	for rows.Next() {
		var instance InstanceStats
		err = rows.StructScan(&instance)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// UpdateInstanceStats updates the instance_stats table with instances checked
// in during a given duration from a given time.
func (api *API) UpdateInstanceStats(t *time.Time, duration *time.Duration) error {
	insertQuery, _, err := goqu.Insert(goqu.T("instance_stats")).
		Cols("timestamp", "channel_name", "arch", "version", "instances").
		FromQuery(api.instanceStatsQuery(t, duration)).
		ToSQL()
	if err != nil {
		return err
	}

	_, err = api.db.Exec(insertQuery)
	if err != nil {
		return err
	}
	return nil
}
