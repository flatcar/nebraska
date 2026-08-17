package runtime

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func nowUTC() time.Time {
	return time.Now().UTC()
}

// NewInstanceApplication creates an InstanceApplication with the fields used for registration.
func NewInstanceApplication(appID, groupID, version string) types.InstanceApplication {
	return types.InstanceApplication{ApplicationID: appID, GroupID: null.StringFrom(groupID), Version: version}
}

// RegisterInstance registers an instance into Nebraska.
func (s *Service) RegisterInstance(inst types.Instance, instApp types.InstanceApplication) (*types.Instance, error) {
	if !dbreads.IsValidSemver(instApp.Version) {
		return nil, types.ErrInvalidSemver
	}

	appID := instApp.ApplicationID
	groupID := instApp.GroupID.String

	var err error
	if appID, groupID, err = s.validateApplicationAndGroup(appID, groupID); err != nil {
		return nil, err
	}

	instanceAlias := inst.Alias
	instanceOEM := inst.OEM
	instanceAlephVersion := inst.AlephVersion

	// We want to avoid having to create an unneeded DB transaction, so we check whether it
	// is necessary (we need it when writing into the two tables, instance and
	// instance_application).

	updateInstance := true
	updateInstanceApplication := true

	instance, err := s.GetInstance(inst.ID, appID)
	if err == nil {
		// Give precedence to an existing alias over an omitted or empty alias field
		if instanceAlias == "" {
			instanceAlias = instance.Alias
		}
		// Give precedence to existing OEM values over omitted or empty fields
		if instanceOEM == "" {
			instanceOEM = instance.OEM
		}
		if instanceAlephVersion == "" {
			instanceAlephVersion = instance.AlephVersion
		}
		// The instance exists, so we just update it if its IP, Alias, OEM or AlephVersion changed
		updateInstance = instance.IP != inst.IP || instance.Alias != instanceAlias || instance.OEM != instanceOEM || instance.AlephVersion != instanceAlephVersion

		recent := nowUTC().Add(-5 * time.Minute)

		// And we only update the instance_application if the latest registry is outdated or
		// older than what we establish as recent.
		updateInstanceApplication = instance.Application.LastCheckForUpdates.UTC().Before(recent) ||
			instance.Application.Version != instApp.Version || instance.Application.GroupID.String != groupID

		// Skip updating anything unnecessary
		if !updateInstance && !updateInstanceApplication {
			return instance, nil
		}
	}

	upsertInstance, _, err := goqu.Insert("instance").
		Cols("id", "ip", "alias", "oem", "aleph_version").
		Vals(goqu.Vals{inst.ID, inst.IP, instanceAlias, instanceOEM, instanceAlephVersion}).
		OnConflict(goqu.DoUpdate("id", goqu.Record{"id": inst.ID, "ip": inst.IP, "alias": instanceAlias, "oem": instanceOEM, "aleph_version": instanceAlephVersion})).
		ToSQL()
	if err != nil {
		return nil, err
	}

	upsertInstanceApplication, _, err := goqu.Insert("instance_application").
		Cols("instance_id", "application_id", "group_id", "version", "last_check_for_updates").
		Vals(goqu.Vals{inst.ID, appID, groupID, instApp.Version, nowUTC()}).
		OnConflict(goqu.DoUpdate("ON CONSTRAINT instance_application_pkey", goqu.Record{"group_id": groupID, "version": instApp.Version, "last_check_for_updates": nowUTC()})).
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
		_, err := s.db.Exec(queryToExec)
		if err != nil {
			return nil, err
		}

		return instance, nil
	}

	// If this is an instance we haven't seen yet, then we write into instance + instance_application
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("RegisterInstance - could not roll back")
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
	return s.GetInstance(inst.ID, appID)
}

func (s *Service) UpdateInstance(instanceID string, alias string) (*types.Instance, error) {
	instance := &types.Instance{}
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
	err = s.db.QueryRowx(query).StructScan(instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// validateApplicationAndGroup validates if the group provided belongs to the
// provided application, returning the normalized uuid version of the appID and
// groupID provided if both are valid and the group belongs to the given
// application, or an error if something goes wrong.
func (s *Service) validateApplicationAndGroup(appID, groupID string) (string, string, error) {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return "", "", err
	}
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return "", "", err
	}

	group, err := s.GetGroup(groupID)
	if err != nil {
		return "", "", err
	}

	if group.ApplicationID != appUUID.String() {
		return "", "", types.ErrInvalidApplicationOrGroup
	}

	return appUUID.String(), groupUUID.String(), nil
}

// updateInstanceStatus updates the status for the provided instance in the
// context of the given application, storing it as well in the instance status
// history registry.
func (s *Service) updateInstanceStatus(instanceID, appID string, newStatus int) error {
	instance, err := s.GetInstance(instanceID, appID)
	if err != nil {
		return err
	}
	return s.updateInstanceObjStatus(instance, newStatus)
}

// InstanceStatusUpdateGranted for an instance and version
func (s *Service) grantUpdate(instance *types.Instance, version string) error {
	insertData := make(map[string]interface{})
	insertData["last_update_granted_ts"] = nowUTC()
	insertData["last_update_version"] = version
	insertData["status"] = types.InstanceStatusUpdateGranted
	insertData["update_in_progress"] = true

	return s.updateInstanceData(instance, insertData)
}

func (s *Service) updateInstanceData(instance *types.Instance, data map[string]interface{}) error {
	appID := instance.Application.ApplicationID

	insertData := data
	newStatus := insertData["status"].(int)

	if instance.Application.Status.Valid && instance.Application.Status.Int64 == int64(newStatus) {
		return nil
	}

	if newStatus == types.InstanceStatusComplete {
		insertData["version"] = goqu.L("CASE WHEN last_update_version IS NOT NULL THEN last_update_version ELSE version END")
	}

	if newStatus == types.InstanceStatusComplete || newStatus == types.InstanceStatusError || newStatus == types.InstanceStatusUndefined || newStatus == types.InstanceStatusOnHold {
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

	_, err = s.db.Exec(insertQuery)
	return err
}

func (s *Service) updateInstanceObjStatus(instance *types.Instance, newStatus int) error {
	insertData := make(map[string]interface{})
	insertData["status"] = newStatus

	return s.updateInstanceData(instance, insertData)
}

// UpdateInstanceStats updates the instance_stats table with instances checked
// in during a given duration from a given time.
func (s *Service) UpdateInstanceStats(t *time.Time, duration *time.Duration) error {
	insertQuery, _, err := goqu.Insert(goqu.T("instance_stats")).
		Cols("timestamp", "channel_name", "arch", "version", "instances").
		FromQuery(s.InstanceStatsQuery(t, duration)).
		ToSQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(insertQuery)
	if err != nil {
		return err
	}
	return nil
}
