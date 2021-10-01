package nebraska

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

type AppConfig codegen.CreateAppJSONBody

type CreateAppOptions struct {
	CloneFrom *string
	commonOptions
}

type GetAppOptions struct {
	commonOptions
}

type UpdateAppOptions struct {
	commonOptions
}

type DeleteAppOptions struct {
	commonOptions
}

type PaginateAppOptions struct {
	commonOptions
}

type ApplicationsClient interface {
	Create(ctx context.Context, conf AppConfig, opts *CreateAppOptions) (Application, error)
	Get(ctx context.Context, ID string, opts *GetAppOptions) (Application, error)
	Update(ctx context.Context, ID string, conf AppConfig, opts *UpdateAppOptions) (Application, error)
	Delete(ctx context.Context, ID string, opts *DeleteAppOptions) error
	Paginate(ctx context.Context, page int, perpage int, opts *PaginateAppOptions) (*Applications, error)
}

type applicationClient struct {
	config *Config
	client *codegen.ClientWithResponses
}

func (ac *applicationClient) Create(ctx context.Context, conf AppConfig, opts *CreateAppOptions) (Application, error) {
	if opts == nil {
		opts = &CreateAppOptions{}
	}

	var params codegen.CreateAppParams
	if opts.CloneFrom != nil {
		params.CloneFrom = opts.CloneFrom
	}

	resp, err := ac.client.CreateApp(ctx, &params, codegen.CreateAppJSONRequestBody(conf), convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("creating app returned: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create app returned invalid response code: %d", resp.StatusCode)
	}

	return ac.parseApplication(resp)
}

func (ac *applicationClient) Get(ctx context.Context, ID string, opts *GetAppOptions) (Application, error) {
	if opts == nil {
		opts = &GetAppOptions{}
	}

	resp, err := ac.client.GetApp(ctx, ID, convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("fetching app id %q returned:  %w", ID, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get app id %q returned invalid response code: %d", ID, resp.StatusCode)
	}
	return ac.parseApplication(resp)
}

func (ac *applicationClient) Update(ctx context.Context, ID string, conf AppConfig, opts *UpdateAppOptions) (Application, error) {
	if opts == nil {
		opts = &UpdateAppOptions{}
	}

	resp, err := ac.client.UpdateApp(ctx, ID, codegen.UpdateAppJSONRequestBody(conf), convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("updating app id %q returned %w", ID, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update app id %q returned invalid response code: %d", ID, resp.StatusCode)
	}

	return ac.parseApplication(resp)
}

func (ac *applicationClient) Delete(ctx context.Context, ID string, opts *DeleteAppOptions) error {
	if opts == nil {
		opts = &DeleteAppOptions{}
	}

	resp, err := ac.client.DeleteAppWithResponse(ctx, ID, convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return fmt.Errorf("deleting app %q returned %w", ID, err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("delete app id %q returned invalid response code: %d", ID, resp.StatusCode())
	}
	return nil
}

func (ac *applicationClient) Paginate(ctx context.Context, page int, perpage int, opts *PaginateAppOptions) (*Applications, error) {
	if opts == nil {
		opts = &PaginateAppOptions{}
	}

	var params codegen.PaginateAppsParams
	params.Page = &page
	params.Perpage = &perpage

	resp, err := ac.client.PaginateApps(ctx, &params, convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("paginate app: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paginate apps returned invalid response code: %d", resp.StatusCode)
	}
	return ac.parseApplications(resp)
}

func (ac *applicationClient) parseApplication(resp *http.Response) (Application, error) {
	var application application

	if !strings.Contains(resp.Header.Get("Content-Type"), "json") {
		return nil, fmt.Errorf("invalid application response content-type: %q", resp.Header.Get("Content-Type"))
	}

	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&application.properties)
	if err != nil {
		return nil, fmt.Errorf("application decode: %w", err)
	}
	application.appsClient = ac
	application.groupsClient = &groupsClient{
		appID:  application.properties.Id,
		config: ac.config,
		client: ac.client,
	}
	application.rawResponse = rawResponse{resp}
	return &application, nil
}

func (ac *applicationClient) parseApplications(resp *http.Response) (*Applications, error) {
	type appsPage struct {
		Applications []json.RawMessage `json:"applications"`
		Count        int               `json:"count"`
		TotalCount   int               `json:"totalCount"`
	}

	var rawApps appsPage

	if !strings.Contains(resp.Header.Get("Content-Type"), "json") {
		return nil, fmt.Errorf("invalid applications response content-type: %q", resp.Header.Get("Content-Type"))
	}

	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&rawApps)
	if err != nil {
		return nil, fmt.Errorf("app page response decode error: %w", err)
	}
	var apps []Application

	for _, rawApp := range rawApps.Applications {
		var app application
		decoder := json.NewDecoder(bytes.NewReader(rawApp))
		err := decoder.Decode(&app.properties)
		if err != nil {
			return nil, fmt.Errorf("application decoding: %w", err)
		}
		app.appsClient = ac
		app.groupsClient = &groupsClient{
			appID:  app.properties.Id,
			config: ac.config,
			client: ac.client,
		}
		// TODO: Figure out raw response for apps
		apps = append(apps, app)
	}
	return &Applications{Apps: apps, TotalCount: rawApps.TotalCount, rawResponse: rawResponse{resp}}, nil
}
