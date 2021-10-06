package updater

import (
	"encoding/xml"
	"testing"

	"github.com/kinvolk/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	updateExistsResponse = `<?xml version="1.0" encoding="UTF-8"?>
	<response protocol="3.0" server="nebraska">
	   <daystart elapsed_seconds="0" />
	   <app appid="e96281a6-d1af-4bde-9a0a-97b76e56dc57" status="ok">
		  <updatecheck status="ok">
			 <urls>
				<url codebase="https://kinvolk.io/test/response" />
			 </urls>
			 <manifest version="2191.5.0">
				<packages>
				   <package name="flatcar_production_update.gz" hash="test+x2zIoeClk=" size="465881871" required="true" />
				</packages>
				<actions>
				   <action event="postinstall" sha256="test/FodbjVgqkyF/y8=" DisablePayloadBackoff="true" />
				</actions>
			 </manifest>
		  </updatecheck>
	   </app>
	</response>
	`

	noUpdateResponse = `<?xml version="1.0" encoding="UTF-8"?>
	<response protocol="3.0" server="nebraska">
	   <daystart elapsed_seconds="0" />
	   <app appid="e96281a6-d1af-4bde-9a0a-97b76e56dc57" status="ok">
		  <updatecheck status="noupdate">
			 <urls />
		  </updatecheck>
	   </app>
	</response>`

	errorResponse = `<?xml version="1.0" encoding="UTF-8"?>
	<response protocol="3.0" server="nebraska">
	   <daystart elapsed_seconds="0" />
	   <app appid="h96281a6-d1af-4bde-9a0a-97b76e56dc57" status="error-failedToRetrieveUpdatePackageInfo">
	      <updatecheck status="error-internal">
	         <urls />
	      </updatecheck>
	   </app>
	</response>`

	nonUpdateCheckResponse = `<?xml version="1.0" encoding="UTF-8"?>
	<response protocol="3.0" server="nebraska">
	   <daystart elapsed_seconds="0" />
	   <app appid="e96281a6-d1af-4bde-9a0a-97b76e56dc57" status="error-internal">
	   </app>
	</response>`
	appID      = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"
	errorAppID = "h96281a6-d1af-4bde-9a0a-97b76e56dc57"
)

func TestUpdateInfo(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		appID         string
		isNil         bool
		hasUpdate     bool
		updateStatus  string
		packagesCount int
		urlCount      int
		version       string
	}{
		{
			name:          "update_exists",
			response:      updateExistsResponse,
			appID:         appID,
			isNil:         false,
			hasUpdate:     true,
			updateStatus:  "ok",
			packagesCount: 1,
			urlCount:      1,
			version:       "2191.5.0",
		},
		{
			name:          "no_update_exists",
			response:      noUpdateResponse,
			appID:         appID,
			isNil:         false,
			hasUpdate:     false,
			updateStatus:  "noupdate",
			packagesCount: 0,
			urlCount:      0,
		},
		{
			name:          "error_response",
			response:      errorResponse,
			appID:         errorAppID,
			isNil:         false,
			hasUpdate:     false,
			updateStatus:  "error-internal",
			packagesCount: 0,
			urlCount:      0,
		},
		{
			name:     "non_update_check_response",
			response: nonUpdateCheckResponse,
			appID:    appID,
			isNil:    true,
		},
		{
			name:     "invalid_app_id",
			response: updateExistsResponse,
			appID:    errorAppID,
			isNil:    true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var omahaResponse omaha.Response
			err := xml.Unmarshal([]byte(tc.response), &omahaResponse)
			require.NoError(t, err)
			updateInfo, err := newUpdateInfo(&omahaResponse, tc.appID)
			if tc.isNil {
				assert.Nil(t, updateInfo)
				assert.Error(t, err)
			} else {
				assert.NotNil(t, updateInfo)
				assert.Equal(t, tc.hasUpdate, updateInfo.HasUpdate)
				assert.Equal(t, tc.updateStatus, updateInfo.UpdateStatus)
				assert.Equal(t, tc.version, updateInfo.Version)
				assert.Equal(t, tc.urlCount, len(updateInfo.URLs))
				if tc.urlCount > 0 {
					assert.NotEqual(t, "", updateInfo.URL())
				} else {
					assert.Equal(t, "", updateInfo.URL())
				}
				assert.Equal(t, tc.packagesCount, len(updateInfo.Packages))
				if tc.packagesCount > 0 {
					assert.NotNil(t, updateInfo.Package())
				} else {
					assert.Nil(t, updateInfo.Package())
				}
				assert.Equal(t, &omahaResponse, updateInfo.OmahaResponse())
			}
		})
	}
}
