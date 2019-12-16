package services

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/event"
	"github.com/imulab/go-scim/pkg/protocol/log"
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
	s.Logger.Debug("received delete request [id=%s]", request.ResourceID)

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
		s.Logger.Error("resource [id=%s] failed to delete from persistence: %s", request.ResourceID, err.Error())
		return err
	}
	s.Logger.Debug("resource [id=%s] deleted from persistence", request.ResourceID)

	if s.Event != nil {
		s.Event.ResourceDeleted(ctx, resource)
	}

	return nil
}
