package types

import (
	"errors"
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
