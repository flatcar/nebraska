package api

import (
	"errors"
	"time"

	"github.com/blang/semver/v4"
	"gopkg.in/guregu/null.v4"
)

const (
	maxParallelUpdates = 900000
)

var (
	// ErrRegisterInstanceFailed indicates that the instance registration did
	// not succeed.
	ErrRegisterInstanceFailed = errors.New("nebraska: register instance failed")

	// ErrUpdateInProgressOnInstance indicates that an update is currently in
	// progress on the instance requesting an update package, so the request
	// will be rejected.
	ErrUpdateInProgressOnInstance = errors.New("nebraska: update in progress on instance")

	// ErrNoPackageFound indicates that the group doesn't have a channel
	// assigned or that the channel doesn't have a package assigned.
	ErrNoPackageFound = errors.New("nebraska: no package found")

	// ErrNoUpdatePackageAvailable indicates that the instance requesting the
	// update has already the latest version of the application.
	ErrNoUpdatePackageAvailable = errors.New("nebraska: no update package available")

	// ErrUpdatesDisabled indicates that updates are not enabled in the group.
	ErrUpdatesDisabled = errors.New("nebraska: updates disabled")

	// ErrGetUpdatesStatsFailed indicates that there was a problem getting the
	// updates stats of the group which are needed to enforce the rollout
	// policy.
	ErrGetUpdatesStatsFailed = errors.New("nebraska: get updates stats failed")

	// ErrMaxUpdatesPerPeriodLimitReached indicates that the maximum number of
	// updates per period has been reached.
	ErrMaxUpdatesPerPeriodLimitReached = errors.New("nebraska: max updates per period limit reached")

	// ErrMaxConcurrentUpdatesLimitReached indicates that the maximum number of
	// concurrent updates has been reached.
	ErrMaxConcurrentUpdatesLimitReached = errors.New("nebraska: max concurrent updates limit reached")

	// ErrMaxTimedOutUpdatesLimitReached indicates that limit of instances that
	// timed out while updating has been reached.
	ErrMaxTimedOutUpdatesLimitReached = errors.New("nebraska: max timed out updates limit reached")

	// ErrGrantingUpdate indicates that something went wrong while granting an
	// update.
	ErrGrantingUpdate = errors.New("nebraska: error granting update")
)

// GetUpdatePackage returns an update package for the instance/application
// provided. The instance details and the application it's running will be
// registered in Nebraska (or updated if it's already registered).
func (api *API) GetUpdatePackage(instanceID, instanceAlias, instanceIP, instanceVersion, appID, groupID string) (*Package, error) {
	group, err := api.GetGroup(groupID)
	if err != nil {
		logger.Error().Msgf("GetUpdatePackage - failed to get group with id %v on app %v. Not registering/updating instance with ID %v (alias=%v)", groupID, appID, instanceID, instanceAlias)
		return nil, err
	}

	instance, err := api.GetInstance(instanceID, appID)
	if err != nil {
		logger.Info().Msgf("GetUpdatePackage - instance %v (alias=%v) not yet registered.", instanceID, instanceAlias)
		instance = &Instance{
			ID: instanceID,
			IP: instanceIP,
			Application: InstanceApplication{
				InstanceID:    instanceID,
				ApplicationID: appID,
				GroupID:       null.StringFrom(groupID),
				Version:       instanceVersion,
			},
			Alias: instanceAlias,
		}
	}

	updateAlreadyGranted := false

	if instance.Application.Status.Valid {
		switch int(instance.Application.Status.Int64) {
		case InstanceStatusDownloading, InstanceStatusDownloaded, InstanceStatusInstalled:
			if _, err := api.upsertInstance(instance); err != nil {
				logger.Error().Err(err).Msgf("GetUpdatePackage - failed to register instance %v", instance.ID)
				return nil, err
			}
			return nil, ErrUpdateInProgressOnInstance
		case InstanceStatusUpdateGranted:
			updateAlreadyGranted = true
		}
	}

	if group.Channel == nil || group.Channel.Package == nil {
		if _, err := api.upsertInstance(instance); err != nil {
			logger.Error().Err(err).Msgf("GetUpdatePackage - failed to register instance %v", instance.ID)
			return nil, err
		}
		if err := api.newGroupActivityEntry(activityPackageNotFound, activityWarning, "0.0.0", appID, groupID); err != nil {
			logger.Error().Err(err).Msg("GetUpdatePackage - could not add new group activity entry")
		}
		return nil, ErrNoPackageFound
	}

	for _, blacklistedChannelID := range group.Channel.Package.ChannelsBlacklist {
		if blacklistedChannelID == group.Channel.ID {
			if updateAlreadyGranted {
				// This logic needs to be reviewed as it doesn't make sense to set the instance as completed when its
				// channel is rather in the blocklist.
				instance.Application.Status = null.IntFrom(int64(InstanceStatusComplete))
				if _, err := api.upsertInstance(instance); err != nil {
					logger.Error().Err(err).Msgf("GetUpdatePackage - could to register/update instance with complete status %v", instance.ID)
					return nil, err
				}
			}
			return nil, ErrNoUpdatePackageAvailable
		}
	}

	instanceSemver, _ := semver.Make(instanceVersion)
	packageSemver, _ := semver.Make(group.Channel.Package.Version)
	if !instanceSemver.LT(packageSemver) {
		if updateAlreadyGranted {
			instance.Application.Status = null.IntFrom(int64(InstanceStatusComplete))
		}
		if _, err := api.upsertInstance(instance); err != nil {
			logger.Error().Err(err).Msgf("GetUpdatePackage - could to register/update instance with complete status %v", instance.ID)
			return nil, err
		}
		return nil, ErrNoUpdatePackageAvailable
	}

	if updateAlreadyGranted {
		if _, err := api.upsertInstance(instance); err != nil {
			logger.Error().Err(err).Msgf("GetUpdatePackage - could to register/update instance %v", instance.ID)
			return nil, err
		}

		return group.Channel.Package, nil
	}

	if err := api.enforceRolloutPolicy(instance, group); err != nil {
		return nil, err
	}

	version := group.Channel.Package.Version

	instance.Application.LastUpdateGrantedTs = null.TimeFrom(time.Now().UTC())
	instance.Application.LastUpdateVersion = null.StringFrom(version)
	instance.Application.Status = null.IntFrom(int64(InstanceStatusUpdateGranted))
	instance.Application.UpdateInProgress = true

	if _, err := api.upsertInstance(instance); err != nil {
		logger.Error().Err(err).Msgf("GetUpdatePackage - grantUpdate error for instance %v:", instanceID)
		return nil, err
	}
	// if instance == nil {
	// 	grantData := api.makeUpdateGrantInfo(version)
	// 	_, err = api.RegisterInstanceWithData(instanceID, instanceAlias, instanceIP, instanceVersion, appID, groupID, grantData)
	// 	if err != nil {
	// 		logger.Error().Err(err).Msg("GetUpdatePackage - could not register instance (propagates as ErrRegisterInstanceFailed)")
	// 		return nil, ErrRegisterInstanceFailed
	// 	}
	// } else if err := api.grantUpdate(instance, version); err != nil {
	// 	logger.Error().Err(err).Msg("GetUpdatePackage - grantUpdate error (propagates as ErrGrantingUpdate):")
	// }

	if !api.hasRecentActivity(activityRolloutStarted, ActivityQueryParams{Severity: activityInfo, AppID: appID, Version: version, GroupID: group.ID}) {
		if err := api.newGroupActivityEntry(activityRolloutStarted, activityInfo, version, appID, group.ID); err != nil {
			logger.Error().Err(err).Msg("GetUpdatePackage - could not add new group activity entry")
		}
	}

	if !group.RolloutInProgress {
		if err := api.setGroupRolloutInProgress(groupID, true); err != nil {
			logger.Error().Err(err).Msg("GetUpdatePackage - could not set rollout progress")
		}
	}

	return group.Channel.Package, nil
}

