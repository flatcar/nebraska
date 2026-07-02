package syncer

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/logger"
	"github.com/flatcar/nebraska/backend/pkg/tlsutil"
)

const (
	flatcarAppID = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"
)

var (
	l = logger.New("syncer")

	// ErrInvalidAPIInstance error indicates that no valid api instance was
	// provided to the syncer constructor.
	ErrInvalidAPIInstance = errors.New("invalid api instance")
)

type channelDescriptor struct {
	name string
	arch api.Arch
}

// Syncer represents a process in charge of checking for updates in the
// different official Flatcar channels and updating the Flatcar application in
// Nebraska as needed (creating new packages and updating channels to point
// to them). When hostPackages is enabled, packages payloads will be downloaded
// into packagesPath and package url/filename will be rewritten.
type Syncer struct {
	api               *api.API
	hostPackages      bool
	packagesPath      string
	packagesURL       string
	checkFrequency    time.Duration
	flatcarUpdatesURL string
	stopCh            chan struct{}
	stopOnce          sync.Once
	machinesIDs       map[channelDescriptor]string
	bootIDs           map[channelDescriptor]string
	versions          map[channelDescriptor]string
	channelsIDs       map[channelDescriptor]string
	httpClient        *http.Client
	ticker            *time.Ticker
}

// Config represents the configuration used to create a new Syncer instance.
type Config struct {
	API               *api.API
	HostPackages      bool
	PackagesPath      string
	PackagesURL       string
	FlatcarUpdatesURL string
	CheckFrequency    time.Duration
	HTTPClient        *http.Client
}

