package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/flatcar/nebraska/backend/pkg/logger"
)

var l = logger.New("notifier")

// SyncEvent represents the payload sent when a new package is synced.
type SyncEvent struct {
	Channel  string    `json:"channel"`
	Arch     string    `json:"arch"`
	Version  string    `json:"version"`
	Filename string    `json:"filename"`
	SyncedAt time.Time `json:"synced_at"`
}

// Notifier defines the interface for dispatching notifications on sync events.
type Notifier interface {
	NotifySync(event SyncEvent) error
}

// Config holds settings for notification dispatchers.
type Config struct {
	EmailFrom  string
	EmailTo    []string
	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	WebhookURL string
}

type multiNotifier struct {
	notifiers []Notifier
}

// New creates a combined Notifier from the given configuration.
func New(conf Config) Notifier {
	var notifiers []Notifier
	if conf.SMTPHost != "" && len(conf.EmailTo) > 0 {
		notifiers = append(notifiers, newEmailNotifier(conf))
	}
	if conf.WebhookURL != "" {
		notifiers = append(notifiers, newWebhookNotifier(conf.WebhookURL))
	}
	return &multiNotifier{notifiers: notifiers}
}

func (m *multiNotifier) NotifySync(event SyncEvent) error {
	var errs []string
	for _, n := range m.notifiers {
		if err := n.NotifySync(event); err != nil {
			l.Error().Err(err).Msg("failed to send notification")
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

type emailNotifier struct {
	conf Config
	sendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

func newEmailNotifier(conf Config) *emailNotifier {
	return &emailNotifier{
		conf:     conf,
		sendMail: smtp.SendMail,
	}
}

func (e *emailNotifier) NotifySync(event SyncEvent) error {
	subject := fmt.Sprintf("[Nebraska] New Flatcar package synced: %s (%s, %s)", event.Version, event.Channel, event.Arch)
	body := fmt.Sprintf("A new Flatcar package has been successfully synced in Nebraska.\n\n"+
		"Channel: %s\n"+
		"Architecture: %s\n"+
		"Version: %s\n"+
		"Filename: %s\n"+
		"Synced At: %s\n",
		event.Channel, event.Arch, event.Version, event.Filename, event.SyncedAt.Format(time.RFC1123))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.conf.EmailFrom, strings.Join(e.conf.EmailTo, ","), subject, body)

	addr := fmt.Sprintf("%s:%d", e.conf.SMTPHost, e.conf.SMTPPort)
	var auth smtp.Auth
	if e.conf.SMTPUser != "" {
		auth = smtp.PlainAuth("", e.conf.SMTPUser, e.conf.SMTPPass, e.conf.SMTPHost)
	}

	return e.sendMail(addr, auth, e.conf.EmailFrom, e.conf.EmailTo, []byte(msg))
}

type webhookNotifier struct {
	webhookURL string
	client     *http.Client
}

func newWebhookNotifier(webhookURL string) *webhookNotifier {
	return &webhookNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *webhookNotifier) NotifySync(event SyncEvent) error {
	payload := map[string]interface{}{
		"text":      fmt.Sprintf("New Flatcar package synced: *%s* (Channel: %s, Arch: %s)", event.Version, event.Channel, event.Arch),
		"event":     "package_synced",
		"channel":   event.Channel,
		"arch":      event.Arch,
		"version":   event.Version,
		"filename":  event.Filename,
		"synced_at": event.SyncedAt.Format(time.RFC3339),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := w.client.Post(w.webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
	}

	return nil
}
