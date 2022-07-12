package tableau

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

// DeleteDataSourceRequest encapsulates the request for deleting a single DataSource.
type DeleteDataSourceRequest struct {
	ID string
}

// GetDataSourceRequest encapsulates the request for getting a single DataSource.
type GetDataSourceRequest struct {
	ID string
}

type dataSourcesResponse struct {
	DataSource *DataSource `json:"dataSource"`
}

// DataSource represents a Tableau data source
type DataSource struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	CertificationNote   string            `json:"CertificationNote"`
	ContentUrl          string            `json:"contentUrl"`
	EncryptExtracts     string            `json:"encryptExtracts"`
	Description         string            `json:"description"`
	WebpageUrl          string            `json:"webpageUrl"`
	IsCertified         bool              `json:"isCertified"`
	UseRemoteQueryAgent bool              `json:"useRemoteQueryAgent"`
	Type                string            `json:"type"`
	Tags                map[string]string `json:"tags"`
	Owner               struct {
		ID string `json:"id"`
	}
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

type dataSourcesService struct {
	client *Client
}

func (dss *dataSourcesService) Get(ctx context.Context, getReq *GetDataSourceRequest) (*DataSource, error) {
	path := fmt.Sprintf("sites/%s/datasources/%s", dss.client.SiteID, getReq.ID)
	req, err := dss.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request for get datasource")
	}

	ds := &dataSourcesResponse{}
	err = dss.client.do(ctx, req, &ds)
	if err != nil {
		return nil, err
	}

	return ds.DataSource, nil
}

func (dss *dataSourcesService) Delete(ctx context.Context, delReq *DeleteDataSourceRequest) error {
	path := fmt.Sprintf("sites/%s/datasources/%s", dss.client.SiteID, delReq.ID)
	req, err := dss.client.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return errors.Wrap(err, "error creating request for deleting datasource")
	}
	err = dss.client.do(ctx, req, nil)
	return err
}
