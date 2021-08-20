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
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/kinvolk/go-omaha/omaha"
)

const defaultClientVersion = "go-omaha"

var (
	// These events are what update_engine sends to CoreUpdate to
	// mark different steps in the update process.
	EventDownloading = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateDownloadStarted,
		Result: omaha.EventResultSuccess,
	}
	EventDownloaded = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateDownloadFinished,
		Result: omaha.EventResultSuccess,
	}
	EventInstalled = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccess,
	}
	EventComplete = &omaha.EventRequest{
		Type:   omaha.EventTypeUpdateComplete,
		Result: omaha.EventResultSuccessReboot,
	}
)

type Updater interface {
	Ping(ctx context.Context) (*omaha.Response, error)
	UpdateCheck(ctx context.Context) (*omaha.Response, error)
	SendEvent(ctx context.Context, event *omaha.EventRequest) (*omaha.Response, error)
	Run(ctx context.Context) error
}

type updater struct {
	omahaURL      string
	clientVersion string
	isMachine     bool

	userID    string
	sessionID string

	appID      string
	appVersion string

	handlers Handlers
	lock     sync.RWMutex
}

func New(omahaURL string, userID string, appID string, appVersion string, handlers Handlers) (Updater, error) {

	_, err := url.Parse(omahaURL)
	if err != nil {
		return nil, err
	}
	return &updater{
		omahaURL:      omahaURL,
		clientVersion: defaultClientVersion,
		userID:        userID,
		sessionID:     uuid.New().String(),
		appID:         appID,
		appVersion:    appVersion,
		handlers:      handlers,
	}, nil
}

func NewAppRequest(u *updater) *omaha.Request {
	req := omaha.NewRequest()
	req.Version = u.clientVersion
	req.UserID = u.userID
	req.SessionID = u.sessionID
	if u.isMachine {
		req.IsMachine = 1
	}

	app := req.AddApp(u.appID, u.appVersion)
	app.MachineID = u.userID
	app.BootID = u.sessionID
	app.Track = "stable"
	return req
}

func OmahaRequest(url string, req *omaha.Request) (*omaha.Response, error) {
	requestByte, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10

	resp, err := retryClient.Post(url, "text/xml", bytes.NewReader(requestByte))
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

func (u *updater) Ping(ctx context.Context) (*omaha.Response, error) {

	req := NewAppRequest(u)
	app := req.GetApp(u.appID)
	app.AddPing()

	return OmahaRequest(u.omahaURL, req)
}

func (u *updater) UpdateCheck(ctx context.Context) (*omaha.Response, error) {

	req := NewAppRequest(u)
	app := req.GetApp(u.appID)
	app.AddUpdateCheck()

	return OmahaRequest(u.omahaURL, req)
}

func (u *updater) SendEvent(ctx context.Context, event *omaha.EventRequest) (*omaha.Response, error) {

	req := NewAppRequest(u)
	app := req.GetApp(u.appID)
	app.Events = append(app.Events, event)

	return OmahaRequest(u.omahaURL, req)
}

func (u *updater) Run(ctx context.Context) error {
	fmt.Println("Version before run:", u.appVersion)
	omahaResp, err := u.UpdateCheck(ctx)
	if err != nil {
		return err
	}

	app := omahaResp.GetApp(u.appID)
	if app.UpdateCheck.Status != "ok" {
		return errors.New("No new updates available for version")
	}

	// get update handler
	err = u.handlers.GetUpdate(ctx)
	if err != nil {
		// TODO: Send error event

		// u.SendEvent(ctx,)
		return err
	}

	omahaEventResp, err := u.SendEvent(ctx, EventDownloaded)
	if err != nil {
		return err
	}

	eventApp := omahaEventResp.GetApp(u.appID)
	if eventApp.Events[0].Status != "ok" {
		return errors.New("Error setting app status to downloaded")
	}

	// apply update handler
	err = u.handlers.ApplyUpdate(ctx)
	if err != nil {
		// TODO: Send error event

		// u.SendEvent(ctx,)
		return err
	}

	// Update current app version
	u.lock.Lock()
	u.appVersion = app.UpdateCheck.Manifest.Version
	u.lock.Unlock()

	// Send success event
	omahaEventResp, err = u.SendEvent(ctx, EventComplete)
	if err != nil {
		return err
	}

	app = omahaEventResp.GetApp(u.appID)
	if app.Events[0].Status != "ok" {
		return errors.New("Error setting app status to success")
	}

	// Ping to update version to server
	pingResp, err := u.Ping(ctx)
	if err != nil {
		return err
	}

	app = pingResp.GetApp(u.appID)

	if app.Ping.Status != "ok" {
		return errors.New("Error pinging")
	}
	fmt.Println("Version after run:", u.appVersion)
	return nil
}

func main() {

	omahaURL := "http://localhost:8000/v1/update/"

	appID := "e96281a6-d1af-4bde-9a0a-97b76e56dc57"
	appVersion := "2191.3.0"

	emptyHandler := NewEmptyHandler()

	updater, err := New(omahaURL, "Restb611-adbf-449d-9583-08d1f385b5d5", appID, appVersion, emptyHandler)

	if err != nil {
		log.Fatal("up err", err)
	}

	// updater.SendEvent(context.TODO(), EventComplete)
	// err = updater.Run(context.TODO())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	s := gocron.NewScheduler(time.UTC)
	s.Every(15).Seconds().Do(func() {
		fmt.Println("Running")
		err := updater.Run(context.TODO())
		if err != nil {
			fmt.Println("Error", err)
		}
	})
	s.StartBlocking()

}
