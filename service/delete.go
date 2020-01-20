package service

import (
	"context"
	"fmt"
	"github.com/elvsn/scim.go/db"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
)

func DeleteService(config *spec.ServiceProviderConfig, database db.DB) Delete {
	return &deleteService{
		Database: database,
		Config:   config,
	}
}

type (
	Delete interface {
		Do(ctx context.Context, req *DeleteRequest) error
	}
	DeleteRequest struct {
		ResourceID    string
		MatchCriteria func(resource *prop.Resource) bool
	}
)

type deleteService struct {
	Database db.DB
	Config   *spec.ServiceProviderConfig
}

func (s *deleteService) Do(ctx context.Context, req *DeleteRequest) (err error) {
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
	return
}