// enforceRolloutPolicy validates if an update should be provided to the
// requesting instance based on the group rollout policy and the current status
// of the updates taking place in the group.
func (api *API) enforceRolloutPolicy(instance *Instance, group *Group) error {
	updateStatusAndReturn := func(errReturn error) error {
		instance.Application.Status = null.IntFrom(int64(InstanceStatusOnHold))
		if _, err := api.upsertInstance(instance); err != nil {
			logger.Error().Err(err).Msg("enforceRolloutPolicy - could not update instance status")
		}

		return errReturn
	}

	if !group.PolicyUpdatesEnabled {
		return ErrUpdatesDisabled
	}

	if group.PolicyOfficeHours && !inOfficeHoursNow(group.PolicyTimezone.String) {
		return ErrUpdatesDisabled
	}

	effectiveMaxUpdates := group.PolicyMaxUpdatesPerPeriod

	// If no policy enforcement is needed, then we skip getting the update stats below.
	if effectiveMaxUpdates >= maxParallelUpdates && !group.PolicySafeMode {
		return nil
	}

	updatesStats, err := api.getGroupUpdatesStats(group)
	if err != nil {
		logger.Error().Err(err).Msg("GetUpdatePackage - getGroupUpdatesStats error (propagates as ErrGetUpdatesStatsFailed):")
		return ErrGetUpdatesStatsFailed
	}

	if group.PolicySafeMode && updatesStats.UpdatesToCurrentVersionAttempted == 0 {
		effectiveMaxUpdates = 1
	}

	if updatesStats.UpdatesGrantedInLastPeriod >= effectiveMaxUpdates {
		return updateStatusAndReturn(ErrMaxUpdatesPerPeriodLimitReached)
	}

	if updatesStats.UpdatesInProgress >= effectiveMaxUpdates {
		return updateStatusAndReturn(ErrMaxConcurrentUpdatesLimitReached)
	}

	if group.PolicySafeMode && updatesStats.UpdatesTimedOut >= effectiveMaxUpdates {
		if group.PolicyUpdatesEnabled {
			if err := api.disableUpdates(group.ID); err != nil {
				logger.Error().Err(err).Msg("enforceRolloutPolicy - could not disable updates")
			}
		}
		return updateStatusAndReturn(ErrMaxTimedOutUpdatesLimitReached)
	}

	return nil
}

// inOfficeHoursNow checks if the provided timezone is now in office hours.
func inOfficeHoursNow(tz string) bool {
	if tz == "" {
		return false
	}

	location, err := time.LoadLocation(tz)
	if err != nil {
		return false
	}

	now := time.Now().In(location)
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}
	if now.Hour() < 9 || now.Hour() >= 17 {
		return false
	}

	return true
}
