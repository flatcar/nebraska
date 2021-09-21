package updater_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kinvolk/nebraska/updater"
)

func someFunctionThatDownloadsAFile(ctx context.Context, url string) (string, error) {
	// Download file logic goes here
	return "/tmp/downloads/examplefile.txt", nil
}

func someFunctionThatExtractsTheUpdateAndInstallIt(ctx context.Context, filePath string) error {
	// Extract and install update logic goes here
	return nil
}

// ExampleUpdater shows how to use the updater package to
// update an application manually.
func ExampleUpdater() error {
	conf := updater.Config{
		OmahaURL:        "http://test.omahaserver.com/v1/update/",
		AppID:           "application_id",
		Channel:         "stable",
		InstanceID:      uuid.NewString(),
		InstanceVersion: "0.0.1",
	}

	appUpdater, err := updater.New(conf)
	if err != nil {
		return fmt.Errorf("init updater: %w", err)
	}

	ctx := context.TODO()

	updateInfo, err := appUpdater.CheckForUpdates(ctx)
	if err != nil {
		return fmt.Errorf("checking updates for app: %q, err: %w", conf.AppID, err)
	}

	if !updateInfo.HasUpdate {
		return fmt.Errorf("No update exists for the application")
	}

	// So we got an update, let's report we'll start downloading it.
	if err := appUpdater.ReportProgress(ctx, updater.ProgressDownloadStarted); err != nil {
		if progressErr := appUpdater.ReportError(ctx, nil); progressErr != nil {
			fmt.Println("Reporting progress error:", progressErr)
		}
		return fmt.Errorf("reporting download started: %w", err)
	}

	// This should be implemented by the caller.
	filePath, err := someFunctionThatDownloadsAFile(ctx, updateInfo.URL())
	if err != nil {
		// Oops something went wrong.
		if progressErr := appUpdater.ReportError(ctx, nil); progressErr != nil {
			fmt.Println("reporting error:", progressErr)
		}
		return fmt.Errorf("downloading update: %w", err)
	}

	// The download was successful, let's inform that to the Omaha server.
	if err := appUpdater.ReportProgress(ctx, updater.ProgressDownloadFinished); err != nil {
		if progressErr := appUpdater.ReportError(ctx, nil); progressErr != nil {
			fmt.Println("Reporting progress error:", progressErr)
		}
		return fmt.Errorf("reporting download finished: %w", err)
	}

	// We got our update file, let's install it!
	if err := appUpdater.ReportProgress(ctx, updater.ProgressInstallationStarted); err != nil {
		if progressErr := appUpdater.ReportError(ctx, nil); progressErr != nil {
			fmt.Println("reporting progress error:", progressErr)
		}
		return fmt.Errorf("reporting installation started: %w", err)
	}

	// This should be your own implementation.
	if err := someFunctionThatExtractsTheUpdateAndInstallIt(ctx, filePath); err != nil {
		// Oops something went wrong.
		if progressErr := appUpdater.ReportError(ctx, nil); progressErr != nil {
			fmt.Println("Reporting error:", progressErr)
		}
		return fmt.Errorf("applying update: %w", err)
	}

	if err := appUpdater.CompleteUpdate(ctx, updateInfo); err != nil {
		if progressErr := appUpdater.ReportError(ctx, nil); progressErr != nil {
			fmt.Println("reporting progress error:", progressErr)
		}
		return fmt.Errorf("reporting complete update: %w", err)
	}

	return nil
}
