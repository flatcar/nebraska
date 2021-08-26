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
	"os"

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

type UpdateHandler interface {
	FetchUpdate(info *UpdateInfo) error
	ApplyUpdate(info *UpdateInfo) error
}

type UpdateInfo struct {
	HasUpdate bool
	omahaResp *omaha.Response
}

func (u *UpdateInfo) GetVersion() string {
	app := u.getApp()
	if app != nil && app.UpdateCheck != nil {
		return app.UpdateCheck.Manifest.Version
	}

	return ""
}

func (u *UpdateInfo) GetURLs() []string {
	app := u.getApp()
	if app != nil && app.UpdateCheck != nil {
		omahaURLs := app.UpdateCheck.URLs
		urls := make([]string, len(omahaURLs))
		for i, url := range omahaURLs {
			urls[i] = url.CodeBase
		}

		return urls
	}

	return nil
}

func (u *UpdateInfo) GetUpdateStatus() string {
	app := u.getApp()
	if app != nil && app.UpdateCheck != nil {
		return string(app.UpdateCheck.Status)
	}

	return ""
}

func (u *UpdateInfo) GetOmahaResponse() *omaha.Response {
	return u.omahaResp
}

func (u *UpdateInfo) getApp() *omaha.AppResponse {
	if u.omahaResp != nil {
		return u.omahaResp.Apps[0]
	}

	return nil
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

func (u *Updater) CheckForUpdates(ctx context.Context) (*UpdateInfo, error) {
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
	err = handler.FetchUpdate(ctx)
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

	err = handler.ApplyUpdate(ctx)
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
