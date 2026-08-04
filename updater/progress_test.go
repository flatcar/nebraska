package updater

import (
	"testing"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProgressToEventRequest verifies that each progress value maps to the
// correct Omaha event type and result. This is a regression test for the bug
// where ProgressInstallationFinished incorrectly mapped to EventTypeInstallStarted
// instead of EventTypeInstallComplete.
func TestProgressToEventRequest(t *testing.T) {
	tests := []struct {
		name           string
		progress       progress
		wantType       omaha.EventType
		wantResult     omaha.EventResult
		wantNil        bool
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
			// Regression test for: ProgressInstallationFinished was incorrectly
			// mapping to EventTypeInstallStarted instead of EventTypeInstallComplete,
			// making installation start and finish indistinguishable in server logs.
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
			name:    "unknown progress value returns nil",
			progress: progress(999),
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := progressToEventRequest(tc.progress)

			if tc.wantNil {
				assert.Nil(t, got, "expected nil for unknown progress value")
				return
			}

			require.NotNil(t, got, "expected non-nil EventRequest")
			assert.Equal(t, tc.wantType, got.Type,
				"progress %v: wrong event type", tc.progress)
			assert.Equal(t, tc.wantResult, got.Result,
				"progress %v: wrong event result", tc.progress)
		})
	}
}

// TestProgressInstallationStartedAndFinishedAreDistinct verifies that
// ProgressInstallationStarted and ProgressInstallationFinished produce
// different event types. This is the core assertion of the bug fix.
func TestProgressInstallationStartedAndFinishedAreDistinct(t *testing.T) {
	started := progressToEventRequest(ProgressInstallationStarted)
	finished := progressToEventRequest(ProgressInstallationFinished)

	require.NotNil(t, started)
	require.NotNil(t, finished)

	assert.NotEqual(t, started.Type, finished.Type,
		"ProgressInstallationStarted and ProgressInstallationFinished must produce different event types")

	assert.Equal(t, omaha.EventTypeInstallStarted, started.Type,
		"ProgressInstallationStarted must produce EventTypeInstallStarted")

	assert.Equal(t, omaha.EventTypeInstallComplete, finished.Type,
		"ProgressInstallationFinished must produce EventTypeInstallComplete")
}
