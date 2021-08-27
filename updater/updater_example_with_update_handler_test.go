package updater

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type exampleUpdateHandler struct{}

func (e exampleUpdateHandler) FetchUpdate(ctx context.Context, info UpdateInfo) error {
	// download, err := someFunctionThatDownloadsAFile(ctx, info.GetURL())
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (e exampleUpdateHandler) ApplyUpdate(ctx context.Context, info UpdateInfo) error {
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
func ExampleUpdater_withUpdateHandler() error {
	conf := Config{
		OmahaURL:        "http://test.omahaserver.com/v1/update/",
		AppID:           "application_id",
		Channel:         "stable",
		InstanceID:      uuid.NewString(),
		InstanceVersion: "0.0.1",
	}

	appUpdater, err := New(conf)
	if err != nil {
		return fmt.Errorf("init updater: %w", err)
	}

	ctx := context.TODO()

	if err := appUpdater.TryUpdate(ctx, exampleUpdateHandler{}); err != nil {
		return fmt.Errorf("trying update: %w", err)
	}

	return nil
}
