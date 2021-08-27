package omaha

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"

	omahaSpec "github.com/kinvolk/go-omaha/omaha"
	"github.com/rs/zerolog"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/util"
)

var (
	logger = util.NewLogger("omaha")

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
		logger.Warn().Msgf("Handle - malformed omaha request error %s", err.Error())
		return fmt.Errorf("%s: %w", ErrMalformedRequest, err)
	}
	trace(omahaReq)

	omahaResp, err := h.buildOmahaResponse(omahaReq, ip)
	if err != nil {
		logger.Warn().Msgf("Handle - error building omaha response error %s", err.Error())
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
	logger.Debug().Msg("getArch - unknown arch, assuming amd64 arch")
	return api.ArchAMD64
}

// We want to use extra fields or attributes to what omahaSpec.Response uses.
type ResponseExtra struct {
	omahaSpec.Response
	Apps []*AppResponseExtra `xml:"app"`
}
type AppResponseExtra struct {
	omahaSpec.AppResponse
	UpdateCheck *UpdateResponseExtra `xml:"updatecheck"`
}
type UpdateResponseExtra struct {
	omahaSpec.UpdateResponse
	Payload Payload
}

// <_payload>
//   <type>application/container-image</type>
//   <registryRef>gchr.io/kinvolk/headlamp</registryRef>
//   <policy>always-pull</policy>
// </_payload>

type Payload struct {
	XMLName  	  xml.Name       `xml:"_payload" json:"-"`
	Type     	  string         `xml:"type"`
	RegistryRef   string         `xml:"registryRef"`
	Policy     	  string         `xml:"policy"`
}
	// v := &Payload{
	// 	Type: "application/container-image",
	// 	RegistryRef: "gchr.io/kinvolk/headlamp",
	// 	Policy: "always-pull",
	// }

func (r *ResponseExtra) AddApp(id string, status omahaSpec.AppStatus) *AppResponseExtra {
	a := &AppResponseExtra{AppResponse: omahaSpec.AppResponse{ID: id, Status: status}}
	r.Apps = append(r.Apps, a)
	return a
}

func (a *AppResponseExtra) AddUpdateCheck(status omahaSpec.UpdateStatus) *UpdateResponseExtra {
	a.UpdateCheck = &UpdateResponseExtra{UpdateResponse: omahaSpec.UpdateResponse{Status: status}}
	return a.UpdateCheck
}

func (h *Handler) buildOmahaResponse(omahaReq *omahaSpec.Request, ip string) (*ResponseExtra, error) {
	omahaResp := &ResponseExtra{Response: *omahaSpec.NewResponse()}
	omahaResp.Server = "nebraska"

	for _, reqApp := range omahaReq.Apps {
		respApp := omahaResp.AddApp(reqApp.ID, omahaSpec.AppOK)

		// Use the Omaha track field to find the group. It preferably contains the group's track name
		// but also allows the old hard-coded CoreOS group UUIDs until we now that they are not used.
		group := reqApp.Track
		if trackName, ok := initialFlatcarGroups[group]; ok {
			logger.Info().Str("machineId", reqApp.MachineID).Str("uuid", group).Msgf("buildOmahaResponse - found client using a hard-coded group UUID")
			group = trackName
		}
		groupID, err := h.crAPI.GetGroupID(respApp.ID, group, getArch(omahaReq.OS, reqApp))
		if err == nil {
			group = groupID
		} else {
			logger.Info().Str("machineId", reqApp.MachineID).Str("track", group).Msgf("buildOmahaResponse - no group found for track and arch error %s", err.Error())
			respApp.Status = h.getStatusMessage(err)
			respApp.AddUpdateCheck(omahaSpec.UpdateInternalError)
			return omahaResp, nil
		}

		for _, event := range reqApp.Events {
			if err := h.processEvent(reqApp.MachineID, reqApp.ID, group, event); err != nil {
				logger.Debug().Str("machineId", reqApp.MachineID).Msgf("processEvent error %s", err.Error())
			}
			respApp.AddEvent()
		}

		if reqApp.Ping != nil {
			if _, err := h.crAPI.RegisterInstance(reqApp.MachineID, reqApp.MachineAlias, ip, reqApp.Version, reqApp.ID, group); err != nil {
				logger.Debug().Str("machineId", reqApp.MachineID).Msgf("processPing error %s", err.Error())
			}
			respApp.AddPing()
		}

		if reqApp.UpdateCheck != nil {
			pkg, err := h.crAPI.GetUpdatePackage(reqApp.MachineID, reqApp.MachineAlias, ip, reqApp.Version, reqApp.ID, group)
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
	logger.Info().Str("machineId", machineID).Str("appID", appID).Str("group", group).Str("event", event.Type.String()+"."+event.Result.String()).Str("previousVersion", event.PreviousVersion).Msgf("processEvent eventError %d", event.ErrorCode)

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

	logger.Warn().Msgf("getStatusMessage error %s", crErr.Error())

	return "error-failedToRetrieveUpdatePackageInfo"
}

// prepareUpdateCheck adds an update check response to the appResp for the given pkg.
// appResp is modified
func (h *Handler) prepareUpdateCheck(appResp *AppResponseExtra, pkg *api.Package) {
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
			logger.Warn().Msgf("prepareUpdateCheck bad package size %s", err.Error())
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
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		raw, err := xml.MarshalIndent(v, "", " ")
		if err != nil {
			logger.Error().Err(err).Msg("")
			return
		}
		logger.Debug().Str("XML", string(raw)).Msg("Omaha trace")
	}
}
