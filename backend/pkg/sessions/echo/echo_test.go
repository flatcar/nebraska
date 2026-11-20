package echo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/flatcar/nebraska/backend/pkg/sessions"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

func TestSaveSessionSetsSecureCookie(t *testing.T) {
	for _, tc := range []struct {
		name       string
		target     string
		header     map[string]string
		wantSecure bool
	}{
		{
			name:       "direct TLS connection",
			target:     "https://nebraska.example.com/login",
			wantSecure: true,
		},
		{
			name:       "behind a TLS-terminating proxy",
			target:     "http://nebraska.example.com/login",
			header:     map[string]string{echo.HeaderXForwardedProto: "https"},
			wantSecure: true,
		},
		{
			name:   "plain HTTP",
			target: "http://nebraska.example.com/login",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			codec := sessions.NewMockCodec()
			codec.AddIDValueMapping("id1", "val1")
			store := sessions.NewStore(sessions.NewMockCache(), codec)

			request := httptest.NewRequest(http.MethodGet, tc.target, nil)
			for key, value := range tc.header {
				request.Header.Set(key, value)
			}
			recorder := httptest.NewRecorder()
			ctx := echo.New().NewContext(request, recorder)

			session := store.GetSessionUse(request, "test")
			assert.NoError(t, SaveSession(ctx, session))

			cookies := recorder.Result().Cookies()
			if assert.Len(t, cookies, 1) {
				cookie := cookies[0]
				assert.Equal(t, tc.wantSecure, cookie.Secure)
				assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
				assert.True(t, cookie.HttpOnly)
			}
		})
	}
}
