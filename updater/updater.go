package updater

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/kinvolk/go-omaha/omaha"
)

const defaultClientVersion = "go-omaha"

type Progress int

const (
	ProgressDownloadStarted Progress = iota
	ProgressDownloadFinished
	ProgressInstallationStarted
	ProgressInstallationFinished
	ProgressUpdateComplete
	ProgressUpdateCompleteAndRebooted
	ProgressError
)

var progressEventMap = map[Progress]*omaha.EventRequest{
	ProgressDownloadStarted: {
		Type:   omaha.EventTypeUpdateDownloadStarted,
		Result: omaha.EventResultSuccess,
	},
	ProgressDownloadFinished: {
		Type:   omaha.EventTypeUpdateDownloadFinished,
		Result: omaha.EventResultSuccess,
	},
	ProgressUpdateComplete: {
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccess,
	},
	ProgressUpdateCompleteAndRebooted: {
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccessReboot,
	},
	ProgressInstallationStarted: {
		Type:   omaha.EventTypeInstallStarted,
		Result: omaha.EventResultSuccess,
	},
	ProgressInstallationFinished: {
		Type:   omaha.EventTypeInstallStarted,
		Result: omaha.EventResultSuccess,
	},
	ProgressError: {
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultError,
	},
}

type OmahaRequestHandler interface {
	Handle(req *omaha.Request) (*omaha.Response, error)
}

type Updater struct {
	omahaURL      string
	clientVersion string

	instanceID      string
	instanceVersion string
	sessionID       string

	appID   string
	channel string

	omahaReqHandler OmahaRequestHandler
}

func New(omahaURL string, appID string, channel string, instanceID string, instanceVersion string) (*Updater, error) {
	return NewWithHttpClient(omahaURL, appID, channel, instanceID, instanceVersion, retryablehttp.NewClient())
}

func NewWithHttpClient(omahaURL string, appID string, channel string, instanceID string, instanceVersion string, httpClient *retryablehttp.Client) (*Updater, error) {
	_, err := url.Parse(omahaURL)
	if err != nil {
		return nil, err
	}
	return &Updater{
		omahaURL:        omahaURL,
		clientVersion:   defaultClientVersion,
		instanceID:      instanceID,
		sessionID:       uuid.New().String(),
		appID:           appID,
		instanceVersion: instanceVersion,
		channel:         channel,
		omahaReqHandler: NewHttpOmahaReqHandler(omahaURL),
	}, nil
}

func NewAppRequest(u *Updater) *omaha.Request {
	req := omaha.NewRequest()
	req.Version = u.clientVersion
	req.UserID = u.instanceID
	req.SessionID = u.sessionID

	app := req.AddApp(u.appID, u.instanceVersion)
	app.MachineID = u.instanceID
	app.BootID = u.sessionID
	app.Track = u.channel

	return req
}

func (u *Updater) SendOmahaRequest(url string, req *omaha.Request) (*omaha.Response, error) {
	return u.omahaReqHandler.Handle(req)
}

func (u *Updater) CheckForUpdates(ctx context.Context) (*UpdateInfo, error) {
	req := NewAppRequest(u)
	app := req.GetApp(u.appID)
	app.AddUpdateCheck()

	resp, err := u.SendOmahaRequest(u.omahaURL, req)
	if err != nil {
		return nil, err
	}
	appResp := resp.GetApp(u.appID)
	info := &UpdateInfo{
		HasUpdate: appResp != nil && appResp.Status == omaha.AppOK && appResp.UpdateCheck.Status == "ok",
		omahaResp: resp,
	}

	return info, nil
}

func (u *Updater) ReportProgress(ctx context.Context, progress Progress) error {
	val, ok := progressEventMap[progress]
	if !ok {
		return errors.New("Invalid Progress value")
	}
	resp, err := u.SendOmahaEvent(ctx, val)
	if err != nil {
		return err
	}

	app := resp.GetApp(u.appID)
	if app.Status != "ok" {
		return errors.New(fmt.Sprintf("Error when reporting progress to omaha server, got not ok response"))
	}

	return nil
}

func (u *Updater) SendOmahaEvent(ctx context.Context, event *omaha.EventRequest) (*omaha.Response, error) {

	req := NewAppRequest(u)
	app := req.GetApp(u.appID)
	app.Events = append(app.Events, event)

	return u.SendOmahaRequest(u.omahaURL, req)
}

func (u *Updater) TryUpdate(ctx context.Context, handler UpdateHandler) error {
	fmt.Println("Version before run:", u.instanceVersion)

	// Check for updates
	info, err := u.CheckForUpdates(ctx)
	if err != nil {
		return err
	}

	if !info.HasUpdate {
		return fmt.Errorf("No update available for app %v, channel %v: %v", u.appID, u.channel, info.GetUpdateStatus())
	}

	// Fetch update
	err = handler.FetchUpdate(ctx, info)
	if err != nil {
		err := u.ReportProgress(ctx, ProgressError)
		if err != nil {
			return err
		}
	}

	err = u.ReportProgress(ctx, ProgressDownloadFinished)
	if err != nil {
		return err
	}

	err = handler.ApplyUpdate(ctx, info)
	if err != nil {
		err := u.ReportProgress(ctx, ProgressError)
		if err != nil {
			return err
		}
	}

	err = u.ReportProgress(ctx, ProgressInstallationFinished)
	if err != nil {
		return err
	}

	version := info.GetVersion()
	fmt.Println("Version after run:", version)

	err = u.ReportProgress(ctx, ProgressUpdateComplete)
	if err != nil {
		return err
	}

	return nil
}
