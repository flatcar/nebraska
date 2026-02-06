package api

import (
	"errors"
	"slices"
	"time"

	"github.com/blang/semver/v4"
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

	// ErrUpdateGrantFailed indicates that the update could not be granted
	// due to a database or internal error.
	ErrUpdateGrantFailed = errors.New("nebraska: failed to grant update")

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
func (api *API) GetUpdatePackage(instanceID, instanceAlias, instanceIP, instanceVersion, appID, groupID, instanceOEM, instanceAlephVersion string) (*Package, error) {
	instance, err := api.RegisterInstance(instanceID, instanceAlias, instanceIP, instanceVersion, appID, groupID, instanceOEM, instanceAlephVersion)
	if err != nil {
		l.Error().Err(err).Msg("GetUpdatePackage - could not register instance")
		return nil, ErrRegisterInstanceFailed
	}

	if instance.Application.Status.Valid {
		switch int(instance.Application.Status.Int64) {
		case InstanceStatusDownloading, InstanceStatusDownloaded, InstanceStatusInstalled:
			return nil, ErrUpdateInProgressOnInstance
		}
	}

	group, err := api.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	if group.Channel == nil || group.Channel.Package == nil {
		if err := api.newGroupActivityEntry(activityPackageNotFound, activityWarning, "0.0.0", appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not add new group activity entry")
		}
		return nil, ErrNoPackageFound
	}

	// Handle already-granted updates
	if instance.Application.Status.Valid && int(instance.Application.Status.Int64) == InstanceStatusUpdateGranted {
		// Check if target package is blacklisted
		if slices.Contains(group.Channel.Package.ChannelsBlacklist, group.Channel.ID) {
			if err := api.updateInstanceObjStatus(instance, InstanceStatusComplete); err != nil {
				l.Error().Err(err).Msg("GetUpdatePackage - could not update instance status")
			}
			return nil, ErrNoUpdatePackageAvailable
		}

		// Check if instance has reached the granted version (not the target!)
		// This allows proper progression through floors
		if instance.Application.LastUpdateVersion.Valid && instance.Application.LastUpdateVersion.String != "" {
			grantedVersion := instance.Application.LastUpdateVersion.String
			instanceSemver, _ := semver.Make(instanceVersion)
			grantedSemver, _ := semver.Make(grantedVersion)

			if !instanceSemver.LT(grantedSemver) {
				// Instance has reached or passed the granted version (floor or target)
				// Complete this grant so they can request the next floor
				if err := api.updateInstanceObjStatus(instance, InstanceStatusComplete); err != nil {
					l.Error().Err(err).Msg("GetUpdatePackage - could not update instance status")
				}
				return nil, ErrNoUpdatePackageAvailable
			}

			// Instance hasn't reached granted version yet - return what's next to install
			// This will be the first floor/target above current instance version
			packages, err := api.getPackagesWithFloorsForUpdate(group, instanceVersion)
			if err != nil {
				return nil, err
			}
			// packages[0] should be the granted version since instance < granted
			return packages[0], nil
		}

		// No granted version tracked (old instances) - safer fallback
		// Complete the update to force a fresh grant cycle with all checks
		l.Warn().Str("instanceID", instance.ID).Msg("Already-granted without LastUpdateVersion - completing to force fresh grant")
		if err := api.updateInstanceObjStatus(instance, InstanceStatusComplete); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not update instance status")
		}
		return nil, ErrNoUpdatePackageAvailable
		// Instance will call again without already-granted status and go through proper checks
	}

	// Check if update is needed
	instanceSemver, _ := semver.Make(instanceVersion)
	packageSemver, _ := semver.Make(group.Channel.Package.Version)
	if !instanceSemver.LT(packageSemver) {
		return nil, ErrNoUpdatePackageAvailable
	}

	packages, err := api.getPackagesWithFloorsForUpdate(group, instanceVersion)
	if err != nil {
		return nil, err
	}

	// Safety check: verify the next package isn't blacklisted for this channel
	// This should never happen (floors/targets can't be blacklisted for their own channel)
	// but we check anyway for data consistency
	nextPkg := packages[0]
	if slices.Contains(nextPkg.ChannelsBlacklist, group.Channel.ID) {
		l.Error().Str("package", nextPkg.Version).Str("channel", group.Channel.ID).
			Msg("Package is blacklisted for its own channel - data inconsistency!")
		return nil, ErrNoUpdatePackageAvailable
	}

	if err := api.enforceRolloutPolicy(instance, group); err != nil {
		return nil, err
	}

	// Grant the update using the version we're actually returning
	version := packages[0].Version
	if err := api.grantUpdate(instance, version); err != nil {
		l.Error().Err(err).Str("version", version).Str("instance", instance.ID).Msg("GetUpdatePackage - grantUpdate error")
		return nil, ErrUpdateGrantFailed
	}

	// Record activity
	if !api.hasRecentActivity(activityRolloutStarted, ActivityQueryParams{
		Severity: activityInfo,
		AppID:    appID,
		Version:  version,
		GroupID:  groupID,
	}) {
		if err := api.newGroupActivityEntry(activityRolloutStarted, activityInfo, version, appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not add new group activity entry")
		}
	}

	// Set rollout in progress
	if !group.RolloutInProgress {
		if err := api.setGroupRolloutInProgress(groupID, true); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not set rollout progress")
		}
	}

	return packages[0], nil
}

