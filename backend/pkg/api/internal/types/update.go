package types

import "errors"

var (
	// ErrRegisterInstanceFailed indicates that the instance registration did
	// not succeed.
	ErrRegisterInstanceFailed = errors.New("nebraska: register instance failed")

	// ErrUpdateInProgressOnInstance indicates that an update is currently in
	// progress on the instance requesting an update package, so the request
	// will be rejected.
	ErrUpdateInProgressOnInstance = errors.New("nebraska: update in progress on instance")

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
)
