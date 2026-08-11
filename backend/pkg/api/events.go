package api

import "github.com/flatcar/nebraska/backend/pkg/api/internal/types"

const (
	// EventUpdateComplete indicates that the update process completed. It could
	// mean a successful or failed updated, depending on the result attached to
	// the event. This applies to all events.
	EventUpdateComplete = types.EventUpdateComplete

	// EventUpdateDownloadStarted indicates that the instance started
	// downloading the update package.
	EventUpdateDownloadStarted = types.EventUpdateDownloadStarted

	// EventUpdateDownloadFinished indicates that the update package was
	// downloaded.
	EventUpdateDownloadFinished = types.EventUpdateDownloadFinished

	// EventUpdateInstalled indicates that the update package was installed.
	EventUpdateInstalled = types.EventUpdateInstalled
)

const (
	// ResultFailed indicates that the operation associated with the event
	// posted failed.
	ResultFailed = types.ResultFailed

	// ResultSuccess indicates that the operation associated with the event
	// posted succeeded.
	ResultSuccess = types.ResultSuccess

	// ResultSuccessReboot also indicates a successful operation, but it's
	// meant only to be used along with events of EventUpdateComplete type.
	// It's important that instances use EventUpdateComplete events in
	// combination with ResultSuccessReboot to communicate a successful update
	// completed as it has a special meaning for Nebraska in order to adjust
	// properly the rollout policies and create activity entries.
	ResultSuccessReboot = types.ResultSuccessReboot
)

var (
	// ErrInvalidInstance indicates that the instance provided is not valid or
	// it doesn't exist.
	ErrInvalidInstance = types.ErrInvalidInstance

	// ErrInvalidApplicationOrGroup indicates that the application or group id
	// provided are not valid or related to each other.
	ErrInvalidApplicationOrGroup = types.ErrInvalidApplicationOrGroup

	// ErrInvalidEventTypeOrResult indicates that the event or result provided
	// are not valid (Nebraska only implements a subset of the Omaha protocol
	// events).
	ErrInvalidEventTypeOrResult = types.ErrInvalidEventTypeOrResult

	// ErrEventRegistrationFailed indicates that the event registration into
	// Nebraska failed.
	ErrEventRegistrationFailed = types.ErrEventRegistrationFailed

	// ErrNoUpdateInProgress indicates that an event was received but there
	// wasn't an update in progress for the provided instance/application, so
	// it was rejected.
	ErrNoUpdateInProgress = types.ErrNoUpdateInProgress

	// ErrFlatcarEventIgnored indicates that a Flatcar updater event was ignored.
	// This is a temporary solution to handle Flatcar specific behaviour.
	ErrFlatcarEventIgnored = types.ErrFlatcarEventIgnored
)