// GetUpdatePackagesForSyncer returns all packages (floors + target) for a syncer client
func (api *API) GetUpdatePackagesForSyncer(instanceID, instanceAlias, instanceIP, instanceVersion, appID, groupID, instanceOEM, instanceAlephVersion string) ([]*Package, error) {
	instance, err := api.RegisterInstance(instanceID, instanceAlias, instanceIP, instanceVersion, appID, groupID, instanceOEM, instanceAlephVersion)
	if err != nil {
		l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not register instance")
		return nil, ErrRegisterInstanceFailed
	}

	if instance.Application.Status.Valid {
		switch int(instance.Application.Status.Int64) {
		case InstanceStatusDownloading, InstanceStatusDownloaded, InstanceStatusInstalled:
			return nil, ErrUpdateInProgressOnInstance
		}
	}

	group, err := api.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	if group.Channel == nil || group.Channel.Package == nil {
		if err := api.newGroupActivityEntry(activityPackageNotFound, activityWarning, "0.0.0", appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not add new group activity entry")
		}
		return nil, ErrNoPackageFound
	}

	// Check if update is needed
	instanceSemver, _ := semver.Make(instanceVersion)
	packageSemver, _ := semver.Make(group.Channel.Package.Version)
	if !instanceSemver.LT(packageSemver) {
		return nil, ErrNoUpdatePackageAvailable
	}

	packages, err := api.getPackagesWithFloorsForUpdate(group, instanceVersion)
	if err != nil {
		return nil, err
	}

	// Safety check: verify no packages are blacklisted for this channel
	// Syncers need all packages, so if any is blacklisted we can't send a valid manifest
	// This should never happen (floors/targets can't be blacklisted for their own channel)
	// but we check anyway for data consistency
	for _, pkg := range packages {
		if slices.Contains(pkg.ChannelsBlacklist, group.Channel.ID) {
			l.Error().Str("package", pkg.Version).Str("channel", group.Channel.ID).
				Msg("Package is blacklisted for its own channel - data inconsistency!")
			return nil, ErrNoUpdatePackageAvailable
		}
	}

	if err := api.enforceRolloutPolicy(instance, group); err != nil {
		return nil, err
	}

	// Grant the update using target version
	targetVersion := packages[len(packages)-1].Version
	if err := api.grantUpdate(instance, targetVersion); err != nil {
		l.Error().Err(err).Str("version", targetVersion).Str("instance", instance.ID).Msg("GetUpdatePackagesForSyncer - grantUpdate error")
		return nil, ErrUpdateGrantFailed
	}

	// Record activity
	if !api.hasRecentActivity(activityRolloutStarted, ActivityQueryParams{
		Severity: activityInfo,
		AppID:    appID,
		Version:  targetVersion,
		GroupID:  groupID,
	}) {
		if err := api.newGroupActivityEntry(activityRolloutStarted, activityInfo, targetVersion, appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not add new group activity entry")
		}
	}

	// Set rollout in progress
	if !group.RolloutInProgress {
		if err := api.setGroupRolloutInProgress(groupID, true); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not set rollout progress")
		}
	}

	return packages, nil
}

// enforceRolloutPolicy validates if an update should be provided to the
// requesting instance based on the group rollout policy and the current status
// of the updates taking place in the group.
func (api *API) enforceRolloutPolicy(instance *Instance, group *Group) error {
	appID := instance.Application.ApplicationID

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
		l.Error().Err(err).Msg("GetUpdatePackage - getGroupUpdatesStats error (propagates as ErrGetUpdatesStatsFailed):")
		return ErrGetUpdatesStatsFailed
	}

	if group.PolicySafeMode && updatesStats.UpdatesToCurrentVersionAttempted == 0 {
		effectiveMaxUpdates = 1
	}

	if updatesStats.UpdatesGrantedInLastPeriod >= effectiveMaxUpdates {
		if err := api.updateInstanceStatus(instance.ID, appID, InstanceStatusOnHold); err != nil {
			l.Error().Err(err).Msg("enforceRolloutPolicy - could not update instance status")
		}
		return ErrMaxUpdatesPerPeriodLimitReached
	}

	if updatesStats.UpdatesInProgress >= effectiveMaxUpdates {
		if err := api.updateInstanceStatus(instance.ID, appID, InstanceStatusOnHold); err != nil {
			l.Error().Err(err).Msg("enforceRolloutPolicy - could not update instance status")
		}
		return ErrMaxConcurrentUpdatesLimitReached
	}

	if group.PolicySafeMode && updatesStats.UpdatesTimedOut >= effectiveMaxUpdates {
		if group.PolicyUpdatesEnabled {
			if err := api.disableUpdates(group.ID); err != nil {
				l.Error().Err(err).Msg("enforceRolloutPolicy - could not disable updates")
			}
		}
		if err := api.updateInstanceStatus(instance.ID, appID, InstanceStatusOnHold); err != nil {
			l.Error().Err(err).Msg("enforceRolloutPolicy - could not update instance status")
		}
		return ErrMaxTimedOutUpdatesLimitReached
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

// getPackagesWithFloorsForUpdate returns floors + target for the given group and instance version
// This is a helper method extracted from the UpdateHandler logic
func (api *API) getPackagesWithFloorsForUpdate(group *Group, instanceVersion string) ([]*Package, error) {
	if group.Channel == nil || group.Channel.Package == nil {
		return nil, ErrNoPackageFound
	}

	// Get required floors using the channel
	requiredFloors, err := api.GetRequiredChannelFloors(
		group.Channel,
		instanceVersion,
	)

	if err != nil {
		return nil, err
	}

	targetPkg := group.Channel.Package
	// Check if target is already included (when target is also a floor)
	if len(requiredFloors) > 0 && requiredFloors[len(requiredFloors)-1].ID == targetPkg.ID {
		return requiredFloors, nil
	}

	// Append target if not already included
	return append(requiredFloors, targetPkg), nil
}
