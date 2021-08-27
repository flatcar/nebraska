package updater

import "github.com/kinvolk/go-omaha/omaha"

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

func (u *UpdateInfo) GetURL() string {
	if urls := u.GetURLs(); urls != nil {
		return urls[0]
	}

	return ""
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
