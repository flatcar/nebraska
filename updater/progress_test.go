package updater

import (
	"context"
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProgressToEventRequest verifies that each progress value maps to the
// correct Omaha event type and result.
func TestProgressToEventRequest(t *testing.T) {
	tests := []struct {
		name       string
		progress   progress
		wantType   omaha.EventType
		wantResult omaha.EventResult
		wantNil    bool
	}{
		{
			name:       "ProgressDownloadStarted maps to EventTypeUpdateDownloadStarted",
			progress:   ProgressDownloadStarted,
			wantType:   omaha.EventTypeUpdateDownloadStarted,
			wantResult: omaha.EventResultSuccess,
		},
		{
			name:       "ProgressDownloadFinished maps to EventTypeUpdateDownloadFinished",
			progress:   ProgressDownloadFinished,
			wantType:   omaha.EventTypeUpdateDownloadFinished,
			wantResult: omaha.EventResultSuccess,
		},
		{
			name:       "ProgressInstallationStarted maps to EventTypeInstallStarted",
			progress:   ProgressInstallationStarted,
			wantType:   omaha.EventTypeInstallStarted,
			wantResult: omaha.EventResultSuccess,
		},
		{
			name:       "ProgressInstallationFinished maps to EventTypeInstallComplete",
			progress:   ProgressInstallationFinished,
			wantType:   omaha.EventTypeInstallComplete,
			wantResult: omaha.EventResultSuccess,
		},
		{
			name:       "ProgressUpdateComplete maps to EventTypeUpdateComplete",
			progress:   ProgressUpdateComplete,
			wantType:   omaha.EventTypeUpdateComplete,
			wantResult: omaha.EventResultSuccess,
		},
		{
			name:       "ProgressUpdateCompleteAndRestarted maps to EventTypeUpdateComplete with reboot result",
			progress:   ProgressUpdateCompleteAndRestarted,
			wantType:   omaha.EventTypeUpdateComplete,
			wantResult: omaha.EventResultSuccessReboot,
		},
		{
			name:       "ProgressError maps to EventTypeUpdateComplete with error result",
			progress:   ProgressError,
			wantType:   omaha.EventTypeUpdateComplete,
			wantResult: omaha.EventResultError,
		},
		{
			name:     "unknown progress value returns nil",
			progress: progress(999),
			wantNil:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := progressToEventRequest(tc.progress)
			if tc.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tc.wantType, got.Type)
			assert.Equal(t, tc.wantResult, got.Result)
		})
	}
}

// TestProgressInstallationStartedAndFinishedAreDistinct verifies that
// ProgressInstallationStarted and ProgressInstallationFinished produce
// different event types.
func TestProgressInstallationStartedAndFinishedAreDistinct(t *testing.T) {
	started := progressToEventRequest(ProgressInstallationStarted)
	finished := progressToEventRequest(ProgressInstallationFinished)

	require.NotNil(t, started)
	require.NotNil(t, finished)

	assert.NotEqual(t, started.Type, finished.Type)
	assert.Equal(t, omaha.EventTypeInstallStarted, started.Type)
	assert.Equal(t, omaha.EventTypeInstallComplete, finished.Type)
}

// recordingHandler is a test UpdateHandler that records which methods were called.
type recordingHandler struct {
	fetchCalled bool
	applyCalled bool
}

func (h *recordingHandler) FetchUpdate(_ context.Context, _ UpdateInfo) error {
	h.fetchCalled = true
	return nil
}

func (h *recordingHandler) ApplyUpdate(_ context.Context, _ UpdateInfo) error {
	h.applyCalled = true
	return nil
}

// recordingOmahaHandler captures all Omaha event types sent during TryUpdate.
type recordingOmahaHandler struct {
	eventTypes []omaha.EventType
	hasUpdate  bool
}

func (h *recordingOmahaHandler) Handle(_ context.Context, _ string, req *omaha.Request) (*omaha.Response, error) {
	resp := omaha.NewResponse()
	for _, app := range req.Apps {
		respApp := resp.AddApp(string(app.ID), omaha.AppOK)

		// Record any events sent
		for _, event := range app.Events {
			h.eventTypes = append(h.eventTypes, event.Type)
			respApp.AddEvent()
		}

		// Respond to update check
		if app.UpdateCheck != nil {
			if h.hasUpdate {
				uc := respApp.AddUpdateCheck(omaha.UpdateOK)
				m := uc.AddManifest("2.0.0")
				m.AddPackage()
				uc.AddURL("http://example.com/")
			} else {
				respApp.AddUpdateCheck(omaha.NoUpdate)
			}
		}
	}
	return resp, nil
}

// TestTryUpdateReportsAllProgressEvents verifies that TryUpdate sends the
// complete sequence of Omaha progress events:
//
//	DownloadStarted → DownloadFinished → InstallStarted → InstallComplete → UpdateComplete
//
// Before the fix, TryUpdate skipped DownloadStarted and InstallationStarted,
// so the server only saw: DownloadFinished → InstallComplete → UpdateComplete.
func TestTryUpdateReportsAllProgressEvents(t *testing.T) {
	omahaHandler := &recordingOmahaHandler{hasUpdate: true}

	u, err := New(Config{
		OmahaURL:        "http://localhost",
		AppID:           "{test-app-id}",
		Channel:         "stable",
		InstanceID:      "test-instance",
		InstanceVersion: "1.0.0",
		OmahaReqHandler: omahaHandler,
	})
	require.NoError(t, err)

	handler := &recordingHandler{}
	err = u.TryUpdate(context.Background(), handler)
	require.NoError(t, err)

	assert.True(t, handler.fetchCalled, "FetchUpdate should have been called")
	assert.True(t, handler.applyCalled, "ApplyUpdate should have been called")

	wantEvents := []omaha.EventType{
		omaha.EventTypeUpdateDownloadStarted,
		omaha.EventTypeUpdateDownloadFinished,
		omaha.EventTypeInstallStarted,
		omaha.EventTypeInstallComplete,
		omaha.EventTypeUpdateComplete,
	}

	assert.Equal(t, wantEvents, omahaHandler.eventTypes,
		"TryUpdate must report the complete event sequence to the Omaha server")
}

// TestTryUpdateNoUpdateReportsNoEvents verifies that when no update is
// available, TryUpdate reports no events.
func TestTryUpdateNoUpdateReportsNoEvents(t *testing.T) {
	omahaHandler := &recordingOmahaHandler{hasUpdate: false}

	u, err := New(Config{
		OmahaURL:        "http://localhost",
		AppID:           "{test-app-id}",
		Channel:         "stable",
		InstanceID:      "test-instance",
		InstanceVersion: "1.0.0",
		OmahaReqHandler: omahaHandler,
	})
	require.NoError(t, err)

	err = u.TryUpdate(context.Background(), &recordingHandler{})
	require.Error(t, err)
	var noUpdateErr NoUpdateError
	assert.ErrorAs(t, err, &noUpdateErr)
	assert.Empty(t, omahaHandler.eventTypes, "no events should be sent when no update is available")
}
