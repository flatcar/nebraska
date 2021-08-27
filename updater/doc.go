/*
Package updater aims to simplify omaha-powered updates.

Its goal is to abstract many of the omaha-protocol details, so users can
perform updates without having to understand the omaha protocal
internals.

Since the omaha protocol is very powerful, it supports many options that
are beyond the scope of this package. So the updater package assumes that
there's only one single application involved in the update, and that the
update is being performed for one single instance of that application.

It also simplifies the update information that is represented by an omaha
response, but allows to retrieve the actual OmahaResponse object if needed.

Basics:

Each update has four main parts involved:
  1. The application: represents the software getting updated;
  2. The instance: represents the instance of the application that's getting
     updated.
  3. The instance version: this is used by the server to decide how to respond
     (whether there's a new version available or not).
  4. The channel: an application may have different channels of releases (this
	 is typical "beta", "stable", etc.).

The way omaha managed updates work is that omaha responds to update checks, and
relies on the information given by the application's instance to keep track of
each update's state. So the basic workflow for using this updater package is:
  1. Check for an update, if there's one available then:
  2. Inform the server we're started it (this is done by sending a progress
	 report of "download started").
  3. Actually perform the download or whatever represents getting fetching the
     update's parts.
  4. Inform the server that the download is finished it and that we're applying
     the update (this is done by sending a progress report of "installation
	 started").
  5. Apply the update (this deeply depends on what each application does, but
	 may involve extracting files into locations, migrating configuration,
	 etc.).
  6.Inform the server that the update installation is finished; run the new
    version of the application and report that the update is now complete
	(these are two different progress reports and may involve running).

Note that if your update requires a reboot, then there's a specific progress
report for that.
The caller is also responsible for keeping any local state the update
implementation needs (like e.g. knowing that a reboot has happened, or that the
version if now running).

Initialization:

An instance of the updater needs to be initialized with the needed details for
the updates in order for it to be used:

	import (
		"os"

		"github.com/kinvolk/nebraska/updater"
	)

	func getInstanceID() string {
		// Your own implementation here...
		return os.Getenv("MACHINE_NAME")
	}

	func getCurrentVersion() string {
		// Your own implementation here...
		return os.Getenv("MACHINE_VERSION")
	}

	u := updater.New("http://myupdateserver.io/v1/update", "io.phony.App", "stable", getInstanceID(), getCurrentVersion())


Performing updates manually:

After we have the updater instance, we can try updating:

	ctx := context.TODO()

	info, err := u.CheckForUpdates(ctx)
	if err != nil {
		fmt.Printf("oops, something didn't go well... %v\n", err)
		return
	}

	if !info.HasUpdate {
		fmt.Printf("no updates, try next time...")
		return
	}

	newVersion := info.GetVersion()

	// So we got an update, let's report we'll start downloading it
	u.ReportProgress(updater.ProgressDownloadStarted)

	// This should be your own implementation
	download, err := someFunctionThatDownloadsAFile(info.GetURL())
	if err != nil {
		// Oops something went wrong
		u.ReportProgress(updater.ProgressError)
		return
	}

	// The download was successful, let's inform of that
	u.ReportProgress(updater.ProgressDownloadFinished)

	// We got our update file, let's install it!
	u.ReportProgress(updater.ProgressInstallationStarted)

	// This should be your own implementation
	err := someFunctionThatExtractsTheUpdateAndInstallIt(download)
	if err != nil {
		// Oops something went wrong
		u.ReportProgress(updater.ProgressError)
		return
	}

	// We've just applied the update
	u.ReportProgress(updater.ProgressInstallationFinished)

	err := someFunctionThatExitsAndRerunsTheApp()
	if err != nil {
		// Oops something went wrong
		u.ReportProgress(updater.ProgressError)
		return
	}

	u.SetVersion(newVersion)

	// We're running the new version! Hurray!
	u.ReportProgress(updater.ProgressUpdateComplete)


If instead of rerunning the application in the example above, we'd perform a
reboot, then upon running the logic again and detecting that we're running a
new version, we could report that we did so:

    u.ReportProgress(updater.ProgressUpdateCompleteAndRebooted)


Performing updates, simplified:

It may be that our update process is very straightforward (doesn't need a
reboot not a lot of state checks in between) and that it can be well divided
in two parts: getting the update, applying the update.

For this use-case, updater offers a simpler way to update that sends the
progress reports automatically: TryUpdate

	// After initializing our Updater instance...

	type updateHandler struct {}

	func (e updateHandler) FetchUpdate(ctx context.Context, info *UpdateInfo) error {
		download, err := someFunctionThatDownloadsAFile(info.GetURL())
		if err != nil {
			// Oops something went wrong
			return err
		}

		setDownloadFile(ctx)

		return nil
	}

	func (e updateHandler) ApplyUpdate(ctx context.Context, info *UpdateInfo) error {
		err := someFunctionThatExtractsTheUpdateAndInstallIt(getDownloadFile(ctx))
		if err != nil {
			// Oops something went wrong
			return err
		}

		err := someFunctionThatExitsAndRerunsTheApp()
		if err != nil {
			// Oops something went wrong
			return err
		}

		return nil
	}

	err := u.TryUpdate(ctx, updateHandler{})
	if err != nil {
		// Oops something went wrong
	}

	// If the update succeeded, then u.GetInstanceVersion() should be set to the new version

*/
package updater
