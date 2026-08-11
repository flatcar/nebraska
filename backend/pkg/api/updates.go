package api

import (
	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

var (
	// ErrRegisterInstanceFailed indicates that the instance registration did
	// not succeed.
	ErrRegisterInstanceFailed = types.ErrRegisterInstanceFailed

	// ErrUpdateInProgressOnInstance indicates that an update is currently in
	// progress on the instance requesting an update package, so the request
	// will be rejected.
	ErrUpdateInProgressOnInstance = types.ErrUpdateInProgressOnInstance

	// ErrNoPackageFound indicates that the group doesn't have a channel
	// assigned or that the channel doesn't have a package assigned.
	ErrNoPackageFound = dbreads.ErrNoPackageFound

	// ErrNoUpdatePackageAvailable indicates that the instance requesting the
	// update has already the latest version of the application.
	ErrNoUpdatePackageAvailable = types.ErrNoUpdatePackageAvailable

	// ErrUpdateGrantFailed indicates that the update could not be granted
	// due to a database or internal error.
	ErrUpdateGrantFailed = types.ErrUpdateGrantFailed

	// ErrUpdatesDisabled indicates that updates are not enabled in the group.
	ErrUpdatesDisabled = types.ErrUpdatesDisabled

	// ErrGetUpdatesStatsFailed indicates that there was a problem getting the
	// updates stats of the group which are needed to enforce the rollout
	// policy.
	ErrGetUpdatesStatsFailed = types.ErrGetUpdatesStatsFailed

	// ErrMaxUpdatesPerPeriodLimitReached indicates that the maximum number of
	// updates per period has been reached.
	ErrMaxUpdatesPerPeriodLimitReached = types.ErrMaxUpdatesPerPeriodLimitReached

	// ErrMaxConcurrentUpdatesLimitReached indicates that the maximum number of
	// concurrent updates has been reached.
	ErrMaxConcurrentUpdatesLimitReached = types.ErrMaxConcurrentUpdatesLimitReached

	// ErrMaxTimedOutUpdatesLimitReached indicates that limit of instances that
	// timed out while updating has been reached.
	ErrMaxTimedOutUpdatesLimitReached = types.ErrMaxTimedOutUpdatesLimitReached
)
