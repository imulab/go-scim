package services

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/log"
)

type (
	GetRequest struct {
		ResourceID string
	}
	GetResponse struct {
		Resource *prop.Resource
		Location string
		Version  string
	}
	GetService struct {
		Logger   log.Logger
		Database db.DB
	}
)

func (s *GetService) GetResource(ctx context.Context, request *GetRequest) (*GetResponse, error) {
	s.Logger.Debug("received get request [id=%s]", request.ResourceID)

	resource, err := s.Database.Get(ctx, request.ResourceID)
	if err != nil {
		s.Logger.Error("failed to get resource [id=%s] from persistence: %s", request.ResourceID, err.Error())
		return nil, err
	}

	return &GetResponse{
		Resource: resource,
		Location: resource.Location(),
		Version:  resource.Version(),
	}, nil
}