// Setup creates a new syncer from config and db connection, and returns it.
func Setup(conf *config.Config, db *api.API) (*Syncer, error) {
	checkFrequency, err := time.ParseDuration(conf.CheckFrequencyVal)
	if err != nil {
		return nil, fmt.Errorf("invalid Check Frequency value: %w", err)
	}

	httpClient := tlsutil.NewHTTPClient(conf.CACertPool)

	if conf.SyncerPkgsURL == "" && conf.HostFlatcarPackages {
		conf.SyncerPkgsURL = conf.NebraskaURL + "/flatcar/"
	}

	syncer, err := New(&Config{
		API:               db,
		HostPackages:      conf.HostFlatcarPackages,
		PackagesPath:      conf.FlatcarPackagesPath,
		PackagesURL:       conf.SyncerPkgsURL,
		FlatcarUpdatesURL: conf.FlatcarUpdatesURL,
		CheckFrequency:    checkFrequency,
		HTTPClient:        httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("error setting up syncer: %w", err)
	}
	return syncer, nil
}

// New creates a new Syncer instance.
func New(conf *Config) (*Syncer, error) {
	if conf.API == nil {
		return nil, ErrInvalidAPIInstance
	}

	if conf.PackagesURL != "" {
		if _, err := url.Parse(conf.PackagesURL); err != nil {
			return nil, fmt.Errorf("invalid package url: %w", err)
		}
	}

	s := &Syncer{
		api:               conf.API,
		hostPackages:      conf.HostPackages,
		packagesPath:      conf.PackagesPath,
		packagesURL:       conf.PackagesURL,
		flatcarUpdatesURL: conf.FlatcarUpdatesURL,
		checkFrequency:    conf.CheckFrequency,
		stopCh:            make(chan struct{}),
		machinesIDs:       make(map[channelDescriptor]string, 8),
		bootIDs:           make(map[channelDescriptor]string, 8),
		channelsIDs:       make(map[channelDescriptor]string, 8),
		versions:          make(map[channelDescriptor]string, 8),
		httpClient:        conf.HTTPClient,
	}

	if s.httpClient == nil {
		s.httpClient = &http.Client{}
	}

	if err := s.initialize(); err != nil {
		return nil, err
	}

	return s, nil
}

// Start makes the syncer start working. It will check for updates every
// checkFrequency until it's asked to stop.
func (s *Syncer) Start() {
	l.Debug().Msg("syncer ready!")
	s.ticker = time.NewTicker(s.checkFrequency)

	_ = s.checkForUpdates()

L:
	for {
		select {
		case <-s.ticker.C:
			_ = s.checkForUpdates()
		case <-s.stopCh:
			break L
		}
	}

	s.api.Close()
}

// Stop stops the polling for updates.
func (s *Syncer) Stop() {
	s.ticker.Stop()
	l.Debug().Msg("stopping syncer..")
	// Close (rather than send) so every receiver observes the stop: both the
	// main Start loop and any in-flight checkForUpdates pass. A single sent token
	// could be consumed by checkForUpdates, leaving Start blocked forever.
	// sync.Once guards against a panic if Stop is ever called more than once.
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

// initialize does some initial setup to prepare the syncer, checking in
// Nebraska the last versions we know about for the different channels in the
// Flatcar application and keeping track of some ids.
func (s *Syncer) initialize() error {
	flatcarApp, err := s.api.GetApp(flatcarAppID)
	if err != nil {
		return err
	}

	for _, c := range flatcarApp.Channels {
		if c.Name == "stable" || c.Name == "beta" || c.Name == "alpha" || c.Name == "edge" || strings.HasPrefix(c.Name, "lts-") {
			descriptor := channelDescriptor{
				name: c.Name,
				arch: c.Arch,
			}
			s.machinesIDs[descriptor] = "{" + uuid.New().String() + "}"
			s.bootIDs[descriptor] = "{" + uuid.New().String() + "}"
			s.channelsIDs[descriptor] = c.ID

			if c.Package != nil {
				s.versions[descriptor] = c.Package.Version
			} else {
				s.versions[descriptor] = "766.0.0"
			}
		}
	}

	return nil
}

// checkForUpdates polls the public Flatcar servers looking for updates in the
// official channels (stable, beta, alpha, edge) sending Omaha requests. When an
// update is received we'll process it, creating packages and updating channels
// in Nebraska as needed.
func (s *Syncer) checkForUpdates() error {
	for descriptor, currentVersion := range s.versions {
		l.Debug().Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Str("currentVersion", currentVersion).Msg("checking for updates")

		update, err := s.doOmahaRequest(descriptor, currentVersion)
		if err != nil {
			return err
		}
		if update != nil && update.Status == "ok" && len(update.Manifests) > 0 {
			// processUpdate handles version tracking internally when appropriate
			if err := s.processUpdate(descriptor, update); err != nil {
				return err
			}
		} else {
			status := "no updatecheck in response"
			if update != nil {
				status = string(update.Status)
			}
			l.Debug().Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Str("currentVersion", currentVersion).Msgf("checkForUpdates, no update available updateStatus %v", status)
		}

		select {
		case <-time.After(5 * time.Second):
		case <-s.stopCh:
			// Stop requested: abort the walk immediately instead of continuing
			// to poll the remaining channels.
			return nil
		}
	}

	return nil
}

// doOmahaRequest sends an Omaha request checking if there is an update for a
// specific Flatcar channel, returning the update check to the caller.
func (s *Syncer) doOmahaRequest(descriptor channelDescriptor, currentVersion string) (*omaha.UpdateResponse, error) {
	req := omaha.NewRequest()
	req.OS.Version = "Chateau"
	req.OS.Platform = "CoreOS"
	req.OS.ServicePack = currentVersion + "_x86_64"
	req.OS.Arch = descriptor.arch.OmahaString()
	req.Version = "CoreOSUpdateEngine-0.1.0.0"
	req.UpdaterVersion = "CoreOSUpdateEngine-0.1.0.0"
	req.InstallSource = "scheduler"
	req.IsMachine = 1
	app := req.AddApp(flatcarAppID, currentVersion)
	app.AddUpdateCheck()
	app.MachineID = s.machinesIDs[descriptor]
	app.BootID = s.bootIDs[descriptor]
	app.Track = descriptor.name
	app.MultiManifestOK = true

	payload, err := xml.Marshal(req)
	if err != nil {
		l.Error().Err(err).Msg("checkForUpdates, marshalling request xml")
		return nil, err
	}
	l.Debug().Str("request", string(payload)).Msg("doOmahaRequest")

	resp, err := s.httpClient.Post(s.flatcarUpdatesURL, "text/xml", bytes.NewReader(payload))
	if err != nil {
		l.Error().Err(err).Msg("checkForUpdates, posting omaha response")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error().Err(err).Msg("checkForUpdates, reading omaha response")
		return nil, err
	}
	l.Debug().Str("response", string(body)).Msg("doOmahaRequest")

	update, err := parseOmahaUpdateResponse(body)
	if err != nil {
		l.Error().Err(err).Msg("checkForUpdates, parsing omaha response")
		return nil, err
	}

	return update, nil
}

// parseOmahaUpdateResponse unmarshals an Omaha response body and returns the
// update check of its first app. It returns an error (rather than panicking)
// when the body is well-formed XML but carries no <app> element. The returned
// update check may be nil when the app has no <updatecheck>.
func parseOmahaUpdateResponse(body []byte) (*omaha.UpdateResponse, error) {
	oresp := &omaha.Response{}
	if err := xml.Unmarshal(body, oresp); err != nil {
		return nil, err
	}
	if len(oresp.Apps) == 0 {
		return nil, fmt.Errorf("omaha response contained no apps")
	}
	return oresp.Apps[0].UpdateCheck, nil
}

// processUpdate is in charge of creating packages in the Flatcar application in
// Nebraska and updating the appropriate channel to point to the new package.
func (s *Syncer) processUpdate(descriptor channelDescriptor, update *omaha.UpdateResponse) error {
	if len(update.Manifests) == 0 {
		return fmt.Errorf("no manifests in update response")
	}

	// The upstream serves one flagged package at a time (is_floor/is_target); a
	// syncer walks the floors and only advances its channel on the target.
	return s.processSingleManifestUpdate(descriptor, update)
}

// processSingleManifestUpdate handles a single flagged package from the upstream.
// A floor is recorded and only the walk cursor is advanced (the channel stays put);
// the target - or a legacy manifest carrying neither flag - advances the channel.
func (s *Syncer) processSingleManifestUpdate(descriptor channelDescriptor, update *omaha.UpdateResponse) error {
	manifest := update.Manifests[0]

	// Get or create the package
	pkg, err := s.getOrCreatePackage(descriptor, manifest, update)
	if err != nil {
		l.Error().Err(err).
			Str("channel", descriptor.name).
			Str("arch", descriptor.arch.String()).
			Str("version", manifest.Version).
			Msg("processSingleManifestUpdate - failed to process package")
		return err
	}

	// Record the floor BEFORE advancing, so a failure retries the same floor and
	// never skips it.
	if manifest.IsFloor {
		if err := s.markPackageAsFloor(descriptor, pkg, manifest); err != nil {
			return err
		}
	}

	// An intermediate floor advances only the walk cursor, leaving the channel
	// where it is (the channel must never point at an intermediate). The target -
	// or a legacy manifest with neither flag - advances the channel directly.
	if manifest.IsFloor && !manifest.IsTarget {
		s.advanceCursor(descriptor, pkg)
		return nil
	}

	if err := s.updateChannelToPackage(descriptor, pkg); err != nil {
		l.Error().Err(err).
			Str("channel", descriptor.name).
			Str("arch", descriptor.arch.String()).
			Msg("processSingleManifestUpdate - failed to update channel")
		return err
	}

	return nil
}

func getArchString(arch api.Arch) string {
	return strings.TrimSuffix(arch.CoreosString(), "-usr")
}

// cleanupDownloadedFiles removes downloaded files from disk, used for error recovery
func (s *Syncer) cleanupDownloadedFiles(filenames []string) {
	for _, filename := range filenames {
		if filename == "" {
			continue
		}
		filePath := filepath.Join(s.packagesPath, filename)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			l.Error().Err(err).Str("file", filePath).Msg("Failed to cleanup downloaded file")
		}
	}
}

// getOrCreatePackage checks if a package exists or creates a new one
func (s *Syncer) getOrCreatePackage(
	descriptor channelDescriptor,
	manifest *omaha.Manifest,
	update *omaha.UpdateResponse,
) (*api.Package, error) {
	version := manifest.Version

	// Check if package already exists
	pkg, err := s.api.GetPackageByVersionAndArch(flatcarAppID, version, descriptor.arch)
	if err == nil && pkg != nil {
		// Package exists. Verify it matches what the upstream advertises, but do NOT
		// fail the walk on a mismatch - that would wedge a syncer permanently on a
		// re-published version. Surface it and keep using the stored package.
		if verr := s.verifyPackageIntegrity(pkg, manifest); verr != nil {
			l.Warn().Err(verr).
				Str("channel", descriptor.name).
				Str("arch", descriptor.arch.String()).
				Str("version", version).
				Msg("existing package differs from upstream manifest; using stored package")
		}
		return pkg, nil
	}

	// Package doesn't exist - create it
	return s.createPackage(descriptor, manifest, update)
}

// verifyPackageIntegrity verifies that an existing package matches manifest data
func (s *Syncer) verifyPackageIntegrity(pkg *api.Package, manifest *omaha.Manifest) error {
	if len(manifest.Packages) == 0 {
		return fmt.Errorf("manifest has no packages")
	}
	omahaPkg := manifest.Packages[0]

	// Verify SHA1 hash
	if pkg.Hash.String != omahaPkg.SHA1 {
		return fmt.Errorf("package %s hash mismatch: expected %s, got %s",
			pkg.Version, omahaPkg.SHA1, pkg.Hash.String)
	}

	// Verify size
	expectedSize := strconv.FormatUint(omahaPkg.Size, 10)
	if pkg.Size.String != expectedSize {
		return fmt.Errorf("package %s size mismatch: expected %s, got %s",
			pkg.Version, expectedSize, pkg.Size.String)
	}

	// Verify FlatcarAction exists if manifest has actions
	if len(manifest.Actions) > 0 && pkg.FlatcarAction == nil {
		return fmt.Errorf("package %s missing FlatcarAction", pkg.Version)
	}

	return nil
}

// createPackage creates a new package from manifest data
func (s *Syncer) createPackage(
	descriptor channelDescriptor,
	manifest *omaha.Manifest,
	update *omaha.UpdateResponse,
) (*api.Package, error) {
	version := manifest.Version
	if len(manifest.Packages) == 0 {
		return nil, fmt.Errorf("manifest %s has no packages", version)
	}
	omahaPkg := manifest.Packages[0]

	// Process extra files
	extraFiles, err := s.processExtraFiles(manifest, update, descriptor, version)
	if err != nil {
		return nil, err
	}

	// Determine URL and filename
	url := update.URLs[0].CodeBase
	filename := omahaPkg.Name

	// Allow URL override
	if s.packagesURL != "" {
		url = strings.ReplaceAll(s.packagesURL, "{{VERSION}}", version)
		url = strings.ReplaceAll(url, "{{ARCH}}", getArchString(descriptor.arch))
	}

	// Handle package download if hosting is enabled
	if s.hostPackages {
		filename = fmt.Sprintf("flatcar-%s-%s.gz", getArchString(descriptor.arch), version)
		if err := s.downloadPackagePayload(manifest, update, omahaPkg, filename); err != nil {
			// Clean up already downloaded extra files on main package failure
			extraFileNames := []string{}
			for _, file := range extraFiles {
				if file.Name.Valid && file.Name.String != "" {
					extraFileNames = append(extraFileNames, file.Name.String)
				}
			}
			s.cleanupDownloadedFiles(extraFileNames)
			return nil, err
		}
	}

	// Build package object
	pkg := &api.Package{
		Type:          api.PkgTypeFlatcar,
		URL:           url,
		Version:       version,
		Filename:      null.StringFrom(filename),
		Size:          null.StringFrom(strconv.FormatUint(omahaPkg.Size, 10)),
		Hash:          null.StringFrom(omahaPkg.SHA1),
		ApplicationID: flatcarAppID,
		Arch:          descriptor.arch,
		ExtraFiles:    extraFiles,
	}

	// Add FlatcarAction if present
	if len(manifest.Actions) > 0 {
		pkg.FlatcarAction = s.buildFlatcarAction(manifest.Actions[0])
	}

	// Create package in database
	pkg, err = s.api.AddPackageWithMetadata(pkg)
	if err != nil {
		// Clean up downloaded files if DB operation failed
		if s.hostPackages {
			filesToClean := []string{filename}
			for _, file := range extraFiles {
				if file.Name.Valid && file.Name.String != "" {
					filesToClean = append(filesToClean, file.Name.String)
				}
			}
			s.cleanupDownloadedFiles(filesToClean)
		}
		return nil, err
	}

	return pkg, nil
}

// downloadPackagePayload downloads package with SHA256 handling
func (s *Syncer) downloadPackagePayload(
	manifest *omaha.Manifest,
	update *omaha.UpdateResponse,
	omahaPkg *omaha.Package,
	filename string,
) error {
	base16sha256 := ""
	// Get SHA256 from action if available
	if len(manifest.Actions) > 0 && manifest.Actions[0].SHA256 != "" {
		binsha256, err := base64.StdEncoding.DecodeString(manifest.Actions[0].SHA256)
		if err != nil {
			l.Error().Err(err).Str("version", manifest.Version).Msg("converting sha256")
			return err
		}
		base16sha256 = hex.EncodeToString(binsha256)
	}

	return s.downloadPackage(update, omahaPkg.Name, omahaPkg.SHA1, base16sha256, filename)
}

// buildFlatcarAction creates a FlatcarAction from Omaha action data
func (s *Syncer) buildFlatcarAction(action *omaha.Action) *api.FlatcarAction {
	return &api.FlatcarAction{
		Event:                 action.Event,
		ChromeOSVersion:       action.DisplayVersion,
		Sha256:                action.SHA256,
		NeedsAdmin:            action.NeedsAdmin,
		IsDelta:               action.IsDeltaPayload,
		DisablePayloadBackoff: action.DisablePayloadBackoff,
		MetadataSignatureRsa:  action.MetadataSignatureRsa,
		MetadataSize:          action.MetadataSize,
		Deadline:              action.Deadline,
		// PackageID will be set by AddPackage
	}
}

// updateChannelToPackage updates a channel to point to a specific package
func (s *Syncer) updateChannelToPackage(
	descriptor channelDescriptor,
	pkg *api.Package,
) error {
	channel, err := s.api.GetChannel(s.channelsIDs[descriptor])
	if err != nil {
		return fmt.Errorf("getting channel: %w", err)
	}

	channel.PackageID = null.StringFrom(pkg.ID)
	if err = s.api.UpdateChannel(channel); err != nil {
		return fmt.Errorf("updating channel: %w", err)
	}

	// Advance the walk cursor to the (now channel-pointed) package.
	s.advanceCursor(descriptor, pkg)

	l.Debug().
		Str("channel", descriptor.name).
		Str("arch", descriptor.arch.String()).
		Str("newVersion", pkg.Version).
		Msg("Updated channel to new version")

	return nil
}

// advanceCursor advances the syncer's walk position (the version it reports upstream)
// without touching the channel. Advancing the cursor is what makes the upstream serve
// the next floor on the following request.
func (s *Syncer) advanceCursor(descriptor channelDescriptor, pkg *api.Package) {
	s.versions[descriptor] = pkg.Version
	s.bootIDs[descriptor] = "{" + uuid.New().String() + "}"
}

// markPackageAsFloor marks a package as a floor for a specific channel
func (s *Syncer) markPackageAsFloor(descriptor channelDescriptor, pkg *api.Package, manifest *omaha.Manifest) error {
	if pkg == nil || !manifest.IsFloor {
		return nil
	}

	channelID := s.channelsIDs[descriptor]
	floorReason := null.StringFrom(manifest.FloorReason)
	if floorReason.String == "" {
		floorReason = null.StringFrom("Synced from upstream Flatcar channel")
	}

	if err := s.api.AddChannelPackageFloor(channelID, pkg.ID, floorReason); err != nil {
		l.Error().Err(err).
			Str("version", pkg.Version).
			Str("channel", descriptor.name).
			Str("arch", descriptor.arch.String()).
			Msg("markPackageAsFloor - failed to mark package as floor")
		return err
	}

	l.Debug().
		Str("version", pkg.Version).
		Str("channel", descriptor.name).
		Str("arch", descriptor.arch.String()).
		Msg("markPackageAsFloor - marked package as floor")
	return nil
}

// processExtraFiles handles extra files (signatures, metadata) from a manifest
// Returns the extra files array and downloads them if hosting is enabled
func (s *Syncer) processExtraFiles(manifest *omaha.Manifest, update *omaha.UpdateResponse, descriptor channelDescriptor, version string) ([]api.File, error) {
	var extraFiles []api.File

	if len(manifest.Packages) > 1 {
		extraFiles = make([]api.File, len(manifest.Packages)-1)
		for i := 1; i < len(manifest.Packages); i++ {
			omahaPkg := manifest.Packages[i]
			size := strconv.FormatUint(omahaPkg.Size, 10)
			extraFiles[i-1] = api.File{
				Name:    null.StringFrom(omahaPkg.Name),
				Size:    null.StringFrom(size),
				Hash:    null.StringFrom(omahaPkg.SHA1),
				Hash256: null.StringFrom(omahaPkg.SHA256),
			}
		}

		// Download extra files if hosting is enabled
		if s.hostPackages {
			downloadedFiles := []string{}
			for i := range extraFiles {
				fileInfo := &extraFiles[i]
				downloadName := fmt.Sprintf("extrafile-%s-%s-%s", getArchString(descriptor.arch), version, fileInfo.Name.String)
				if err := s.downloadPackage(update, fileInfo.Name.String, fileInfo.Hash.String, fileInfo.Hash256.String, downloadName); err != nil {
					// Clean up any extra files we already downloaded
					s.cleanupDownloadedFiles(downloadedFiles)
					l.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).
						Msgf("processExtraFiles - downloading package %s", fileInfo.Name.String)
					return nil, err
				}
				fileInfo.Name = null.StringFrom(downloadName)
				downloadedFiles = append(downloadedFiles, downloadName)
			}
		}
	}

	return extraFiles, nil
}

