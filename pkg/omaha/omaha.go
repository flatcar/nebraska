package omaha

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"

	omahaSpec "github.com/coreos/go-omaha/omaha"
	"github.com/google/uuid"
	log "github.com/mgutz/logxi/v1"

	"github.com/kinvolk/nebraska/pkg/api"
)

var (
	logger = log.New("omaha")

	flatcarAppID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"

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
		logger.Warn("Handle - malformed omaha request", "error", err.Error())
		return fmt.Errorf("%s: %w", ErrMalformedRequest, err)
	}
	trace(omahaReq)

	omahaResp, err := h.buildOmahaResponse(omahaReq, ip)
	if err != nil {
		logger.Warn("Handle - error building omaha response", "error", err.Error())
		return ErrMalformedResponse
	}
	trace(omahaResp)

	return xml.NewEncoder(respWriter).Encode(omahaResp)
}

func getArch(os *omahaSpec.OS, appReq *omahaSpec.AppRequest) api.Arch {
	arch, err := api.ArchFromCoreosString(appReq.Board)
	if err == nil {
		return arch
	}
	if os != nil {
		arch, err = api.ArchFromOmahaString(os.Arch)
		if err == nil {
			return arch
		}
	}
	logger.Warn("getArch - unknown arch, assuming amd64 arch")
	return api.ArchAMD64
}

func (h *Handler) buildOmahaResponse(omahaReq *omahaSpec.Request, ip string) (*omahaSpec.Response, error) {
	omahaResp := omahaSpec.NewResponse()
	omahaResp.Server = "nebraska"

	for _, reqApp := range omahaReq.Apps {
		respApp := omahaResp.AddApp(reqApp.ID, omahaSpec.AppOK)

		// Use Track field as the group to ask CR for updates. For the Flatcar
		// app, map group name to its id if available.
		group := reqApp.Track
		if reqAppUUID, err := uuid.Parse(reqApp.ID); err == nil {
			if reqAppUUID.String() == flatcarAppID {
				descriptor := api.ChannelDescriptor{Name: group, Arch: getArch(omahaReq.OS, reqApp)}
				if flatcarGroupID, ok := h.crAPI.LookupGroupFromTrack(descriptor); ok {
					group = flatcarGroupID
				}
			}
		}

		for _, event := range reqApp.Events {
			if err := h.processEvent(reqApp.MachineID, reqApp.ID, group, event); err != nil {
				logger.Warn("processEvent", "error", err.Error())
			}
			respApp.AddEvent()
		}

		if reqApp.Ping != nil {
			if _, err := h.crAPI.RegisterInstance(reqApp.MachineID, ip, reqApp.Version, reqApp.ID, group); err != nil {
				logger.Warn("processPing", "error", err.Error())
			}
			respApp.AddPing()
		}

		if reqApp.UpdateCheck != nil {
			pkg, err := h.crAPI.GetUpdatePackage(reqApp.MachineID, ip, reqApp.Version, reqApp.ID, group)
			if err != nil && err != api.ErrNoUpdatePackageAvailable {
				respApp.Status = h.getStatusMessage(err)
				respApp.AddUpdateCheck(omahaSpec.UpdateInternalError)
			} else {
				h.prepareUpdateCheck(respApp, pkg)
			}
		}
	}

	return omahaResp, nil
}

func (h *Handler) processEvent(machineID string, appID string, group string, event *omahaSpec.EventRequest) error {
	logger.Info("processEvent", "appID", appID, "group", group, "event", event.Type.String()+"."+event.Result.String(), "eventError", event.ErrorCode, "previousVersion", event.PreviousVersion)

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

	logger.Warn("getStatusMessage", "error", crErr.Error())

	return "error-failedToRetrieveUpdatePackageInfo"
}

func (h *Handler) prepareUpdateCheck(appResp *omahaSpec.AppResponse, pkg *api.Package) {
	if pkg == nil {
		appResp.AddUpdateCheck(omahaSpec.NoUpdate)
		return
	}

	// Create a manifest, but do not add it to UpdateCheck until it's successful
	manifest := &omahaSpec.Manifest{Version: pkg.Version}
	mpkg := manifest.AddPackage()
	mpkg.Name = pkg.Filename.String
	mpkg.SHA1 = pkg.Hash.String
	if pkg.Size.Valid {
		size, err := strconv.ParseUint(pkg.Size.String, 10, 64)
		if err != nil {
			logger.Warn("prepareUpdateCheck", "bad package size", err.Error())
		} else {
			mpkg.Size = size
		}
	}
	mpkg.Required = true

	switch pkg.Type {
	case api.PkgTypeFlatcar:
		cra, err := h.crAPI.GetFlatcarAction(pkg.ID)
		if err != nil {
			appResp.AddUpdateCheck(omahaSpec.UpdateInternalError)
			return
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
	}

	updateCheck := appResp.AddUpdateCheck(omahaSpec.UpdateOK)
	updateCheck.Manifest = manifest
	updateCheck.AddURL(pkg.URL)
}

func trace(v interface{}) {
	if logger.IsDebug() {
		raw, err := xml.MarshalIndent(v, "", " ")
		if err != nil {
			_ = logger.Error(err.Error())
			return
		}
		logger.Debug("Omaha trace", "XML", string(raw))
	}
}
