package runtime

import (
	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

const flatcarAppID = types.FlatcarAppID

// RegisterEvent registers an event posted by an instance in Nebraska. The
// event will be bound to an application/group combination.
func (s *Service) RegisterEvent(instanceID, appID, groupID string, etype, eresult int, previousVersion, errorCode string) error {
	var err error
	if appID, groupID, err = s.validateApplicationAndGroup(appID, groupID); err != nil {
		return err
	}
	instance, err := s.GetInstance(instanceID, appID)
	if err != nil {
		l.Info().Err(err).Msg("RegisterEvent - could not get instance, maybe it is a first contact (propagates as ErrInvalidInstance)")
		return types.ErrInvalidInstance
	}
	if instance.Application.ApplicationID != appID {
		return types.ErrInvalidApplicationOrGroup
	}
	if !instance.Application.UpdateInProgress {
		// Do not log the event when we don't know about an update going on.
		// There is no need to reset the instance state here because update_in_progress
		// is only set to "false" for states in which we will grant an update.
		return types.ErrNoUpdateInProgress
	}

	// Temporary hack to handle Flatcar updater specific behaviour
	if appID == flatcarAppID && etype == types.EventUpdateComplete && eresult == types.ResultSuccessReboot {
		if previousVersion == "" || previousVersion == "0.0.0.0" {
			// Do not log the Complete event for already updated instances but reset the instance state to
			// ensure it can update and is not stuck in some other state because according to the DB it,
			// e.g., is updating and thus shouldn't be granted any update. The instance can't be in a Completed
			// state because of the ErrNoUpdateInProgress check above, thus no need to cover this case here.
			// The Undefined state is chosen because the instance did not tell that it updated from a previous
			// version ("" and "0.0.0.0" are not valid but "0.0.0" is because it is used when forcing an update).
			if err := s.updateInstanceObjStatus(instance, types.InstanceStatusUndefined); err != nil {
				l.Error().Err(err).Msg("RegisterEvent - could not update instance status")
			}
			return types.ErrFlatcarEventIgnored
		}
	}

	var eventTypeID int
	query, _, err := goqu.From("event_type").
		Select("id").
		Where(goqu.C("type").Eq(etype), goqu.C("result").Eq(eresult)).
		ToSQL()
	if err != nil {
		return err
	}
	err = s.db.QueryRow(query).Scan(&eventTypeID)
	if err != nil {
		return types.ErrInvalidEventTypeOrResult
	}

	insertQuery, _, err := goqu.Insert("event").
		Cols("event_type_id", "instance_id", "application_id", "previous_version", "error_code").
		Vals(goqu.Vals{eventTypeID, instanceID, appID, previousVersion, errorCode}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(insertQuery)

	if err != nil {
		return types.ErrEventRegistrationFailed
	}

	lastUpdateVersion := instance.Application.LastUpdateVersion.String
	if err := s.triggerEventConsequences(instanceID, appID, groupID, lastUpdateVersion, etype, eresult); err != nil {
		l.Error().Err(err).Msgf("RegisterEvent - could not trigger event consequences")
	}

	return nil
}

// triggerEventConsequences is in charge of triggering the consequences of a
// given event. Depending on the type of the event and its result, the status
// of the instance may be updated, new activity entries could be created, etc.
func (s *Service) triggerEventConsequences(instanceID, appID, groupID, lastUpdateVersion string, etype, result int) error {
	group, err := s.GetGroup(groupID)
	if err != nil {
		return err
	}

	// We allow the plain ResultSuccess here only if the app is not Flatcar because Flatcar is relying on
	// having only the update-complete logic on ResultSuccessReboot.
	if etype == types.EventUpdateComplete && (result == types.ResultSuccessReboot || (appID != flatcarAppID && result == types.ResultSuccess)) {
		if err := s.updateInstanceStatus(instanceID, appID, types.InstanceStatusComplete); err != nil {
			l.Error().Err(err).Msg("triggerEventConsequences - could not update instance status")
		}

		updatesStats, err := s.GetGroupUpdatesStats(group)
		if err != nil {
			return err
		}
		if updatesStats.UpdatesToCurrentVersionSucceeded == updatesStats.TotalInstances {
			if err := s.setGroupRolloutInProgress(groupID, false); err != nil {
				l.Error().Err(err).Msg("triggerEventConsequences - could not set rollout progress")
			}
			if err := s.newGroupActivityEntry(types.ActivityRolloutFinished, types.ActivitySuccess, lastUpdateVersion, appID, groupID); err != nil {
				l.Error().Err(err).Msg("triggerEventConsequences - could not add group activity")
			}
		}
	}

	if etype == types.EventUpdateDownloadStarted && result == types.ResultSuccess {
		if err := s.updateInstanceStatus(instanceID, appID, types.InstanceStatusDownloading); err != nil {
			l.Error().Err(err).Msg("triggerEventConsequences - could not update instance status")
		}
	}

	if etype == types.EventUpdateDownloadFinished && result == types.ResultSuccess {
		if err := s.updateInstanceStatus(instanceID, appID, types.InstanceStatusDownloaded); err != nil {
			l.Error().Err(err).Msg("triggerEventConsequences - could not update instance status")
		}
	}

	if etype == types.EventUpdateInstalled && result == types.ResultSuccess {
		if err := s.updateInstanceStatus(instanceID, appID, types.InstanceStatusInstalled); err != nil {
			l.Error().Err(err).Msg("triggerEventConsequences - could not update instance status")
		}
	}

	if result == types.ResultFailed {
		if err := s.updateInstanceStatus(instanceID, appID, types.InstanceStatusError); err != nil {
			l.Error().Err(err).Msg("triggerEventConsequences - could not update instance status")
		}
		if err := s.newInstanceActivityEntry(types.ActivityInstanceUpdateFailed, types.ActivityError, lastUpdateVersion, appID, groupID, instanceID); err != nil {
			l.Error().Err(err).Msg("triggerEventConsequences - could not add instance activity")
		}

		if s.disableUpdatesOnFailedRollout {
			updatesStats, err := s.GetGroupUpdatesStats(group)
			if err != nil {
				return err
			}
			if updatesStats.UpdatesToCurrentVersionAttempted == 1 {
				if err := s.disableUpdates(groupID); err != nil {
					l.Error().Err(err).Msg("triggerEventConsequences - could not disable updates")
				}
				if err := s.setGroupRolloutInProgress(groupID, false); err != nil {
					l.Error().Err(err).Msg("triggerEventConsequences - could not set rollout progress")
				}
				if err := s.newGroupActivityEntry(types.ActivityRolloutFailed, types.ActivityError, lastUpdateVersion, appID, groupID); err != nil {
					l.Error().Err(err).Msg("triggerEventConsequences - could not add group activity")
				}
			}
		}
	}

	return nil
}
