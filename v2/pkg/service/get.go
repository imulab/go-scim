package service

import (
	"context"
	"github.com/imulab/go-scim/v2/pkg/crud"
	"github.com/imulab/go-scim/v2/pkg/db"
	"github.com/imulab/go-scim/v2/pkg/prop"
)

// GetService returns a Get service.
func GetService(database db.DB) Get {
	return &getService{database: database}
}

type (
	Get interface {
		Do(ctx context.Context, req *GetRequest) (resp *GetResponse, err error)
	}
	GetRequest struct {
		ResourceID string
		Projection *crud.Projection
	}
	GetResponse struct {
		Resource *prop.Resource
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