// downloadPackage downloads and verifies the package payload referenced in the
// update provided. The downloaded package payload is stored in packagesPath
// using the filename provided.
func (s *Syncer) downloadPackage(update *omaha.UpdateResponse, pkgName, sha1Base64Checksum, sha256Base16Checksum, filename string) error {
	tmpFile, err := os.CreateTemp(s.packagesPath, "tmp_flatcar_pkg_")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	updateURL, err := url.Parse(update.URLs[0].CodeBase)
	if err != nil {
		return err
	}

	updateURL.Path = path.Join(updateURL.Path, pkgName)

	pkgURL := updateURL.String()
	resp, err := s.httpClient.Get(pkgURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received unexpected status code (%d)", resp.StatusCode)
	}
	defer resp.Body.Close()

	hashSha1 := sha1.New()
	hashSha256 := sha256.New()
	l.Debug().Msgf("downloadPackage, downloading.. url %s", pkgURL)
	if _, err := io.Copy(io.MultiWriter(tmpFile, hashSha256, hashSha1), resp.Body); err != nil {
		return err
	}
	// Only check the checksums if provided
	if sha1Base64Checksum != "" && base64.StdEncoding.EncodeToString(hashSha1.Sum(nil)) != sha1Base64Checksum {
		return errors.New("downloaded file sha1 hash mismatch")
	}
	if sha256Base16Checksum != "" && hex.EncodeToString(hashSha256.Sum(nil)) != sha256Base16Checksum {
		return errors.New("downloaded file sha256 hash mismatch")
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpFile.Name(), filepath.Join(s.packagesPath, filename)); err != nil {
		return err
	}

	return nil
}
