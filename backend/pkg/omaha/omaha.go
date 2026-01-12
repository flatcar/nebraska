package omaha

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"

	omahaSpec "github.com/flatcar/go-omaha/omaha"
	"github.com/rs/zerolog"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

var (
	l = logger.New("omaha")

	initialFlatcarGroups = map[string]string{
		// amd64
		"5b810680-e36a-4879-b98a-4f989e80b899": "alpha",
		"3fe10490-dd73-4b49-b72a-28ac19acfcdc": "beta",
		"9a2deb70-37be-4026-853f-bfdd6b347bbe": "stable",
		"72834a2b-ad86-4d6d-b498-e08a19ebe54e": "edge",
		// arm64
		"e641708d-fb48-4260-8bdf-ba2074a1147a": "alpha",
		"d112ec01-ba34-4a9e-9d4b-9814a685f266": "beta",
		"11a585f6-9418-4df0-8863-78b2fd3240f8": "stable",
		"b4b2fa22-c1ea-498c-a8ac-c1dc0b1d7c17": "edge",
	}

	// ErrMalformedRequest error indicates that the omaha request it has
	// received is malformed.
	ErrMalformedRequest = errors.New("omaha: request is malformed")

	// ErrMalformedResponse error indicates that the omaha response it wants to
	// send is malformed.
	ErrMalformedResponse = errors.New("omaha: response is malformed")
)

// Handler represents a component capable of processing Omaha requests. It uses
// the Nebraska API to get packages updates, process events, etc.
type Handler struct {
	crAPI *api.API
}

// NewHandler creates a new Handler instance.
func NewHandler(crAPI *api.API) *Handler {
	return &Handler{
		crAPI: crAPI,
	}
}

// Handle is in charge of processing an Omaha request.
func (h *Handler) Handle(rawReq io.Reader, respWriter io.Writer, ip string) error {
	var omahaReq *omahaSpec.Request

	if err := xml.NewDecoder(rawReq).Decode(&omahaReq); err != nil {
		l.Warn().Msgf("Handle - malformed omaha request error %s", err.Error())
		return fmt.Errorf("%s: %w", ErrMalformedRequest, err)
	}
	trace(omahaReq)

	omahaResp, err := h.buildOmahaResponse(omahaReq, ip)
	if err != nil {
		l.Warn().Msgf("Handle - error building omaha response error %s", err.Error())
		return ErrMalformedResponse
	}
	trace(omahaResp)

	return xml.NewEncoder(respWriter).Encode(omahaResp)
}

func getArch(os *omahaSpec.OS, appReq *omahaSpec.AppRequest) api.Arch {
	if appReq != nil {
		if arch, err := api.ArchFromCoreosString(appReq.Board); err == nil {
			return arch
		}
	}

	if os != nil {
		if arch, err := api.ArchFromOmahaString(os.Arch); err == nil {
			return arch
		}
		if arch, err := api.ArchFromString(os.Arch); err == nil {
			return arch
		}
	}
	l.Debug().Msgf("getArch - unknown arch, assuming amd64 arch")
	return api.ArchAMD64
}

// isSyncerClient detects if the client is a Nebraska syncer based on request characteristics
func isSyncerClient(req *omahaSpec.Request) bool {
	// Nebraska syncers must have BOTH:
	// 1. The exact hardcoded request version "CoreOSUpdateEngine-0.1.0.0"
	// 2. InstallSource set to "scheduler"
	// Regular Flatcar clients use version strings like "update_engine-0.4.2" and different install sources
	if req.InstallSource == "scheduler" &&
		req.Version == "CoreOSUpdateEngine-0.1.0.0" {
		return true
	}
	return false
}

