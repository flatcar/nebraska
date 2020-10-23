package api

import (
	"database/sql"
	"fmt"
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
)

// Instance represents an instance running one or more applications for which
// Nebraska can provide updates.
type Instance struct {
	ID          string              `db:"id" json:"id"`
	IP          string              `db:"ip" json:"ip"`
	CreatedTs   time.Time           `db:"created_ts" json:"created_ts"`
	Application InstanceApplication `db:"application" json:"application,omitempty"`
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
}

// RegisterInstance registers an instance into Nebraska.
func (api *API) RegisterInstance(instanceID, instanceIP, instanceVersion, appID, groupID string) (*Instance, error) {
	if !isValidSemver(instanceVersion) {
		return nil, ErrInvalidSemver
	}
	var err error
	if appID, groupID, err = api.validateApplicationAndGroup(appID, groupID); err != nil {
		return nil, err
	}
	tx, err := api.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logger.Error("RegisterInstance - could not roll back", err)
		}
	}()
	query, _, err := goqu.Insert("instance").
		Cols("id", "ip").
		Vals(goqu.Vals{instanceID, instanceIP}).
		OnConflict(goqu.DoUpdate("id", goqu.Record{"id": instanceID, "ip": instanceIP})).
		ToSQL()
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(query)
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
	upsertQuery, _, err := goqu.Insert("instance_application").
		Cols("instance_id", "application_id", "group_id", "version", "last_check_for_updates").
		Vals(goqu.Vals{instanceID, appID, groupID, instanceVersion, nowUTC()}).
		OnConflict(goqu.DoUpdate("ON CONSTRAINT instance_application_pkey", goqu.Record{"group_id": groupID, "version": instanceVersion, "last_check_for_updates": nowUTC()})).
		ToSQL()
	if err != nil {
		return nil, err
	}
	result, err = tx.Exec(upsertQuery)
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
	instanceApplication, err := api.getInstanceApp(appID, instance.ID, validityInterval)
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

func (api *API) getInstanceApp(appID, instanceID string, duration postgresDuration) (*InstanceApplication, error) {
	var instanceApp InstanceApplication
	query, _, err := api.instanceAppQuery(appID, instanceID, duration).ToSQL()
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
	query, _, err := api.instancesQuery(p, dbDuration).
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
		err := rows.StructScan(&instance)
		if err != nil {
			return InstancesWithTotal{}, err
		}
		application, err := api.getInstanceApp(p.ApplicationID, instance.ID, dbDuration)
		switch err {
		case nil:
			instance.Application = *application
		case sql.ErrNoRows:
			instance.Application = InstanceApplication{}
		default:
			return InstancesWithTotal{}, err
		}
		instances = append(instances, &instance)
	}
	if err := rows.Err(); err != nil {
		return InstancesWithTotal{}, err
	}
	result := InstancesWithTotal{
		TotalInstances: totalCount,
		Instances:      instances,
	}
	return result, nil
}

func (api *API) GetInstancesCount(p InstancesQueryParams, duration string) (uint64, error) {
	var totalCount uint64
	var err error

	var dbDuration postgresDuration
	dbDuration, _, err = durationParamToPostgresTimings(durationParam(duration))
	if err != nil {
		return 0, err
	}
	countQuery, _, err := api.getFilterInstancesQuery(goqu.L("COUNT (*)"), p, dbDuration).ToSQL()
	if err != nil {
		return 0, err
	}
	err = api.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return 0, err
	}
	return totalCount, nil
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
	if instance.Application.Status.Valid && instance.Application.Status.Int64 == int64(newStatus) {
		return nil
	}
	var insertData = make(map[string]interface{})
	insertData["status"] = newStatus
	if newStatus == InstanceStatusComplete {
		insertData["version"] = goqu.L("CASE WHEN last_update_version IS NOT NULL THEN last_update_version ELSE version END")
	}

	if newStatus == InstanceStatusComplete || newStatus == InstanceStatusError {
		insertData["update_in_progress"] = false
	}

	// We update the instance_application in this query but use it together with the insert below
	updateQuery, _, err := goqu.Update("instance_application").
		Set(insertData).
		Where(goqu.C("instance_id").Eq(instanceID), goqu.C("application_id").Eq(appID)).
		Returning("instance_id", "application_id", "last_update_version", "group_id").
		ToSQL()
	if err != nil {
		return err
	}

	const helperTableName = "inst_app"

	// This insert is used with values returned from the update call, so we do one transaction
	// in the DB only.
	insertQuery, _, err := goqu.Insert("instance_status_history").
		Cols("status", "version", "instance_id", "application_id", "group_id").
		FromQuery(goqu.From(goqu.L(helperTableName)).
			Select(goqu.V(newStatus).As("status"), goqu.C("last_update_version").As("version"), goqu.C("instance_id"), goqu.C("application_id"), goqu.C("group_id"))).
		ToSQL()

	if err != nil {
		return err
	}

	// @todo: Unfortunately goqu doesn't seem to support "WITH" statements when they are used
	// together with an "INSERT"; hence this semi-manual use of SQL below.
	_, err = api.db.Exec(fmt.Sprintf("WITH %1s AS (%2s) %3s", helperTableName, updateQuery, insertQuery))
	return err
}

// instanceAppQuery returns a SelectDataset prepared to return the app status
// of the app identified by the application id provided for a given instance.
func (api *API) instanceAppQuery(appID, instanceID string, duration postgresDuration) *goqu.SelectDataset {
	query := goqu.From("instance_application").
		Select("version", "status", "last_check_for_updates", "last_update_version", "update_in_progress", "application_id", "group_id").
		Where(goqu.C("instance_id").Eq(instanceID), goqu.C("application_id").Eq(appID)).
		Where(goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", duration)).
		Where(goqu.L(ignoreFakeInstanceCondition("instance_id")))
	return query
}

func ignoreFakeInstanceCondition(instanceIDField string) string {
	return fmt.Sprintf(`(%[1]s IS NULL OR %[1]s NOT SIMILAR TO '\{[a-fA-F0-9-]{36}\}')`, instanceIDField)
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
