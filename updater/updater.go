package updater

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/google/uuid"
	"github.com/kinvolk/go-omaha/omaha"
)

const defaultClientVersion = "go-omaha"

// NoUpdateError is returned by TryUpdate when no update is available for app.
type NoUpdateError struct {
	AppID        string
	Channel      string
	UpdateStatus string
}

func (e NoUpdateError) Error() string {
	return "no update available for app"
}

type progress int

const (
	ProgressDownloadStarted progress = iota
	ProgressDownloadFinished
	ProgressInstallationStarted
	ProgressInstallationFinished
	ProgressUpdateComplete
	ProgressUpdateCompleteAndRestarted
	ProgressError
)

func progressToEventRequest(p progress) *omaha.EventRequest {
	switch p {
	case ProgressDownloadStarted:
		return &omaha.EventRequest{
			Type:   omaha.EventTypeUpdateDownloadStarted,
			Result: omaha.EventResultSuccess,
		}
	case ProgressDownloadFinished:
		return &omaha.EventRequest{
			Type:   omaha.EventTypeUpdateDownloadFinished,
			Result: omaha.EventResultSuccess,
		}
	case ProgressUpdateComplete:
		return &omaha.EventRequest{
			Type:   omaha.EventTypeUpdateComplete,
			Result: omaha.EventResultSuccess,
		}
	case ProgressUpdateCompleteAndRestarted:
		return &omaha.EventRequest{
			Type:   omaha.EventTypeUpdateComplete,
			Result: omaha.EventResultSuccessReboot,
		}
	case ProgressInstallationStarted:
		return &omaha.EventRequest{
			Type:   omaha.EventTypeInstallStarted,
			Result: omaha.EventResultSuccess,
		}
	case ProgressInstallationFinished:
		return &omaha.EventRequest{
			Type:   omaha.EventTypeInstallStarted,
			Result: omaha.EventResultSuccess,
		}
	case ProgressError:
		return &omaha.EventRequest{
			Type:   omaha.EventTypeUpdateComplete,
			Result: omaha.EventResultError,
		}
	default:
		return nil
	}
}

// OmahaRequestHandler wraps the Handle function which
// takes in context, url, omaha.Request and
// returns omaha.Response.
type OmahaRequestHandler interface {
	Handle(ctx context.Context, url string, req *omaha.Request) (*omaha.Response, error)
}

// Updater interface wraps functions required to update an application
type Updater interface {
	// SendOmahaRequest sends raw Omaha request provided to the Omaha server.
	SendOmahaRequest(ctx context.Context, req *omaha.Request) (*omaha.Response, error)
	// SendOmahaEvent sends Omaha event request to the Omaha server.
	SendOmahaEvent(ctx context.Context, event *omaha.EventRequest) (*omaha.Response, error)
	// CheckForUpdates checks if there are new updated versions for the application.
	CheckForUpdates(ctx context.Context) (*UpdateInfo, error)
	// ReportProgress reports progress of update to the Omaha server.
	ReportProgress(ctx context.Context, progress progress) error
	// ReportError reports errors with custom error code.
	ReportError(ctx context.Context, errorCode *int) error
	// CompleteUpdate takes the UpdateInfo, updates the application
	// version with the UpdateInfo version and reports ProgressUpdateComplete to
	// the Omaha server.
	CompleteUpdate(ctx context.Context, info *UpdateInfo) error
	// TryUpdate function takes an implementation of UpdateHandler
	// and runs the complete flow from checking updates to reporting update status.
	TryUpdate(ctx context.Context, handler UpdateHandler) error
	// InstanceVersion returns the current version of the application.
	InstanceVersion() string
	// SetInstanceVersion takes a string and sets it as the application version.
	SetInstanceVersion(version string)
}

// updater implements the Updater interface.
type updater struct {
	omahaURL      string
	clientVersion string

	instanceID      string
	instanceVersion string
	sessionID       string

	appID   string
	channel string

	debug           bool
	omahaReqHandler OmahaRequestHandler

	mu sync.RWMutex
}

// Config is used to configure new updater instance.
type Config struct {
	OmahaURL        string
	AppID           string
	Channel         string
	InstanceID      string
	InstanceVersion string
	Debug           bool
	OmahaReqHandler OmahaRequestHandler
}

// New takes config and returns Updater and error,
// returns an error if OmahaURL in the config is invalid.
func New(config Config) (Updater, error) {
	if _, err := url.Parse(config.OmahaURL); err != nil {
		return nil, fmt.Errorf("parsing URL %q: %w", config.OmahaURL, err)
	}

	updater := updater{
		omahaURL:        config.OmahaURL,
		clientVersion:   defaultClientVersion,
		instanceID:      config.InstanceID,
		sessionID:       uuid.New().String(),
		appID:           config.AppID,
		instanceVersion: config.InstanceVersion,
		channel:         config.Channel,
		debug:           config.Debug,
		omahaReqHandler: config.OmahaReqHandler,
	}
	if config.OmahaReqHandler == nil {
		updater.omahaReqHandler = NewOmahaRequestHandler(nil)
	}

	return &updater, nil
}

// newAppRequest create an Omaha request containing
// the application configured in the updater.
func (u *updater) newAppRequest() *omaha.Request {
	req := omaha.NewRequest()
	req.Version = u.clientVersion
	req.UserID = u.instanceID
	req.SessionID = u.sessionID

	app := req.AddApp(u.appID, u.InstanceVersion())
	app.MachineID = u.instanceID
	app.BootID = u.sessionID
	app.Track = u.channel

	return req
}

