package service

import (
	"context"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
)

// GetService returns a get resource service.
func GetService(database db.DB) Get {
	return &getService{database: database}
}

type (
	// Get resource resource
	Get interface {
		Do(ctx context.Context, req *GetRequest) (resp *GetResponse, err error)
	}
	// Get resource request
	GetRequest struct {
		ResourceID string           // id of the resource to get
		Projection *crud.Projection // field projection to be considered when fetching resource
	}
	// Get resource response
	GetResponse struct {
		Resource *prop.Resource // resource got from database
	}
)

type getService struct {
	database db.DB
}

func (s *getService) Do(ctx context.Context, req *GetRequest) (resp *GetResponse, err error) {
	resource, err := s.database.Get(ctx, req.ResourceID, req.Projection)
	if err != nil {
		return
	}

	resp = &GetResponse{Resource: resource}
	return
}
