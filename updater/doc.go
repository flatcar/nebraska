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
  1. Check for an update, if there's one available then. If there is not, try again later.
  2. Inform the server we're started it (this is done by sending a progress
	 report of "download started").
  3. Actually perform the download or whatever represents fetching the
     update's parts.
  4. Inform the server that the download is finished it and that we're applying
     the update (this is done by sending a progress report of "installation
	 started").
  5. Apply the update (this deeply depends on what each application does, but
	 may involve extracting files into locations, migrating configuration,
	 etc.).
  6. Inform the server that the update installation is finished; run the new
    version of the application and report that the update is now complete
	(these are two different progress reports and may involve running).

Note that if the update requires a restart, then there's a specific progress
report for that.
The caller is also responsible for keeping any local state the update
implementation needs (like e.g. knowing that a restart has happened, or that the
version if now running).
*/
package updater
