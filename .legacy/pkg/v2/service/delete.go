package service

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// DeleteService returns a delete resource service
func DeleteService(config *spec.ServiceProviderConfig, database db.DB) Delete {
	return &deleteService{
		Database: database,
		Config:   config,
	}
}

type (
	// Delete resource service
	Delete interface {
		Do(ctx context.Context, req *DeleteRequest) (resp *DeleteResponse, err error)
	}
	// Delete resource request
	DeleteRequest struct {
		ResourceID    string                             // id of the resource to be deleted
		MatchCriteria func(resource *prop.Resource) bool // extra criteria the resource has to meet in order to be deleted
	}
	// Delete resource response
	DeleteResponse struct {
		Deleted *prop.Resource // the resource that was deleted
	}
)

type deleteService struct {
	Database db.DB
	Config   *spec.ServiceProviderConfig
}

func (s *deleteService) Do(ctx context.Context, req *DeleteRequest) (resp *DeleteResponse, err error) {
	resource, err := s.Database.Get(ctx, req.ResourceID, nil)
	if err != nil {
		return
	}

	if s.Config.ETag.Supported && req.MatchCriteria != nil {
		if !req.MatchCriteria(resource) {
			err = fmt.Errorf("%w: resource does not meet pre condition", spec.ErrConflict)
			return
		}
	}

	err = s.Database.Delete(ctx, resource)
	if err != nil {
		return
	}

	resp = &DeleteResponse{Deleted: resource}
	return
}
