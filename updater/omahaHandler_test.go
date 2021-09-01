package updater

import (
	"encoding/xml"
	"testing"

	"github.com/kinvolk/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
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
	serverURL := "127.0.0.1:6000"
	s, err := omaha.NewTrivialServer(serverURL)

	assert.NoError(t, err)
	assert.NotNil(t, s)
	defer s.Destroy()
	go s.Serve()

	omahaRequestHandler := NewHttpOmahaReqHandler("http://127.0.0.1:6000/v1/update")

	var omahaRequest omaha.Request

	err = xml.Unmarshal([]byte(sampleRequest), &omahaRequest)
	assert.NoError(t, err)
	resp, err := omahaRequestHandler.Handle(&omahaRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
