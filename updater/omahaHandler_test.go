package updater_test

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kinvolk/go-omaha/omaha"
	"github.com/kinvolk/nebraska/updater"
	"github.com/stretchr/testify/require"
)

const (
	sampleRequest = `<?xml version="1.0" encoding="UTF-8"?>
<request protocol="3.0" version="ChromeOSUpdateEngine-0.1.0.0" updaterversion="ChromeOSUpdateEngine-0.1.0.0" installsource="ondemandupdate" ismachine="1">
<os version="Indy" platform="Chrome OS" sp="ForcedUpdate_x86_64"></os>
<app appid="{87efface-864d-49a5-9bb3-4b050a7c227a}" bootid="{7D52A1CC-7066-40F0-91C7-7CB6A871BFDE}" machineid="{8BDE4C4D-9083-4D61-B41C-3253212C0C37}" oem="ec3000" version="ForcedUpdate" track="dev-channel" from_track="developer-build" lang="en-US" board="amd64-generic" hardware_class="" delta_okay="false" >
<ping active="1" a="-1" r="-1"></ping>
<updatecheck targetversionprefix=""></updatecheck>
<event eventtype="3" eventresult="2" previousversion=""></event>
</app>
</request>
`
)

func TestHttpOmahaHandler(t *testing.T) {
	t.Parallel()

	serverURL := "127.0.0.1:0"
	s, err := omaha.NewTrivialServer(serverURL)

	require.NoError(t, err)
	require.NotNil(t, s)
	go s.Serve()

	t.Cleanup(func() {
		s.Destroy()
	})

	var omahaRequest omaha.Request

	err = xml.Unmarshal([]byte(sampleRequest), &omahaRequest)

	validHandler := updater.NewDefaultOmahaRequestHandler(fmt.Sprintf("http://%s/v1/update", s.Addr().String()))

	// Default retryablehttp Client has retry max count of 4 which makes
	// the test case too long to complete, so we use a custom Client
	// with retry max count of 0.
	client := retryablehttp.NewClient()
	client.RetryMax = 0
	invalidHandler := updater.NewHttpOmahaRequestHandler("http:/127.0.0.67/v1/update", client)

	type test struct {
		name    string
		handler updater.OmahaRequestHandler
		request *omaha.Request
		isErr   bool
	}

	tests := []test{
		{
			name:    "valid request",
			handler: validHandler,
			request: &omahaRequest,
			isErr:   false,
		},
		{
			name:    "invalid request",
			handler: validHandler,
			request: nil,
			isErr:   true,
		},
		{
			name:    "invalid server url",
			handler: invalidHandler,
			request: &omahaRequest,
			isErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.handler.Handle(tc.request)
			if tc.isErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}
