package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
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

type Updater struct {
	omahaURL      string
	clientVersion string

	instanceID      string
	instanceVersion string
	sessionID       string

	appID   string
	channel string

	httpClient *retryablehttp.Client
}

func New(omahaURL string, instanceID string, instanceVersion string, appID string, channel string) (*Updater, error) {
	return NewWithHttpClient(omahaURL, instanceID, instanceVersion, appID, channel, retryablehttp.NewClient())
}

func NewWithHttpClient(omahaURL string, instanceID string, instanceVersion string, appID string, channel string, httpClient *retryablehttp.Client) (*Updater, error) {
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
		httpClient:      retryablehttp.NewClient(),
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
	requestByte, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := u.httpClient.Post(url, "text/xml", bytes.NewReader(requestByte))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// A response over 1M in size is certainly bogus.
	respBody := &io.LimitedReader{R: resp.Body, N: 1024 * 1024}
	contentType := resp.Header.Get("Content-Type")
	omahaResp, err := omaha.ParseResponse(contentType, respBody)
	if err != nil {
		// TODO: Convert to omaha error response
		return nil, err
	}
	return omahaResp, nil
}

func (u *Updater) CheckForUpdates(ctx context.Context) (*omaha.Response, error) {
	req := NewAppRequest(u)
	app := req.GetApp(u.appID)
	app.AddUpdateCheck()

	resp, err := u.SendOmahaRequest(u.omahaURL, req)
	requestByte, _ := xml.Marshal(resp)
	fmt.Println(string(requestByte))
	if err != nil {
		return nil, err
	}
	appResp := resp.GetApp(u.appID)
	if appResp.Status != omaha.AppOK {
		u.ReportProgress(ctx, ProgressError)
		return nil, errors.New(fmt.Sprintf("No updates avaiable for appID: %s appVersion: %s", u.appID, u.instanceVersion))
	}

	return resp, nil
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

func (u *Updater) TryUpdate(ctx context.Context, handlers Handlers) error {

	fmt.Println("Version before run:", u.instanceVersion)

	// Check for updates
	resp, err := u.CheckForUpdates(ctx)
	if err != nil {
		return err
	}

	// Fetch update
	err = handlers.FetchUpdate(ctx)
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

	err = handlers.ApplyUpdate(ctx)
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

	app := resp.GetApp(u.appID)
	fmt.Println("Version after run:", app.UpdateCheck.Manifest.Version)

	err = u.ReportProgress(ctx, ProgressUpdateComplete)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	omahaURL := "http://localhost:8000/v1/update/"

	appID := "4f101e1e-7fca-4a45-a222-c2aa2fd78d84"

	instanceID := "Test8cd4-a18e-466c-b8e1-431b849bfb4c"
	instanceVersion := "0.2.0"

	emptyHandler := NewEmptyHandler()

	updater, err := New(omahaURL, instanceID, instanceVersion, appID, "stable")

	if err != nil {
		log.Fatal("up err", err)
	}

	ctx := context.TODO()

	err = updater.TryUpdate(ctx, emptyHandler)
	fmt.Println("Err", err)

}
