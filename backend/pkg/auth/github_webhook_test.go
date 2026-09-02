package auth

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // matches GitHub's webhook signature scheme, not used for anything sensitive here
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWebhookRequest(t *testing.T, signature string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	body := []byte(`{"action":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/login/webhook", strings.NewReader(string(body)))
	if signature != "" {
		req.Header.Set("X-Hub-Signature", signature)
	}
	req.Header.Set("X-Github-Event", "organization")
	rec := httptest.NewRecorder()
	e := echo.New()
	return e.NewContext(req, rec), rec
}

// TestLoginWebhookMalformedSignature checks that signature headers shorter
// than the "sha1=" prefix (or missing it) are rejected with 400 instead of
// panicking on the old signature[5:] slice.
func TestLoginWebhookMalformedSignature(t *testing.T) {
	tests := []struct {
		name      string
		signature string
	}{
		{"missing header", ""},
		{"empty string", ""},
		{"shorter than prefix", "abc"},
		{"exactly the prefix length, wrong content", "sha1x"},
		{"no equals sign", "sha1abc"},
	}

	gha := NewGithubAuthenticator(&GithubAuthConfig{WebhookSecret: "test-secret"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := newWebhookRequest(t, tt.signature)

			require.NotPanics(t, func() {
				err := gha.LoginWebhook(ctx)
				assert.NoError(t, err)
			})

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// TestLoginWebhookValidSignature checks that a correctly signed payload is
// still accepted after switching from signature[5:] to strings.CutPrefix.
func TestLoginWebhookValidSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"action":"test"}`)
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(payload)
	signature := "sha1=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/login/webhook", strings.NewReader(string(payload)))
	req.Header.Set("X-Hub-Signature", signature)
	req.Header.Set("X-Github-Event", "organization")
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)

	gha := NewGithubAuthenticator(&GithubAuthConfig{WebhookSecret: secret})

	require.NotPanics(t, func() {
		err := gha.LoginWebhook(ctx)
		assert.NoError(t, err)
	})

	// A valid signature is accepted and processed rather than rejected
	// with the 400 used for a missing/malformed signature.
	assert.NotEqual(t, http.StatusBadRequest, rec.Code)
}
