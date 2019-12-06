package services

import (
	"context"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
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
		Logger      protocol.LogProvider
		Persistence protocol.PersistenceProvider
	}
)

func (s *GetService) GetResource(ctx context.Context, request *GetRequest) (*GetResponse, error) {
	s.Logger.Debug("received get request [id=%s]", request.ResourceID)

	resource, err := s.Persistence.Get(ctx, request.ResourceID)
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
