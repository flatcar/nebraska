package runtime

import (
	"slices"
	"time"

	"github.com/blang/semver/v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

const (
	maxParallelUpdates = 900000
)

// GetUpdatePackage returns an update package for the instance/application
// provided. The instance details and the application it's running will be
// registered in Nebraska (or updated if it's already registered).
func (s *Service) GetUpdatePackage(inst types.Instance, instApp types.InstanceApplication) (*types.Package, error) {
	instance, err := s.RegisterInstance(inst, instApp)
	if err != nil {
		l.Error().Err(err).Msg("GetUpdatePackage - could not register instance")
		return nil, types.ErrRegisterInstanceFailed
	}

	instanceVersion := instApp.Version
	appID := instApp.ApplicationID
	groupID := instApp.GroupID.String

	if instance.Application.Status.Valid {
		switch int(instance.Application.Status.Int64) {
		case types.InstanceStatusDownloading, types.InstanceStatusDownloaded, types.InstanceStatusInstalled:
			return nil, types.ErrUpdateInProgressOnInstance
		}
	}

	group, err := s.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	if group.Channel == nil || group.Channel.Package == nil {
		if err := s.newGroupActivityEntry(types.ActivityPackageNotFound, types.ActivityWarning, "0.0.0", appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not add new group activity entry")
		}
		return nil, dbreads.ErrNoPackageFound
	}

	// Handle already-granted updates
	if instance.Application.Status.Valid && int(instance.Application.Status.Int64) == types.InstanceStatusUpdateGranted {
		// Check if target package is blacklisted
		if slices.Contains(group.Channel.Package.ChannelsBlacklist, group.Channel.ID) {
			if err := s.updateInstanceObjStatus(instance, types.InstanceStatusComplete); err != nil {
				l.Error().Err(err).Msg("GetUpdatePackage - could not update instance status")
			}
			return nil, types.ErrNoUpdatePackageAvailable
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
				if err := s.updateInstanceObjStatus(instance, types.InstanceStatusComplete); err != nil {
					l.Error().Err(err).Msg("GetUpdatePackage - could not update instance status")
				}
				return nil, types.ErrNoUpdatePackageAvailable
			}

			// Instance hasn't reached granted version yet - return what's next to install
			// This will be the first floor/target above current instance version
			packages, err := s.getPackagesWithFloorsForUpdate(group, instanceVersion)
			if err != nil {
				return nil, err
			}
			// packages[0] should be the granted version since instance < granted
			return packages[0], nil
		}

		// No granted version tracked (old instances) - safer fallback
		// Complete the update to force a fresh grant cycle with all checks
		l.Warn().Str("instanceID", instance.ID).Msg("Already-granted without LastUpdateVersion - completing to force fresh grant")
		if err := s.updateInstanceObjStatus(instance, types.InstanceStatusComplete); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not update instance status")
		}
		return nil, types.ErrNoUpdatePackageAvailable
		// Instance will call again without already-granted status and go through proper checks
	}

	// Check if update is needed
	instanceSemver, _ := semver.Make(instanceVersion)
	packageSemver, _ := semver.Make(group.Channel.Package.Version)
	if !instanceSemver.LT(packageSemver) {
		return nil, types.ErrNoUpdatePackageAvailable
	}

	packages, err := s.getPackagesWithFloorsForUpdate(group, instanceVersion)
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
		return nil, types.ErrNoUpdatePackageAvailable
	}

	if err := s.enforceRolloutPolicy(instance, group); err != nil {
		return nil, err
	}

	// Grant the update using the version we're actually returning
	version := packages[0].Version
	if err := s.grantUpdate(instance, version); err != nil {
		l.Error().Err(err).Str("version", version).Str("instance", instance.ID).Msg("GetUpdatePackage - grantUpdate error")
		return nil, types.ErrUpdateGrantFailed
	}

	// Record activity
	if !s.HasRecentRuntimeActivity(types.ActivityRolloutStarted, types.ActivityQueryParams{
		Severity: types.ActivityInfo,
		AppID:    appID,
		Version:  version,
		GroupID:  groupID,
	}) {
		if err := s.newGroupActivityEntry(types.ActivityRolloutStarted, types.ActivityInfo, version, appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not add new group activity entry")
		}
	}

	// Set rollout in progress
	if !group.RolloutInProgress {
		if err := s.setGroupRolloutInProgress(groupID, true); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackage - could not set rollout progress")
		}
	}

	return packages[0], nil
}

