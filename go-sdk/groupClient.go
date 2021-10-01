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

type GroupConfig codegen.CreateGroupJSONBody

type CreateGroupOptions struct {
	commonOptions
}
type GetGroupOptions struct {
	commonOptions
}
type UpdateGroupOptions struct {
	commonOptions
}
type DeleteGroupOptions struct {
	commonOptions
}
type PaginateGroupOptions struct {
	commonOptions
}

type GroupsClient interface {
	Create(ctx context.Context, conf GroupConfig, opts *CreateGroupOptions) (Group, error)
	Get(ctx context.Context, groupID string, opts *GetGroupOptions) (Group, error)
	Update(ctx context.Context, groupID string, conf GroupConfig, opts *UpdateGroupOptions) (Group, error)
	Delete(ctx context.Context, groupID string, opts *DeleteGroupOptions) error
	Paginate(ctx context.Context, page int, perpage int, opts *PaginateGroupOptions) (*Groups, error)
}

type groupsClient struct {
	appID  string
	config *Config
	client *codegen.ClientWithResponses
}

func (gc *groupsClient) Create(ctx context.Context, conf GroupConfig, opts *CreateGroupOptions) (Group, error) {
	if opts == nil {
		opts = &CreateGroupOptions{}
	}

	resp, err := gc.client.CreateGroup(ctx, gc.appID, codegen.CreateGroupJSONRequestBody(conf), convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("creating group app id %q returned %w", gc.appID, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create group for app id %q returned invalid response code: %d", gc.appID, resp.StatusCode)
	}

	return gc.parseGroup(resp)
}

func (gc *groupsClient) Get(ctx context.Context, groupID string, opts *GetGroupOptions) (Group, error) {
	if opts == nil {
		opts = &GetGroupOptions{}
	}

	resp, err := gc.client.GetGroup(ctx, gc.appID, groupID, convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("get group with app id: %q, group id: %q returned: %w", gc.appID, groupID, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get group with app id: %q, group id: %q returned invalid response code: %d", gc.appID, groupID, resp.StatusCode)
	}

	return gc.parseGroup(resp)
}

func (gc *groupsClient) Update(ctx context.Context, groupID string, conf GroupConfig, opts *UpdateGroupOptions) (Group, error) {
	if opts == nil {
		opts = &UpdateGroupOptions{}
	}

	resp, err := gc.client.UpdateGroup(ctx, gc.appID, groupID, codegen.UpdateGroupJSONRequestBody(conf), convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("update app id: %q, group id: %q returned: %w", gc.appID, groupID, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update app id: %q, group id: %q returned invalid response code %d", gc.appID, groupID, resp.StatusCode)
	}

	return gc.parseGroup(resp)
}

func (gc *groupsClient) Delete(ctx context.Context, groupID string, opts *DeleteGroupOptions) error {
	if opts == nil {
		opts = &DeleteGroupOptions{}
	}

	resp, err := gc.client.DeleteGroup(ctx, gc.appID, groupID, convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return fmt.Errorf("deleting group groupId: %q, appId: %q returned: %w", groupID, gc.appID, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete group with groupId: %q, appId: %q, returned invalid status code: %d", groupID, gc.appID, resp.StatusCode)
	}
	return nil
}

func (gc *groupsClient) Paginate(ctx context.Context, page int, perpage int, opts *PaginateGroupOptions) (*Groups, error) {
	if opts == nil {
		opts = &PaginateGroupOptions{}
	}

	var params codegen.PaginateGroupsParams
	params.Page = &page
	params.Perpage = &perpage

	resp, err := gc.client.PaginateGroups(ctx, gc.appID, &params, convertReqEditors(opts.RequestEditors...)...)
	if err != nil {
		return nil, fmt.Errorf("paginate groups, appId: %q returned: %w", gc.appID, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paginate group, appId: %q returned invalid response code: %d", gc.appID, resp.StatusCode)
	}
	return gc.parseGroups(resp)
}

func (gc *groupsClient) parseGroup(resp *http.Response) (Group, error) {
	var group group

	if !strings.Contains(resp.Header.Get("Content-Type"), "json") {
		return nil, fmt.Errorf("invalid group response content-type: %q", resp.Header.Get("Content-Type"))
	}

	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&group.properties)
	if err != nil {
		return nil, fmt.Errorf("group decode: %w", err)
	}
	group.groupClient = gc
	// group.channelClient = gc.channelClient
	group.rawResponse = rawResponse{resp}

	return group, nil
}

func (gc *groupsClient) parseGroups(resp *http.Response) (*Groups, error) {
	type groupsPage struct {
		Groups     []json.RawMessage `json:"groups"`
		Count      int               `json:"count"`
		TotalCount int               `json:"totalCount"`
	}

	var rawGroups groupsPage

	if !strings.Contains(resp.Header.Get("Content-Type"), "json") {
		return nil, fmt.Errorf("invalid groups response content-type: %q", resp.Header.Get("Content-Type"))
	}

	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&rawGroups)
	if err != nil {
		return nil, fmt.Errorf("groups page response decode: %w", err)
	}
	var groups []Group

	for _, rawGroup := range rawGroups.Groups {
		var group group
		decoder := json.NewDecoder(bytes.NewReader(rawGroup))
		err := decoder.Decode(&group.properties)
		if err != nil {
			return nil, fmt.Errorf("group decoding: %w", err)
		}
		group.groupClient = gc
		// group.channelClient = ac.channelClient
		// TODO: Figure out raw response for groups
		groups = append(groups, group)
	}
	return &Groups{Groups: groups, TotalCount: rawGroups.TotalCount, rawResponse: rawResponse{resp}}, nil
}
