package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailNotifier(t *testing.T) {
	sent := false
	n := newEmailNotifier(Config{
		EmailFrom: "nebraska@example.com",
		EmailTo:   []string{"admin@example.com"},
		SMTPHost:  "smtp.example.com",
		SMTPPort:  25,
	})

	n.sendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sent = true
		assert.Equal(t, "smtp.example.com:25", addr)
		assert.Equal(t, "nebraska@example.com", from)
		assert.Equal(t, []string{"admin@example.com"}, to)
		assert.Contains(t, string(msg), "Subject: [Nebraska] New Flatcar package synced: 3578.0.0 (stable, amd64)")
		return nil
	}

	err := n.NotifySync(SyncEvent{
		Channel:  "stable",
		Arch:     "amd64",
		Version:  "3578.0.0",
		Filename: "flatcar-amd64-3578.0.0.gz",
		SyncedAt: time.Now(),
	})

	require.NoError(t, err)
	assert.True(t, sent)
}

func TestWebhookNotifier(t *testing.T) {
	received := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)
		assert.Equal(t, "3578.0.0", payload["version"])
		assert.Equal(t, "stable", payload["channel"])
		assert.Equal(t, "amd64", payload["arch"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := newWebhookNotifier(server.URL)
	err := n.NotifySync(SyncEvent{
		Channel:  "stable",
		Arch:     "amd64",
		Version:  "3578.0.0",
		Filename: "flatcar-amd64-3578.0.0.gz",
		SyncedAt: time.Now(),
	})

	require.NoError(t, err)
	assert.True(t, received)
}
