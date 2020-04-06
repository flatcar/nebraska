package syncer

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-omaha/omaha"
	"github.com/google/uuid"
	log "github.com/mgutz/logxi/v1"
	"gopkg.in/mgutz/dat.v1"

	"github.com/kinvolk/nebraska/pkg/api"
)

const validityInterval = "1 days"
const (
	flatcarUpdatesURL = "https://public.update.flatcar-linux.net/v1/update/"
	flatcarAppID      = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"
	checkFrequency    = 1 * time.Hour
)

var (
	logger = log.New("syncer")

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
	api          *api.API
	hostPackages bool
	packagesPath string
	packagesURL  string
	stopCh       chan struct{}
	machinesIDs  map[channelDescriptor]string
	bootIDs      map[channelDescriptor]string
	versions     map[channelDescriptor]string
	channelsIDs  map[channelDescriptor]string
	httpClient   *http.Client
	ticker       *time.Ticker
}

// Config represents the configuration used to create a new Syncer instance.
type Config struct {
	API          *api.API
	HostPackages bool
	PackagesPath string
	PackagesURL  string
}

// New creates a new Syncer instance.
func New(conf *Config) (*Syncer, error) {
	if conf.API == nil {
		return nil, ErrInvalidAPIInstance
	}

	s := &Syncer{
		api:          conf.API,
		hostPackages: conf.HostPackages,
		packagesPath: conf.PackagesPath,
		packagesURL:  conf.PackagesURL,
		stopCh:       make(chan struct{}),
		machinesIDs:  make(map[channelDescriptor]string, 8),
		bootIDs:      make(map[channelDescriptor]string, 8),
		channelsIDs:  make(map[channelDescriptor]string, 8),
		versions:     make(map[channelDescriptor]string, 8),
		httpClient:   &http.Client{},
	}

	if err := s.initialize(); err != nil {
		return nil, err
	}

	return s, nil
}

// Start makes the syncer start working. It will check for updates every
// checkFrequency until it's asked to stop.
func (s *Syncer) Start() {
	logger.Debug("syncer ready!")
	s.ticker = time.NewTicker(checkFrequency)

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
	logger.Debug("stopping syncer..")
	s.stopCh <- struct{}{}
}

// initialize does some initial setup to prepare the syncer, checking in
// Nebraska the last versions we know about for the different channels in the
// Flatcar application and keeping track of some ids.
func (s *Syncer) initialize() error {
	flatcarApp, err := s.api.GetApp(flatcarAppID, validityInterval)
	if err != nil {
		return err
	}

	for _, c := range flatcarApp.Channels {
		if c.Name == "stable" || c.Name == "beta" || c.Name == "alpha" || c.Name == "edge" {
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
		logger.Debug("checking for updates", "channel", descriptor.name, "arch", descriptor.arch.String(), "currentVersion", currentVersion)

		update, err := s.doOmahaRequest(descriptor, currentVersion)
		if err != nil {
			return err
		}
		if update != nil && update.Status == "ok" {
			logger.Debug("checkForUpdates, got an update", "channel", descriptor.name, "arch", descriptor.arch.String(), "currentVersion", currentVersion, "availableVersion", update.Manifest.Version)
			if err := s.processUpdate(descriptor, update); err != nil {
				return err
			}
			s.versions[descriptor] = update.Manifest.Version
			s.bootIDs[descriptor] = "{" + uuid.New().String() + "}"
		} else {
			logger.Debug("checkForUpdates, no update available", "channel", descriptor.name, "arch", descriptor.arch.String(), "currentVersion", currentVersion, "updateStatus", update.Status)
		}

		select {
		case <-time.After(1 * time.Minute):
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
		logger.Error("checkForUpdates, marshalling request xml", "error", err)
		return nil, err
	}
	logger.Debug("doOmahaRequest", "request", string(payload))

	resp, err := s.httpClient.Post(flatcarUpdatesURL, "text/xml", bytes.NewReader(payload))
	if err != nil {
		logger.Error("checkForUpdates, posting omaha response", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("checkForUpdates, reading omaha response", "error", err)
		return nil, err
	}
	logger.Debug("doOmahaRequest", "response", string(body))

	oresp := &omaha.Response{}
	err = xml.Unmarshal(body, oresp)
	if err != nil {
		logger.Error("checkForUpdates, unmarshalling omaha response", "error", err)
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

		if s.hostPackages {
			url = s.packagesURL
			filename = fmt.Sprintf("flatcar-%s-%s.gz", getArchString(descriptor.arch), update.Manifest.Version)
			if err := s.downloadPackage(update, filename); err != nil {
				logger.Error("processUpdate, downloading package", "error", err, "channel", descriptor.name, "arch", descriptor.arch.String())
				return err
			}
		}

		pkg = &api.Package{
			Type:          api.PkgTypeFlatcar,
			URL:           url,
			Version:       update.Manifest.Version,
			Filename:      dat.NullStringFrom(filename),
			Size:          dat.NullStringFrom(strconv.FormatUint(update.Manifest.Packages[0].Size, 10)),
			Hash:          dat.NullStringFrom(update.Manifest.Packages[0].SHA1),
			ApplicationID: flatcarAppID,
			Arch:          descriptor.arch,
		}
		if _, err = s.api.AddPackage(pkg); err != nil {
			logger.Error("processUpdate, adding package", "error", err, "channel", descriptor.name, "arch", descriptor.arch.String())
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
			logger.Error("processUpdate, adding flatcar action", "error", err, "channel", descriptor.name, "arch", descriptor.arch.String())
			return err
		}
	}

	// Update channel to point to the package with the new version
	channel, err := s.api.GetChannel(s.channelsIDs[descriptor])
	if err != nil {
		logger.Error("processUpdate, getting channel to update", "error", err, "channel", descriptor.name, "arch", descriptor.arch.String())
		return err
	}
	channel.PackageID = dat.NullStringFrom(pkg.ID)
	if err = s.api.UpdateChannel(channel); err != nil {
		logger.Error("processUpdate, updating channel", "error", err, "channel", descriptor.name, "arch", descriptor.arch.String())
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
func (s *Syncer) downloadPackage(update *omaha.UpdateResponse, filename string) error {
	tmpFile, err := ioutil.TempFile(s.packagesPath, "tmp_flatcar_pkg_")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	updateURL, err := url.Parse(update.URLs[0].CodeBase)
	if err != nil {
		return err
	}

	updateURL.Path = path.Join(updateURL.Path, update.Manifest.Packages[0].Name)

	pkgURL := updateURL.String()
	resp, err := http.Get(pkgURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received unexpected status code (%d)", resp.StatusCode)
	}
	defer resp.Body.Close()

	hashSha256 := sha256.New()
	logger.Debug("downloadPackage, downloading..", "url", pkgURL)
	if _, err := io.Copy(io.MultiWriter(tmpFile, hashSha256), resp.Body); err != nil {
		return err
	}
	if base64.StdEncoding.EncodeToString(hashSha256.Sum(nil)) != update.Manifest.Actions[0].SHA256 {
		return errors.New("downloaded file hash mismatch")
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpFile.Name(), filepath.Join(s.packagesPath, filename)); err != nil {
		return err
	}

	return nil
}
