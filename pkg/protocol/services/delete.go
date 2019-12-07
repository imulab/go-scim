package services

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol"
)

type (
	DeleteRequest struct {
		ResourceID    string
		MatchCriteria func(resource *prop.Resource) bool
	}
	DeleteService struct {
		Logger      protocol.LogProvider
		Lock        protocol.LockProvider
		Persistence protocol.PersistenceProvider
		Events      protocol.EventPublisher
	}
)

func (s *DeleteService) DeleteResource(ctx context.Context, request *DeleteRequest) errors {
	s.Logger.Debug("received delete request [id=%s]", request.ResourceID)

	resource, err := s.Persistence.Get(ctx, request.ResourceID)
	if err != nil {
		return err
	} else if request.MatchCriteria != nil && !request.MatchCriteria(resource) {
		return errors.PreConditionFailed("resource [id=%s] does not meet pre condition", request.ResourceID)
	}

	defer s.Lock.Unlock(ctx, resource)
	if err := s.Lock.Lock(ctx, resource); err != nil {
		s.Logger.Error("failed to obtain lock for resource [id=%s]: %s", request.ResourceID, err.Error())
		return err
	}

	err = s.Persistence.Delete(ctx, request.ResourceID)
	if err != nil {
		s.Logger.Error("resource [id=%s] failed to delete from persistence: %s", request.ResourceID, err.Error())
		return err
	}
	s.Logger.Debug("resource [id=%s] deleted from persistence", request.ResourceID)

	if s.Events != nil {
		s.Events.ResourceDeleted(ctx, resource)
	}

	return nil
}
