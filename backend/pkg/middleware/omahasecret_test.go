package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOmahaSecret(t *testing.T) {
	const secret = "/mysecret"

	for _, tt := range []struct {
		name         string
		path         string
		wantStatus   int
		wantReached  bool
		wantRewrites string
	}{
		{
			name:         "correct secret is accepted and stripped",
			path:         "/v1/update" + secret,
			wantStatus:   http.StatusOK,
			wantReached:  true,
			wantRewrites: "/v1/update",
		},
		{
			name:        "wrong secret is rejected",
			path:        "/v1/update/wrongsecret",
			wantStatus:  http.StatusNotImplemented,
			wantReached: false,
		},
		{
			name:        "missing secret is rejected",
			path:        "/v1/update",
			wantStatus:  http.StatusNotImplemented,
			wantReached: false,
		},
		{
			name:        "secret prefix is not enough",
			path:        "/v1/update/mysec",
			wantStatus:  http.StatusNotImplemented,
			wantReached: false,
		},
		{
			name:        "secret with trailing characters is rejected",
			path:        "/v1/update" + secret + "extra",
			wantStatus:  http.StatusNotImplemented,
			wantReached: false,
		},
		{
			name:        "unrelated path is passed through untouched",
			path:        "/health",
			wantStatus:  http.StatusOK,
			wantReached: true,
			// Paths outside /v1/update are not rewritten.
			wantRewrites: "/health",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			reached := false
			handler := OmahaSecret(secret)(func(c echo.Context) error {
				reached = true
				return c.NoContent(http.StatusOK)
			})

			require.NoError(t, handler(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantReached, reached)
			if tt.wantReached {
				assert.Equal(t, tt.wantRewrites, c.Request().URL.Path)
			}
		})
	}
}
