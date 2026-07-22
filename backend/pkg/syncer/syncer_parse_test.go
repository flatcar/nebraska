package syncer

import (
	"encoding/xml"
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseOmahaUpdateResponse covers the guard that previously let a malformed
// upstream response panic the whole syncer process: a response with no <app>
// must yield an error, and an <app> without <updatecheck> must yield a nil
// update check (not a panic).
func TestParseOmahaUpdateResponse(t *testing.T) {
	const appID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"

	t.Run("app with updatecheck returns it", func(t *testing.T) {
		resp := omaha.NewResponse()
		resp.AddApp(appID, omaha.AppOK).AddUpdateCheck(omaha.UpdateOK)
		body, err := xml.Marshal(resp)
		require.NoError(t, err)

		uc, err := parseOmahaUpdateResponse(body)
		require.NoError(t, err)
		require.NotNil(t, uc)
		assert.Equal(t, omaha.UpdateOK, uc.Status)
	})

	t.Run("no app element returns an error, not a panic", func(t *testing.T) {
		body, err := xml.Marshal(omaha.NewResponse())
		require.NoError(t, err)

		uc, err := parseOmahaUpdateResponse(body)
		require.Error(t, err)
		assert.Nil(t, uc)
		assert.Contains(t, err.Error(), "no apps")
	})

	t.Run("app without updatecheck returns nil and no error", func(t *testing.T) {
		resp := omaha.NewResponse()
		resp.AddApp(appID, omaha.AppOK)
		body, err := xml.Marshal(resp)
		require.NoError(t, err)

		uc, err := parseOmahaUpdateResponse(body)
		require.NoError(t, err)
		assert.Nil(t, uc)
	})

	t.Run("malformed body returns an error", func(t *testing.T) {
		_, err := parseOmahaUpdateResponse([]byte("not xml at all"))
		require.Error(t, err)
	})
}