// SendOmahaRequest uses the OmahaReqHandler of the updater to send
// request to the Omaha server. If updater is configured with debug
// value as true the raw request and response is printed.
func (u *updater) SendOmahaRequest(ctx context.Context, req *omaha.Request) (*omaha.Response, error) {
	if u.debug {
		requestByte, err := xml.Marshal(req)
		if err == nil {
			fmt.Println("Raw Request:\n", string(requestByte))
		}
	}
	resp, err := u.omahaReqHandler.Handle(ctx, u.omahaURL, req)
	if u.debug {
		responseByte, err := xml.Marshal(resp)
		if err == nil {
			fmt.Println("Raw Response:\n", string(responseByte))
		}
	}
	return resp, err
}

// CheckForUpdates sends a request checking if the application has any new updates
// to the Omaha server.
func (u *updater) CheckForUpdates(ctx context.Context) (*UpdateInfo, error) {
	req := u.newAppRequest()
	app := req.GetApp(u.appID)
	app.AddUpdateCheck()

	resp, err := u.SendOmahaRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("sending update check omaha request: %w", err)
	}

	return newUpdateInfo(resp, u.appID)
}

// ReportProgress takes the progress value and converts it
// to corresponding Omaha event request to report current
// progress of the application update to Omaha server.
func (u *updater) ReportProgress(ctx context.Context, progress progress) error {
	eventRequest := progressToEventRequest(progress)
	if eventRequest == nil {
		return errors.New("invalid Progress value")
	}
	err := u.processEventRequest(ctx, eventRequest)
	if err != nil {
		return fmt.Errorf("reporting progress: %w", err)
	}
	return nil
}

// ReportError takes an optional errorCode and reports
// that an error occured during the installation process.
// The optional errorCode can be used to send custom
// error codes to the server. This error code can then be
// used to trace out errors custom to the application
// installation process.
func (u *updater) ReportError(ctx context.Context, errorCode *int) error {
	errorEvent := progressToEventRequest(ProgressError)
	if errorCode != nil {
		errorEvent.ErrorCode = *errorCode
	}

	err := u.processEventRequest(ctx, errorEvent)
	if err != nil {
		return fmt.Errorf("reporting error: %w", err)
	}
	return nil
}

// CompleteUpdate sends an ProgressUpdateComplete event to the Omaha server
// and sets the version in the UpdateInfo as the current version of
// the instance in the updater.
func (u *updater) CompleteUpdate(ctx context.Context, info *UpdateInfo) error {
	if info == nil {
		return errors.New("invalid UpdateInfo")
	}

	version := info.Version
	if version == "" {
		return fmt.Errorf("invalid version, can't report complete event to omaha server")
	}

	err := u.ReportProgress(ctx, ProgressUpdateComplete)
	if err != nil {
		return fmt.Errorf("reporting ProgressUpdateComplete to omaha server: %w", err)
	}

	u.SetInstanceVersion(version)
	return nil
}

// SendOmahaEvent sends the event request to the Omaha server
// and returns the Omaha.Response returned by the Omaha server.
func (u *updater) SendOmahaEvent(ctx context.Context, event *omaha.EventRequest) (*omaha.Response, error) {
	req := u.newAppRequest()
	app := req.GetApp(u.appID)
	app.Events = append(app.Events, event)

	resp, err := u.SendOmahaRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("sending omaha event request: %w", err)
	}
	return resp, nil
}

// InstanceVersion returns the current version of the application.
func (u *updater) InstanceVersion() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.instanceVersion
}

// SetInstanceVersion sets the current instance version
// of the application to the updater.
func (u *updater) SetInstanceVersion(version string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.instanceVersion = version
}

// TryUpdate function takes in an UpdateHandler and performs
// the complete flow from checking for updates to reporting
// status etc and returns an error if anything fails in the flow.
func (u *updater) TryUpdate(ctx context.Context, handler UpdateHandler) error {
	if handler == nil {
		return errors.New("invalid UpdateHandler")
	}

	info, err := u.CheckForUpdates(ctx)
	if err != nil {
		return fmt.Errorf("checking for update: %w", err)
	}

	if !info.HasUpdate {
		return NoUpdateError{
			AppID:        u.appID,
			Channel:      u.channel,
			UpdateStatus: info.UpdateStatus,
		}
	}

	if err := handler.FetchUpdate(ctx, *info); err != nil {
		if reportErr := u.ReportError(ctx, nil); reportErr != nil && u.debug {
			fmt.Println("Reporting error to omaha server:", errors.Unwrap(reportErr))
		}
		return fmt.Errorf("fetching update: %w", err)
	}

	if err := u.ReportProgress(ctx, ProgressDownloadFinished); err != nil {
		return fmt.Errorf("reporting progress download finished: %w", err)
	}

	if err := handler.ApplyUpdate(ctx, *info); err != nil {
		if reportErr := u.ReportError(ctx, nil); reportErr != nil && u.debug {
			fmt.Println("Reporting error to omaha server:", errors.Unwrap(reportErr))
		}
		return fmt.Errorf("applying update: %w", err)
	}

	if err := u.ReportProgress(ctx, ProgressInstallationFinished); err != nil {
		return fmt.Errorf("reporting progress install finished: %w", err)
	}

	return u.CompleteUpdate(ctx, info)
}

func (u *updater) processEventRequest(ctx context.Context, eventReq *omaha.EventRequest) error {
	resp, err := u.SendOmahaEvent(ctx, eventReq)
	if err != nil {
		return fmt.Errorf("sending progress event to omaha server: %w", err)
	}

	app := resp.GetApp(u.appID)
	if app.Status != "ok" {
		return fmt.Errorf("reporting progress to omaha server, got status %q", app.Status)
	}
	return nil
}
