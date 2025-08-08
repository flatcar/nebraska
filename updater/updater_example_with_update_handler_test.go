package updater_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/flatcar/nebraska/updater"
)

type exampleUpdateHandler struct{}

func (e exampleUpdateHandler) FetchUpdate(_ context.Context, _ updater.UpdateInfo) error {
	// download, err := someFunctionThatDownloadsAFile(ctx, info.GetURL())
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (e exampleUpdateHandler) ApplyUpdate(_ context.Context, _ updater.UpdateInfo) error {
	// err := someFunctionThatExtractsTheUpdateAndInstallIt(ctx, getDownloadFile(ctx))
	// if err != nil {
	// 	// Oops something went wrong
	// 	return err
	// }

	// err := someFunctionThatExitsAndRerunsTheApp(ctx)
	// if err != nil {
	// 	// Oops something went wrong
	// 	return err
	// }
	return nil
}

// ExampleUpdater_withUpdateHandler shows how to use the updater package to
// update an application automatically using exampleUpdateHandler.
func ExampleUpdater_withUpdateHandler() { //nolint: govet
	conf := updater.Config{
		OmahaURL:        "http://test.omahaserver.com/v1/update/",
		AppID:           "application_id",
		Channel:         "stable",
		InstanceID:      uuid.NewString(),
		InstanceVersion: "0.0.1",
	}

	appUpdater, err := updater.New(conf)
	if err != nil {
		fmt.Println(fmt.Errorf("init updater: %w", err))
	}

	ctx := context.TODO()

	if err := appUpdater.TryUpdate(ctx, exampleUpdateHandler{}); err != nil {
		fmt.Println(fmt.Errorf("trying update: %w", err))
	}
}