func (h *Handler) buildOmahaResponse(omahaReq *omahaSpec.Request, ip string) (*omahaSpec.Response, error) {
	omahaResp := omahaSpec.NewResponse()
	omahaResp.Server = "nebraska"

	for _, reqApp := range omahaReq.Apps {
		var respApp *omahaSpec.AppResponse

		appID, err := h.crAPI.GetAppID(reqApp.ID)
		if err != nil {
			l.Info().Str("machineId", reqApp.MachineID).Str("app", reqApp.ID).Msgf("buildOmahaResponse - no app found for %s", err.Error())

			respApp = omahaResp.AddApp(reqApp.ID, omahaSpec.AppUnknownID)
			respApp.Status = h.getStatusMessage(err)
			respApp.AddUpdateCheck(omahaSpec.UpdateInternalError)

			return omahaResp, nil
		}

		respApp = omahaResp.AddApp(reqApp.ID, omahaSpec.AppOK)
		// Use the Omaha track field to find the group. It preferably contains the group's track name
		// but also allows the old hard-coded CoreOS group UUIDs until we now that they are not used.
		group := reqApp.Track
		if trackName, ok := initialFlatcarGroups[group]; ok {
			l.Info().Str("machineId", reqApp.MachineID).Str("uuid", group).Msgf("buildOmahaResponse - found client using a hard-coded group UUID")
			group = trackName
		}
		groupID, err := h.crAPI.GetGroupID(appID, group, getArch(omahaReq.OS, reqApp))
		if err == nil {
			group = groupID
		} else {
			l.Info().Str("machineId", reqApp.MachineID).Str("track", group).Msgf("buildOmahaResponse - no group found for track and arch error %s", err.Error())
			respApp.Status = h.getStatusMessage(err)
			respApp.AddUpdateCheck(omahaSpec.UpdateInternalError)
			return omahaResp, nil
		}

		for _, event := range reqApp.Events {
			if err := h.processEvent(reqApp.MachineID, appID, group, event); err != nil {
				l.Debug().Str("machineId", reqApp.MachineID).Msgf("processEvent error %s", err.Error())
			}
			respApp.AddEvent()
		}

		if reqApp.Ping != nil {
			if _, err := h.crAPI.RegisterInstance(reqApp.MachineID, reqApp.MachineAlias, ip, reqApp.Version, appID, group, reqApp.OEM, reqApp.OEMVersion); err != nil {
				l.Debug().Str("machineId", reqApp.MachineID).Msgf("processPing error %s", err.Error())
			}
			respApp.AddPing()
		}

		if reqApp.UpdateCheck != nil {
			if isSyncerClient(omahaReq) {
				// Syncer - get floors and target separately
				// target is nil when more floors remain beyond NEBRASKA_MAX_FLOORS_PER_RESPONSE limit
				floors, target, err := h.crAPI.GetUpdatePackagesForSyncer(reqApp.MachineID, reqApp.MachineAlias, ip, reqApp.Version, appID, group, reqApp.OEM, reqApp.OEMVersion)
				if err != nil {
					if err == api.ErrNoUpdatePackageAvailable || err == api.ErrUpdateGrantFailed {
						respApp.AddUpdateCheck(omahaSpec.NoUpdate)
					} else {
						respApp.Status = h.getStatusMessage(err)
						respApp.AddUpdateCheck(omahaSpec.UpdateInternalError)
					}
					continue
				}

				// Check if we got anything to send
				if len(floors) == 0 && target == nil {
					respApp.AddUpdateCheck(omahaSpec.NoUpdate)
					continue
				}

				// Critical safety rule: old syncers without MultiManifestOK cannot skip floors
				if len(floors) > 0 && !reqApp.MultiManifestOK {
					l.Warn().Str("instanceID", reqApp.MachineID).
						Int("floorCount", len(floors)).
						Msg("Syncer without multi-manifest support blocked due to floor requirements")
					respApp.AddUpdateCheck(omahaSpec.NoUpdate)
					continue
				}

				// Either multi-manifest capable syncer or no floors exist
				h.prepareMultiManifestUpdateCheck(respApp, floors, target)
			} else {
				// Regular client - get single package
				pkg, err := h.crAPI.GetUpdatePackage(reqApp.MachineID, reqApp.MachineAlias, ip, reqApp.Version, appID, group, reqApp.OEM, reqApp.OEMVersion)
				if err != nil {
					if err == api.ErrNoUpdatePackageAvailable || err == api.ErrUpdateGrantFailed {
						respApp.AddUpdateCheck(omahaSpec.NoUpdate)
					} else {
						respApp.Status = h.getStatusMessage(err)
						respApp.AddUpdateCheck(omahaSpec.UpdateInternalError)
					}
					continue
				}

				// Single package response
				h.prepareUpdateCheck(respApp, pkg)
			}
		}
	}

	return omahaResp, nil
}

