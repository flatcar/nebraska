package nebraska

import (
	"context"
	"net/http"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

type Application interface {
	Update(ctx context.Context, conf AppConfig, opts *UpdateAppOptions) (Application, error)
	Delete(ctx context.Context, opts *DeleteAppOptions) error
	Props() codegen.Application
	CreateGroup(ctx context.Context, conf GroupConfig, opts *CreateGroupOptions) (Group, error)
	PaginateGroups(ctx context.Context, page int, perpage int, opts *PaginateGroupOptions) (*Groups, error)
	Response() *http.Response
}

type Applications struct {
	TotalCount int
	Apps       []Application
	rawResponse
}

type application struct {
	appsClient   ApplicationsClient
	groupsClient GroupsClient
	properties   codegen.Application
	rawResponse
}

func (app application) Update(ctx context.Context, conf AppConfig, opts *UpdateAppOptions) (Application, error) {
	return app.appsClient.Update(ctx, app.properties.Id, conf, opts)
}

func (app application) Delete(ctx context.Context, opts *DeleteAppOptions) error {
	return app.appsClient.Delete(ctx, app.properties.Id, opts)
}

func (app application) Props() codegen.Application {
	return app.properties
}

func (app application) Response() *http.Response {
	return app.resp
}

func (app application) CreateGroup(ctx context.Context, conf GroupConfig, opts *CreateGroupOptions) (Group, error) {
	return app.groupsClient.Create(ctx, conf, opts)
}

func (app application) PaginateGroups(ctx context.Context, page int, perpage int, opts *PaginateGroupOptions) (*Groups, error) {
	return app.groupsClient.Paginate(ctx, page, perpage, opts)
}
