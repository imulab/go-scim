package services

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/event"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services/filter"
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
	for _, f := range s.Filters {
		err = f.Filter(ctx, request.Payload)
		if err != nil {
			s.Logger.Error("create request encounter error during filter.", log.Args{
				"error": err,
			})
			return
		}
	}

	err = s.Database.Insert(ctx, request.Payload)
	if err != nil {
		s.Logger.Error("failed to insert into persistence", log.Args{
			"resourceId": request.Payload.ID(),
			"error":      err,
		})
		return
	}
	s.Logger.Debug("inserted into persistence", log.Args{
		"resourceId": request.Payload.ID(),
	})

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