func (h *Handler) processEvent(machineID string, appID string, group string, event *omahaSpec.EventRequest) error {
	l.Info().Str("machineId", machineID).Str("appID", appID).Str("group", group).Str("event", event.Type.String()+"."+event.Result.String()).Str("previousVersion", event.PreviousVersion).Msgf("processEvent eventError %d", event.ErrorCode)

	return h.crAPI.RegisterEvent(machineID, appID, group, int(event.Type), int(event.Result), event.PreviousVersion, strconv.Itoa(event.ErrorCode))
}

func (h *Handler) getStatusMessage(crErr error) omahaSpec.AppStatus {
	return omahaSpec.AppStatus(h.getStatusMessageStr(crErr))
}

// TODO(krnowak): This seems to return a bunch of custom errors. Not
// sure if we should try to match it to the standard or extra
// AppStatus constants.
func (h *Handler) getStatusMessageStr(crErr error) string {
	switch crErr {
	case api.ErrNoPackageFound:
		return "error-noPackageFound"
	case api.ErrInvalidApplicationOrGroup:
		return "error-unknownApplicationOrGroup"
	case api.ErrRegisterInstanceFailed:
		return "error-instanceRegistrationFailed"
	case api.ErrMaxUpdatesPerPeriodLimitReached:
		return "error-maxUpdatesPerPeriodLimitReached"
	case api.ErrMaxConcurrentUpdatesLimitReached:
		return "error-maxConcurrentUpdatesLimitReached"
	case api.ErrMaxTimedOutUpdatesLimitReached:
		return "error-maxTimedOutUpdatesLimitReached"
	case api.ErrUpdatesDisabled:
		return "error-updatesDisabled"
	case api.ErrGetUpdatesStatsFailed:
		return "error-couldNotCheckUpdatesStats"
	case api.ErrUpdateInProgressOnInstance:
		return "error-updateInProgressOnInstance"
	}

	l.Warn().Msgf("getStatusMessage error %s", crErr.Error())

	return "error-failedToRetrieveUpdatePackageInfo"
}

// addPackageToManifest adds a package and its extra files to the manifest
func (h *Handler) addPackageToManifest(manifest *omahaSpec.Manifest, pkg *api.Package) {
	mpkg := manifest.AddPackage()
	mpkg.Name = pkg.Filename.String
	mpkg.SHA1 = pkg.Hash.String
	if pkg.Size.Valid {
		size, err := strconv.ParseUint(pkg.Size.String, 10, 64)
		if err != nil {
			l.Warn().Msgf("addPackageToManifest bad package size %s", err.Error())
		} else {
			mpkg.Size = size
		}
	}
	mpkg.Required = true

	// Add extra files
	for _, pkgFile := range pkg.ExtraFiles {
		fpkg := manifest.AddPackage()
		fpkg.Name = pkgFile.Name.String
		if pkgFile.Hash.Valid {
			fpkg.SHA1 = pkgFile.Hash.String
			fpkg.SHA256 = pkgFile.Hash256.String
		}
		if pkgFile.Size.Valid && pkgFile.Size.String != "" {
			size, err := strconv.ParseUint(pkgFile.Size.String, 10, 64)
			if err != nil {
				l.Warn().Msgf("addPackageToManifest bad size %s for extra file %s", err.Error(), pkgFile.Name.String)
			} else {
				fpkg.Size = size
			}
		}
	}
}

