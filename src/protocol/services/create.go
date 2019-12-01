package services

import (
	"context"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
)

type (
	CreateRequest struct {
		Payload *prop.Resource
	}
	CreateResponse struct {
		Resource *prop.Resource
		Location string
		Version  string
	}
	CreateService struct {
		Logger      protocol.LogProvider
		Filters     []protocol.ResourceFilter
		Persistence protocol.PersistenceProvider
		Events      protocol.EventPublisher
	}
)

func (s *CreateService) CreateResource(ctx context.Context, request *CreateRequest) (cr *CreateResponse, err error) {
	s.Logger.Debug("received create request")

	for _, filter := range s.Filters {
		err = filter.Filter(ctx, request.Payload)
		if err != nil {
			s.Logger.Error("create request encounter error during filter: %s", err.Error())
			return
		}
	}

	var (
		id = request.Payload.ID()
		location = request.Payload.Location()
		version = request.Payload.Version()
	)
	s.Logger.Debug("resource [id=%s] passed create filters", id)

	err = s.Persistence.Insert(ctx, request.Payload)
	if err != nil {
		s.Logger.Error("resource [id=%s] failed to insert into persistence: %s", err.Error())
		return
	}
	s.Logger.Debug("resource [id=%s] inserted into persistence", id)

	if s.Events != nil {
		s.Events.ResourceCreated(ctx, request.Payload)
	}

	cr = &CreateResponse{
		Resource: request.Payload,
		Location: location,
		Version:  version,
	}
	return
}
