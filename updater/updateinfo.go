package updater

import (
	"errors"

	"github.com/flatcar/go-omaha/omaha"
)

// UpdateInfo wraps helper functions and fields
// to fetch specific values from the omaha response
// that was recieved for check if any new update
// exists request.
type UpdateInfo struct {
	HasUpdate     bool
	Version       string
	UpdateStatus  string
	AppID         string
	URLs          []string
	Packages      []*omaha.Package
	omahaResponse *omaha.Response
}

// newUpdateInfo returns UpdateInfo from omaha.Response and appID.
func newUpdateInfo(resp *omaha.Response, appID string) (*UpdateInfo, error) {
	if resp == nil {
		return nil, errors.New("invalid omaha response")
	}
	app := resp.GetApp(appID)
	if app == nil {
		return nil, errors.New("invalid omaha response and appID")
	}
	if app.UpdateCheck == nil {
		return nil, errors.New("omaha response is not a valid update check response")
	}

	version := ""
	if app.UpdateCheck.Manifest != nil {
		version = app.UpdateCheck.Manifest.Version
	}

	var packages []*omaha.Package
	if app.UpdateCheck.Manifest != nil && app.UpdateCheck.Manifest.Packages != nil {
		packages = app.UpdateCheck.Manifest.Packages
	}

	var urls []string
	if app.UpdateCheck.URLs != nil {
		for _, url := range app.UpdateCheck.URLs {
			urls = append(urls, url.CodeBase)
		}
	}

	return &UpdateInfo{
		HasUpdate:     app.Status == omaha.AppOK && app.UpdateCheck.Status == "ok",
		Version:       version,
		UpdateStatus:  string(app.UpdateCheck.Status),
		AppID:         appID,
		URLs:          urls,
		Packages:      packages,
		omahaResponse: resp,
	}, nil
}

// URL returns the first update URL in the omaha response,
// returns "" if the URL is not present in the omaha response.
func (u *UpdateInfo) URL() string {
	urls := u.URLs
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

// Package returns the first package from the omaha response,
// returns nil if the package is not present in the omaha response.
func (u *UpdateInfo) Package() *omaha.Package {
	pkgs := u.Packages
	if len(pkgs) == 0 {
		return nil
	}
	return pkgs[0]
}

// OmahaReponse returns the raw omaha response.
func (u *UpdateInfo) OmahaResponse() *omaha.Response {
	return u.omahaResponse
}
