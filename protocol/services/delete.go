package services

import (
	"context"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/event"
	"github.com/imulab/go-scim/protocol/log"
)

type (
	DeleteRequest struct {
		ResourceID    string
		MatchCriteria func(resource *prop.Resource) bool
	}
	DeleteService struct {
		Logger                log.Logger
		Database              db.DB
		Event                 event.Publisher
		ServiceProviderConfig *spec.ServiceProviderConfig
	}
)

func (s *DeleteService) DeleteResource(ctx context.Context, request *DeleteRequest) error {
	resource, err := s.Database.Get(ctx, request.ResourceID, nil)
	if err != nil {
		return err
	}
	if s.ServiceProviderConfig.ETag.Supported && request.MatchCriteria != nil {
		if !request.MatchCriteria(resource) {
			return errors.PreConditionFailed("resource [id=%s] does not meet pre condition", request.ResourceID)
		}
	}

	err = s.Database.Delete(ctx, request.ResourceID)
	if err != nil {
		s.Logger.Error("failed to delete from persistence", log.Args{
			"resourceId": request.ResourceID,
			"error":      err,
		})
		return err
	}
	s.Logger.Debug("deleted from persistence", log.Args{
		"resourceId": request.ResourceID,
	})

	if s.Event != nil {
		s.Event.ResourceDeleted(ctx, resource)
	}

	return nil
}
