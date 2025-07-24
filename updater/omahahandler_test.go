package updater_test

import (
	"context"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/updater"
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
	serverURL := "127.0.0.1:0"
	s, err := omaha.NewTrivialServer(serverURL)

	require.NoError(t, err)
	require.NotNil(t, s)
	//nolint:errcheck
	go s.Serve()

	t.Cleanup(func() {
		err := s.Destroy()
		t.Log("Error destroying trivial omaha server:", err)
	})

	var omahaRequest omaha.Request

	err = xml.Unmarshal([]byte(sampleRequest), &omahaRequest)
	require.NoError(t, err)

	tests := []struct {
		name    string
		url     string
		request *omaha.Request
		isErr   bool
	}{
		{
			name:    "valid_request",
			url:     fmt.Sprintf("http://%s/v1/update", s.Addr().String()),
			request: &omahaRequest,
			isErr:   false,
		},
		{
			name:    "invalid_request",
			url:     fmt.Sprintf("http://%s/v1/update", s.Addr().String()),
			request: nil,
			isErr:   true,
		},
		{
			name:    "invalid_server_url",
			url:     "http:/127.0.0.67/v1/update",
			request: &omahaRequest,
			isErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handler := updater.NewOmahaRequestHandler(nil)
			resp, err := handler.Handle(context.TODO(), tc.url, tc.request)
			if tc.isErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
