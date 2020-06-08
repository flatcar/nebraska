package api

import (
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"
)

const (
	// EventUpdateComplete indicates that the update process completed. It could
	// mean a successful or failed updated, depending on the result attached to
	// the event. This applies to all events.
	EventUpdateComplete = 3

	// EventUpdateDownloadStarted indicates that the instance started
	// downloading the update package.
	EventUpdateDownloadStarted = 13

	// EventUpdateDownloadFinished indicates that the update package was
	// downloaded.
	EventUpdateDownloadFinished = 14

	// EventUpdateInstalled indicates that the update package was installed.
	EventUpdateInstalled = 800
)

const (
	// ResultFailed indicates that the operation associated with the event
	// posted failed.
	ResultFailed = 0

	// ResultSuccess indicates that the operation associated with the event
	// posted succeeded.
	ResultSuccess = 1

	// ResultSuccessReboot also indicates a successful operation, but it's
	// meant only to be used along with events of EventUpdateComplete type.
	// It's important that instances use EventUpdateComplete events in
	// combination with ResultSuccessReboot to communicate a successful update
	// completed as it has a special meaning for Nebraska in order to adjust
	// properly the rollout policies and create activity entries.
	ResultSuccessReboot = 2
)

var (
	// ErrInvalidInstance indicates that the instance provided is not valid or
	// it doesn't exist.
	ErrInvalidInstance = errors.New("nebraska: invalid instance")

	// ErrInvalidApplicationOrGroup indicates that the application or group id
	// provided are not valid or related to each other.
	ErrInvalidApplicationOrGroup = errors.New("nebraska: invalid application or group")

	// ErrInvalidEventTypeOrResult indicates that the event or result provided
	// are not valid (Nebraska only implements a subset of the Omaha protocol
	// events).
	ErrInvalidEventTypeOrResult = errors.New("nebraska: invalid event type or result")

	// ErrEventRegistrationFailed indicates that the event registration into
	// Nebraska failed.
	ErrEventRegistrationFailed = errors.New("nebraska: event registration failed")

	// ErrNoUpdateInProgress indicates that an event was received but there
	// wasn't an update in progress for the provided instance/application, so
	// it was rejected.
	ErrNoUpdateInProgress = errors.New("nebraska: no update in progress")

	// ErrFlatcarEventIgnored indicates that a Flatcar updater event was ignored.
	// This is a temporary solution to handle Flatcar specific behaviour.
	ErrFlatcarEventIgnored = errors.New("nebraska: flatcar event ignored")
)

// Event represents an event posted by an instance to Nebraska.
type Event struct {
	ID              int         `db:"id" json:"id"`
	CreatedTs       time.Time   `db:"created_ts" json:"created_ts"`
	PreviousVersion null.String `db:"previous_version" json:"previous_version"`
	ErrorCode       null.String `db:"error_code" json:"error_code"`
	InstanceID      string      `db:"instance_id" json:"instance_id"`
	ApplicationID   string      `db:"application_id" json:"application_id"`
	EventTypeID     string      `db:"event_type_id" json:"event_type_id"`
}

// RegisterEvent registers an event posted by an instance in Nebraska. The
// event will be bound to an application/group combination.
func (api *API) RegisterEvent(instanceID, appID, groupID string, etype, eresult int, previousVersion, errorCode string) error {
	var err error
	if appID, groupID, err = api.validateApplicationAndGroup(appID, groupID); err != nil {
		return err
	}
	instance, err := api.GetInstance(instanceID, appID)
	if err != nil {
		return ErrInvalidInstance
	}
	if instance.Application.ApplicationID != appID {
		return ErrInvalidApplicationOrGroup
	}
	if !instance.Application.UpdateInProgress {
		return ErrNoUpdateInProgress
	}

	// Temporary hack to handle Flatcar updater specific behaviour
	if appID == flatcarAppID && etype == EventUpdateComplete && eresult == ResultSuccessReboot {
		if previousVersion == "" || previousVersion == "0.0.0.0" || previousVersion != instance.Application.Version {
			return ErrFlatcarEventIgnored
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
	err = api.db.QueryRow(query).Scan(&eventTypeID)
	if err != nil {
		return ErrInvalidEventTypeOrResult
	}

	insertQuery, _, err := goqu.Insert("event").
		Cols("event_type_id", "instance_id", "application_id", "previous_version", "error_code").
		Vals(goqu.Vals{eventTypeID, instanceID, appID, previousVersion, errorCode}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(insertQuery)

	if err != nil {
		return ErrEventRegistrationFailed
	}

	lastUpdateVersion := instance.Application.LastUpdateVersion.String
	_ = api.triggerEventConsequences(instanceID, appID, groupID, lastUpdateVersion, etype, eresult)

	return nil
}

// triggerEventConsequences is in charge of triggering the consequences of a
// given event. Depending on the type of the event and its result, the status
// of the instance may be updated, new activity entries could be created, etc.
func (api *API) triggerEventConsequences(instanceID, appID, groupID, lastUpdateVersion string, etype, result int) error {
	group, err := api.GetGroup(groupID)
	if err != nil {
		return err
	}

	// TODO: should we also consider ResultSuccess in the next check? Flatcar ~ generic conflicts?
	if etype == EventUpdateComplete && result == ResultSuccessReboot {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusComplete)

		updatesStats, err := api.getGroupUpdatesStats(group)
		if err != nil {
			return err
		}
		if updatesStats.UpdatesToCurrentVersionSucceeded == updatesStats.TotalInstances {
			_ = api.setGroupRolloutInProgress(groupID, false)
			_ = api.newGroupActivityEntry(activityRolloutFinished, activitySuccess, lastUpdateVersion, appID, groupID)
		}
	}

	if etype == EventUpdateDownloadStarted && result == ResultSuccess {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusDownloading)
	}

	if etype == EventUpdateDownloadFinished && result == ResultSuccess {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusDownloaded)
	}

	if etype == EventUpdateInstalled && result == ResultSuccess {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusInstalled)
	}

	if result == ResultFailed {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusError)
		_ = api.newInstanceActivityEntry(activityInstanceUpdateFailed, activityError, lastUpdateVersion, appID, groupID, instanceID)

		updatesStats, err := api.getGroupUpdatesStats(group)
		if err != nil {
			return err
		}
		if updatesStats.UpdatesToCurrentVersionAttempted == 1 {
			if api.disableUpdatesOnFailedRollout {
				_ = api.disableUpdates(groupID)
				_ = api.setGroupRolloutInProgress(groupID, false)
				_ = api.newGroupActivityEntry(activityRolloutFailed, activityError, lastUpdateVersion, appID, groupID)
			}
		}
	}

	return nil
}
