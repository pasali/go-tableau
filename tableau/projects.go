package tableau

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ProjectContentPermission represents a projects' content permissions
type ProjectContentPermission string

const (
	ProjectContentPermissionLockedToProject              ProjectContentPermission = "LockedToProject"
	ProjectContentPermissionManagedByOwner               ProjectContentPermission = "ManagedByOwner"
	ProjectContentPermissionLockedToProjectWithoutNested ProjectContentPermission = "LockedToProjectWithoutNested"
)

type projectsService struct {
	client *Client
}

func (ps *projectsService) Query(ctx context.Context, opts ...QueryOption) ([]*Project, error) {
	path := fmt.Sprintf("sites/%s/projects", ps.client.SiteID)

	queryOpts := &QueryOptions{
		URLValues: &url.Values{},
	}

	for _, opt := range opts {
		err := opt(queryOpts)
		if err != nil {
			panic(err)
		}
	}

	if vals := queryOpts.URLValues.Encode(); vals != "" {
		path += "?" + vals
	}
	req, err := ps.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request for query projects")
	}

	resp := &queryProjectResponse{}
	err = ps.client.do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Projects.Project, nil
}

func (ps *projectsService) Create(ctx context.Context, createReq *CreateProjectRequest) (*Project, error) {
	path := fmt.Sprintf("sites/%s/projects", ps.client.SiteID)

	request := struct {
		Project *CreateProjectRequest `json:"project"`
	}{
		Project: createReq,
	}

	req, err := ps.client.newRequest(http.MethodPost, path, request)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request for create project")
	}
	resp := &createProjectResponse{}
	err = ps.client.do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Project, nil
}

func (ps *projectsService) Update(ctx context.Context, updateReq *UpdateProjectRequest) (*Project, error) {
	path := fmt.Sprintf("sites/%s/projects/%s", ps.client.SiteID, updateReq.ID)

	request := struct {
		Project *UpdateProjectRequest `json:"project"`
	}{
		Project: updateReq,
	}
	req, err := ps.client.newRequest(http.MethodPut, path, request)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request for update project")
	}
	resp := &createProjectResponse{}
	err = ps.client.do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Project, nil
}

func (ps *projectsService) Delete(ctx context.Context, deleteReq *DeleteProjectRequest) (*Project, error) {
	path := fmt.Sprintf("sites/%s/projects/%s", ps.client.SiteID, deleteReq.ID)
	req, err := ps.client.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request for delete project")
	}
	resp := &createProjectResponse{}
	err = ps.client.do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Project, nil
}

// QueryOptions are options for querying projects.
type QueryOptions struct {
	URLValues *url.Values
}

type QueryOption func(*QueryOptions) error

// WithPageSize returns a QueryOption that sets the "pageSize" URL parameter.
func WithPageSize(pageSize int) QueryOption {
	return func(opt *QueryOptions) error {
		if pageSize > 0 {
			opt.URLValues.Set("pageSize", strconv.Itoa(pageSize))
		}
		return nil
	}
}

// WithPageNumber returns a QueryOption that sets the "pageNumber" URL parameter.
func WithPageNumber(pageNumber int) QueryOption {
	return func(opt *QueryOptions) error {
		if pageNumber > 0 {
			opt.URLValues.Set("pageNumber", strconv.Itoa(pageNumber))
		}
		return nil
	}
}

// WithFilterExpression returns a QueryOption that sets the "filter" URL parameter.
func WithFilterExpression(filterExp string) QueryOption {
	return func(opt *QueryOptions) error {
		if filterExp != "" {
			opt.URLValues.Set("filter", filterExp)
		}
		return nil
	}
}

// WithSortExpression returns a QueryOption that sets the "sort" URL parameter.
func WithSortExpression(sortExp string) QueryOption {
	return func(opt *QueryOptions) error {
		if sortExp != "" {
			opt.URLValues.Set("sort", sortExp)
		}
		return nil
	}
}

// CreateProjectRequest encapsulates the request for creating a new project.
type CreateProjectRequest struct {
	ParentProjectId    string                   `json:"parentProjectId,omitempty"`
	Name               string                   `json:"name"`
	Description        string                   `json:"description,omitempty"`
	ContentPermissions ProjectContentPermission `json:"contentPermissions,omitempty"`
}

type createProjectResponse struct {
	Project *Project `json:"project"`
}

// DeleteProjectRequest encapsulates the request for deleting a new project.
type DeleteProjectRequest struct {
	ID string `json:"id"`
}

type queryProjectResponse struct {
	Pagination struct {
		PageSize       string `json:"pageSize"`
		PageNumber     string `json:"pageNumber"`
		TotalAvailable string `json:"totalAvailabe"`
	}
	Projects struct {
		Project []*Project `json:"project"`
	}
}

// UpdateProjectRequest encapsulates the request for updating project.
type UpdateProjectRequest struct {
	ID                 string                   `json:"id,"`
	ParentProjectId    string                   `json:"parentProjectId,omitempty"`
	Name               string                   `json:"name"`
	Description        string                   `json:"description,omitempty"`
	ContentPermissions ProjectContentPermission `json:"contentPermissions,omitempty"`
}

// Project represents a Tableau project
type Project struct {
	ID                              string `json:"id"`
	ParentProjectId                 string `json:"parentProjectId"`
	Name                            string `json:"name"`
	Description                     string `json:"description"`
	ContentPermissions              string `json:"contentPermissions"`
	ControllingPermissionsProjectId string `json:"controllingPermissionsProjectId"`
	Writeable                       bool   `json:"writeable"`
	TopLevelProject                 bool   `json:"topLevelProject"`
	Owner                           struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		Name      string    `json:"name"`
		FullName  string    `json:"fullName"`
		SiteRole  string    `json:"siteRole"`
		LastLogin time.Time `json:"lastLogin"`
	}
	ContentCounts struct {
		ProjectCount    int `json:"projectCount"`
		WorkbookCount   int `json:"workbookCount"`
		ViewCount       int `json:"viewCount"`
		DatasourceCount int `json:"datasourceCount"`
	}
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}