// addFlatcarActionToManifest adds Flatcar-specific action to manifest if applicable
func (h *Handler) addFlatcarActionToManifest(manifest *omahaSpec.Manifest, pkg *api.Package) error {
	if pkg.Type != api.PkgTypeFlatcar {
		return nil
	}

	cra, err := h.crAPI.GetFlatcarAction(pkg.ID)
	if err != nil {
		return err
	}

	a := manifest.AddAction(cra.Event)
	a.DisplayVersion = cra.ChromeOSVersion
	a.SHA256 = cra.Sha256
	a.NeedsAdmin = cra.NeedsAdmin
	a.IsDeltaPayload = cra.IsDelta
	a.DisablePayloadBackoff = cra.DisablePayloadBackoff
	a.MetadataSignatureRsa = cra.MetadataSignatureRsa
	a.MetadataSize = cra.MetadataSize
	a.Deadline = cra.Deadline

	return nil
}

func (h *Handler) prepareUpdateCheck(appResp *omahaSpec.AppResponse, pkg *api.Package) {
	if pkg == nil {
		appResp.AddUpdateCheck(omahaSpec.NoUpdate)
		return
	}

	manifest := &omahaSpec.Manifest{Version: pkg.Version}

	// Add package and its files
	h.addPackageToManifest(manifest, pkg)

	// Add Flatcar action if applicable
	if err := h.addFlatcarActionToManifest(manifest, pkg); err != nil {
		appResp.AddUpdateCheck(omahaSpec.UpdateInternalError)
		return
	}

	updateCheck := appResp.AddUpdateCheck(omahaSpec.UpdateOK)
	updateCheck.AddManifest(manifest.Version)
	updateCheck.Manifests[0] = manifest
	updateCheck.AddURL(pkg.URL)
}

// prepareMultiManifestUpdateCheck creates a response with multiple manifests.
// When target is nil, more floors remain and syncer should request again.
func (h *Handler) prepareMultiManifestUpdateCheck(appResp *omahaSpec.AppResponse, floors []*api.Package, target *api.Package) {
	if len(floors) == 0 && target == nil {
		appResp.AddUpdateCheck(omahaSpec.NoUpdate)
		return
	}

	targetAlreadyInFloors := target != nil && len(floors) > 0 && floors[len(floors)-1].ID == target.ID

	var packages []*api.Package
	packages = append(packages, floors...)
	if target != nil && !targetAlreadyInFloors {
		packages = append(packages, target)
	}

	var codeBase string
	if target != nil {
		codeBase = target.URL
	} else {
		codeBase = packages[len(packages)-1].URL
	}

	updateCheck := appResp.AddUpdateCheck(omahaSpec.UpdateOK)
	updateCheck.AddURL(codeBase)

	for _, pkg := range packages {
		manifest := updateCheck.AddManifest(pkg.Version)

		if pkg.IsFloor {
			manifest.IsFloor = true
			if pkg.FloorReason.Valid {
				manifest.FloorReason = pkg.FloorReason.String
			} else {
				manifest.FloorReason = "Required intermediate version"
			}
		}

		isTarget := target != nil && pkg.ID == target.ID
		if isTarget {
			manifest.IsTarget = true
		}

		h.addPackageToManifest(manifest, pkg)
		if err := h.addFlatcarActionToManifest(manifest, pkg); err != nil {
			appResp.UpdateCheck.Status = omahaSpec.UpdateInternalError
			return
		}
	}
}

func trace(v interface{}) {
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		raw, err := xml.MarshalIndent(v, "", " ")
		if err != nil {
			l.Error().Err(err).Msg("")
			return
		}
		l.Debug().Str("XML", string(raw)).Msg("Omaha trace")
	}
}
