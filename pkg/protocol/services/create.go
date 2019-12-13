package services

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/event"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/imulab/go-scim/pkg/protocol/services/filter"
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
		Logger   log.Logger
		Filters  []filter.ForResource
		Database db.DB
		Event    event.Publisher
	}
)

func (s *CreateService) CreateResource(ctx context.Context, request *CreateRequest) (cr *CreateResponse, err error) {
	s.Logger.Debug("received create request")

	for _, f := range s.Filters {
		err = f.Filter(ctx, request.Payload)
		if err != nil {
			s.Logger.Error("create request encounter error during filter: %s", err.Error())
			return
		}
	}

	err = s.Database.Insert(ctx, request.Payload)
	if err != nil {
		s.Logger.Error("resource [id=%s] failed to insert into persistence: %s", request.Payload.ID(), err.Error())
		return
	}
	s.Logger.Debug("resource [id=%s] inserted into persistence", request.Payload.ID())

	if s.Event != nil {
		s.Event.ResourceCreated(ctx, request.Payload)
	}

	cr = &CreateResponse{
		Resource: request.Payload,
		Location: request.Payload.Location(),
		Version:  request.Payload.Version(),
	}
	return
}
