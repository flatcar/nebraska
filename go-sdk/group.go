package nebraska

import (
	"context"
	"net/http"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

// Client for single instance.
type Group interface {
	Update(ctx context.Context, conf GroupConfig, opts *UpdateGroupOptions) (Group, error)
	Delete(ctx context.Context, opts *DeleteGroupOptions) error
	Props() codegen.Group
	Response() *http.Response
}

type Groups struct {
	TotalCount int
	Groups     []Group
	rawResponse
}

type group struct {
	groupClient GroupsClient
	// channelClient ChannelsClient
	properties codegen.Group
	rawResponse
}

func (group group) Update(ctx context.Context, conf GroupConfig, opts *UpdateGroupOptions) (Group, error) {
	return group.groupClient.Update(ctx, group.properties.Id, conf, opts)
}

func (group group) Delete(ctx context.Context, opts *DeleteGroupOptions) error {
	return group.groupClient.Delete(ctx, group.properties.ApplicationID, opts)
}

func (group group) Props() codegen.Group {
	return group.properties
}

func (group group) Response() *http.Response {
	return group.resp
}