// GetUpdatePackagesForSyncer returns all packages (floors + target) for a syncer client
func (s *Service) GetUpdatePackagesForSyncer(inst types.Instance, instApp types.InstanceApplication) ([]*types.Package, error) {
	instance, err := s.RegisterInstance(inst, instApp)
	if err != nil {
		l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not register instance")
		return nil, types.ErrRegisterInstanceFailed
	}

	instanceVersion := instApp.Version
	appID := instApp.ApplicationID
	groupID := instApp.GroupID.String

	if instance.Application.Status.Valid {
		switch int(instance.Application.Status.Int64) {
		case types.InstanceStatusDownloading, types.InstanceStatusDownloaded, types.InstanceStatusInstalled:
			return nil, types.ErrUpdateInProgressOnInstance
		}
	}

	group, err := s.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	if group.Channel == nil || group.Channel.Package == nil {
		if err := s.newGroupActivityEntry(types.ActivityPackageNotFound, types.ActivityWarning, "0.0.0", appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not add new group activity entry")
		}
		return nil, dbreads.ErrNoPackageFound
	}

	// Check if update is needed
	instanceSemver, _ := semver.Make(instanceVersion)
	packageSemver, _ := semver.Make(group.Channel.Package.Version)
	if !instanceSemver.LT(packageSemver) {
		return nil, types.ErrNoUpdatePackageAvailable
	}

	packages, err := s.getPackagesWithFloorsForUpdate(group, instanceVersion)
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
			return nil, types.ErrNoUpdatePackageAvailable
		}
	}

	if err := s.enforceRolloutPolicy(instance, group); err != nil {
		return nil, err
	}

	// Grant the update using target version
	targetVersion := packages[len(packages)-1].Version
	if err := s.grantUpdate(instance, targetVersion); err != nil {
		l.Error().Err(err).Str("version", targetVersion).Str("instance", instance.ID).Msg("GetUpdatePackagesForSyncer - grantUpdate error")
		return nil, types.ErrUpdateGrantFailed
	}

	// Record activity
	if !s.HasRecentRuntimeActivity(types.ActivityRolloutStarted, types.ActivityQueryParams{
		Severity: types.ActivityInfo,
		AppID:    appID,
		Version:  targetVersion,
		GroupID:  groupID,
	}) {
		if err := s.newGroupActivityEntry(types.ActivityRolloutStarted, types.ActivityInfo, targetVersion, appID, groupID); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not add new group activity entry")
		}
	}

	// Set rollout in progress
	if !group.RolloutInProgress {
		if err := s.setGroupRolloutInProgress(groupID, true); err != nil {
			l.Error().Err(err).Msg("GetUpdatePackagesForSyncer - could not set rollout progress")
		}
	}

	return packages, nil
}

// enforceRolloutPolicy validates if an update should be provided to the
// requesting instance based on the group rollout policy and the current status
// of the updates taking place in the group.
func (s *Service) enforceRolloutPolicy(instance *types.Instance, group *types.Group) error {
	appID := instance.Application.ApplicationID

	if !group.PolicyUpdatesEnabled {
		return types.ErrUpdatesDisabled
	}

	if group.PolicyOfficeHours && !inOfficeHoursNow(group.PolicyTimezone.String) {
		return types.ErrUpdatesDisabled
	}

	effectiveMaxUpdates := group.PolicyMaxUpdatesPerPeriod

	// If no policy enforcement is needed, then we skip getting the update stats below.
	if effectiveMaxUpdates >= maxParallelUpdates && !group.PolicySafeMode {
		return nil
	}

	updatesStats, err := s.GetGroupUpdatesStats(group)
	if err != nil {
		l.Error().Err(err).Msg("GetUpdatePackage - getGroupUpdatesStats error (propagates as ErrGetUpdatesStatsFailed):")
		return types.ErrGetUpdatesStatsFailed
	}

	if group.PolicySafeMode && updatesStats.UpdatesToCurrentVersionAttempted == 0 {
		effectiveMaxUpdates = 1
	}

	if updatesStats.UpdatesGrantedInLastPeriod >= effectiveMaxUpdates {
		if err := s.updateInstanceStatus(instance.ID, appID, types.InstanceStatusOnHold); err != nil {
			l.Error().Err(err).Msg("enforceRolloutPolicy - could not update instance status")
		}
		return types.ErrMaxUpdatesPerPeriodLimitReached
	}

	if updatesStats.UpdatesInProgress >= effectiveMaxUpdates {
		if err := s.updateInstanceStatus(instance.ID, appID, types.InstanceStatusOnHold); err != nil {
			l.Error().Err(err).Msg("enforceRolloutPolicy - could not update instance status")
		}
		return types.ErrMaxConcurrentUpdatesLimitReached
	}

	if group.PolicySafeMode && updatesStats.UpdatesTimedOut >= effectiveMaxUpdates {
		if group.PolicyUpdatesEnabled {
			if err := s.disableUpdates(group.ID); err != nil {
				l.Error().Err(err).Msg("enforceRolloutPolicy - could not disable updates")
			}
		}
		if err := s.updateInstanceStatus(instance.ID, appID, types.InstanceStatusOnHold); err != nil {
			l.Error().Err(err).Msg("enforceRolloutPolicy - could not update instance status")
		}
		return types.ErrMaxTimedOutUpdatesLimitReached
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
func (s *Service) getPackagesWithFloorsForUpdate(group *types.Group, instanceVersion string) ([]*types.Package, error) {
	if group.Channel == nil || group.Channel.Package == nil {
		return nil, dbreads.ErrNoPackageFound
	}

	// Get required floors using the channel
	requiredFloors, err := s.GetRequiredChannelFloors(
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
