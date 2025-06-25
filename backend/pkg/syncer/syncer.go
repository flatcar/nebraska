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
	"time"

	"github.com/google/uuid"
	"github.com/kinvolk/go-omaha/omaha"
	"gopkg.in/guregu/null.v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/config"
	"github.com/kinvolk/nebraska/backend/pkg/util"
)

const (
	flatcarAppID = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"
)

var (
	logger = util.NewLogger("syncer")

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
}

// Setup creates a new syncer from config and db connection, and returns it.
func Setup(conf *config.Config, db *api.API) (*Syncer, error) {
	checkFrequency, err := time.ParseDuration(conf.CheckFrequencyVal)
	if err != nil {
		return nil, fmt.Errorf("invalid Check Frequency value: %w", err)
	}

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
		httpClient:        &http.Client{},
	}

	if err := s.initialize(); err != nil {
		return nil, err
	}

	return s, nil
}

// Start makes the syncer start working. It will check for updates every
// checkFrequency until it's asked to stop.
func (s *Syncer) Start() {
	logger.Debug().Msg("syncer ready!")
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
	logger.Debug().Msg("stopping syncer..")
	s.stopCh <- struct{}{}
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
		logger.Debug().Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Str("currentVersion", currentVersion).Msg("checking for updates")

		update, err := s.doOmahaRequest(descriptor, currentVersion)
		if err != nil {
			return err
		}
		if update != nil && update.Status == "ok" {
			logger.Debug().Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Str("currentVersion", currentVersion).Str("availableVersion", update.Manifest.Version).Send()
			if err := s.processUpdate(descriptor, update); err != nil {
				return err
			}
			s.versions[descriptor] = update.Manifest.Version
			s.bootIDs[descriptor] = "{" + uuid.New().String() + "}"
		} else {
			logger.Debug().Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Str("currentVersion", currentVersion).Msgf("checkForUpdates, no update available updateStatus %v", update.Status)
		}

		select {
		case <-time.After(5 * time.Second):
		case <-s.stopCh:
			break
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

	payload, err := xml.Marshal(req)
	if err != nil {
		logger.Error().Err(err).Msg("checkForUpdates, marshalling request xml")
		return nil, err
	}
	logger.Debug().Str("request", string(payload)).Msg("doOmahaRequest")

	resp, err := s.httpClient.Post(s.flatcarUpdatesURL, "text/xml", bytes.NewReader(payload))
	if err != nil {
		logger.Error().Err(err).Msg("checkForUpdates, posting omaha response")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error().Err(err).Msg("checkForUpdates, reading omaha response")
		return nil, err
	}
	logger.Debug().Str("response", string(body)).Msg("doOmahaRequest")

	oresp := &omaha.Response{}
	err = xml.Unmarshal(body, oresp)
	if err != nil {
		logger.Error().Err(err).Msg("checkForUpdates, unmarshalling omaha response")
		return nil, err
	}

	return oresp.Apps[0].UpdateCheck, nil
}

// processUpdate is in charge of creating packages in the Flatcar application in
// Nebraska and updating the appropriate channel to point to the new channel.
func (s *Syncer) processUpdate(descriptor channelDescriptor, update *omaha.UpdateResponse) error {
	// Create new package and action for Flatcar application in Nebraska if
	// needed (package may already exist and we just need to update the channel
	// reference to it)
	pkg, err := s.api.GetPackageByVersionAndArch(flatcarAppID, update.Manifest.Version, descriptor.arch)
	if err != nil {
		url := update.URLs[0].CodeBase
		filename := update.Manifest.Packages[0].Name

		// Allow to override the URL if needed.
		if s.packagesURL != "" {
			url = strings.ReplaceAll(s.packagesURL, "{{VERSION}}", update.Manifest.Version)
			url = strings.ReplaceAll(url, "{{ARCH}}", getArchString(descriptor.arch))
		}

		var extraFiles []api.File
		if len(update.Manifest.Packages) > 1 {
			extraFiles = make([]api.File, len(update.Manifest.Packages)-1)
			for i := 1; i < len(update.Manifest.Packages); i++ {
				omahaPkg := update.Manifest.Packages[i]
				size := strconv.FormatUint(omahaPkg.Size, 10)
				extraFiles[i-1] = api.File{
					Name:    null.StringFrom(omahaPkg.Name),
					Size:    null.StringFrom(size),
					Hash:    null.StringFrom(omahaPkg.SHA1),
					Hash256: null.StringFrom(omahaPkg.SHA256),
				}
			}
		}

		if s.hostPackages {
			filename = fmt.Sprintf("flatcar-%s-%s.gz", getArchString(descriptor.arch), update.Manifest.Version)
			base16sha256 := ""
			if update.Manifest.Actions[0].SHA256 != "" {
				binsha256, err := base64.StdEncoding.DecodeString(update.Manifest.Actions[0].SHA256)
				if err != nil {
					logger.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Msg("processUpdate, converting sha256")
					return err
				}
				base16sha256 = hex.EncodeToString(binsha256)
			}
			if err := s.downloadPackage(update, update.Manifest.Packages[0].Name, update.Manifest.Packages[0].SHA1, base16sha256, filename); err != nil {
				logger.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Msg("processUpdate, downloading package")
				return err
			}

			for i := range extraFiles {
				fileInfo := &extraFiles[i]
				downloadName := fmt.Sprintf("extrafile-%s-%s-%s", getArchString(descriptor.arch), update.Manifest.Version, fileInfo.Name.String)
				if err := s.downloadPackage(update, fileInfo.Name.String, fileInfo.Hash.String, fileInfo.Hash256.String, downloadName); err != nil {
					logger.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Msgf("processUpdate, downloading package %s", fileInfo.Name.String)
					return err
				}
				fileInfo.Name = null.StringFrom(downloadName)
			}
		}

		pkg = &api.Package{
			Type:          api.PkgTypeFlatcar,
			URL:           url,
			Version:       update.Manifest.Version,
			Filename:      null.StringFrom(filename),
			Size:          null.StringFrom(strconv.FormatUint(update.Manifest.Packages[0].Size, 10)),
			Hash:          null.StringFrom(update.Manifest.Packages[0].SHA1),
			ApplicationID: flatcarAppID,
			Arch:          descriptor.arch,
			ExtraFiles:    extraFiles,
		}
		if _, err = s.api.AddPackage(pkg); err != nil {
			logger.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Msg("processUpdate, adding package")
			return err
		}

		flatcarAction := &api.FlatcarAction{
			Event:                 update.Manifest.Actions[0].Event,
			ChromeOSVersion:       update.Manifest.Actions[0].DisplayVersion,
			Sha256:                update.Manifest.Actions[0].SHA256,
			NeedsAdmin:            update.Manifest.Actions[0].NeedsAdmin,
			IsDelta:               update.Manifest.Actions[0].IsDeltaPayload,
			DisablePayloadBackoff: update.Manifest.Actions[0].DisablePayloadBackoff,
			MetadataSignatureRsa:  update.Manifest.Actions[0].MetadataSignatureRsa,
			MetadataSize:          update.Manifest.Actions[0].MetadataSize,
			Deadline:              update.Manifest.Actions[0].Deadline,
			PackageID:             pkg.ID,
		}
		if _, err = s.api.AddFlatcarAction(flatcarAction); err != nil {
			logger.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Msgf("processUpdate, adding flatcar action")
			return err
		}
	}

	// Update channel to point to the package with the new version
	channel, err := s.api.GetChannel(s.channelsIDs[descriptor])
	if err != nil {
		logger.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Msg("processUpdate, getting channel to update")
		return err
	}
	channel.PackageID = null.StringFrom(pkg.ID)
	if err = s.api.UpdateChannel(channel); err != nil {
		logger.Error().Err(err).Str("channel", descriptor.name).Str("arch", descriptor.arch.String()).Msg("processUpdate, updating")
		return err
	}

	return nil
}

func getArchString(arch api.Arch) string {
	return strings.TrimSuffix(arch.CoreosString(), "-usr")
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
	resp, err := http.Get(pkgURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received unexpected status code (%d)", resp.StatusCode)
	}
	defer resp.Body.Close()

	hashSha1 := sha1.New()
	hashSha256 := sha256.New()
	logger.Debug().Msgf("downloadPackage, downloading.. url %s", pkgURL)
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
