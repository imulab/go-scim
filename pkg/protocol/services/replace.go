package services

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol"
)

type (
	ReplaceRequest struct {
		ResourceID    string
		Payload       *prop.Resource
		MatchCriteria func(resource *prop.Resource) bool
	}
	ReplaceResponse struct {
		Resource   *prop.Resource
		Location   string
		OldVersion string
		NewVersion string
	}
	ReplaceService struct {
		Logger      protocol.LogProvider
		Lock        protocol.LockProvider
		Filters     []protocol.ResourceFilter
		Persistence protocol.PersistenceProvider
		Events      protocol.EventPublisher
	}
)

func (s *ReplaceService) ReplaceResource(ctx context.Context, request *ReplaceRequest) (*ReplaceResponse, errors) {
	s.Logger.Debug("received replace request [id=%s]", request.ResourceID)

	ref, err := s.Persistence.Get(ctx, request.ResourceID)
	if err != nil {
		return nil, err
	} else if request.MatchCriteria != nil && !request.MatchCriteria(ref) {
		return nil, errors.PreConditionFailed("resource [id=%s] does not meet pre condition", request.ResourceID)
	}

	defer s.Lock.Unlock(ctx, ref)
	if err := s.Lock.Lock(ctx, ref); err != nil {
		s.Logger.Error("failed to obtain lock for resource [id=%s]: %s", request.ResourceID, err.Error())
		return nil, err
	}

	fctx := protocol.NewFilterContext(ctx)
	for _, filter := range s.Filters {
		if err := filter.FilterRef(fctx, request.Payload, ref); err != nil {
			s.Logger.Error("replace request encounter error during filter for resource [id=%s]: %s", request.ResourceID, err.Error())
			return nil, err
		}
	}

	// Only replace when version is bumped
	if request.Payload.Version() != ref.Version() {
		err = s.Persistence.Replace(ctx, request.Payload)
		if err != nil {
			s.Logger.Error("resource [id=%s] failed to save into persistence: %s", request.ResourceID, err.Error())
			return nil, err
		}
		s.Logger.Debug("resource [id=%s] saved in persistence", request.ResourceID)

		if s.Events != nil {
			s.Events.ResourceUpdated(ctx, request.Payload)
		}
	}

	return &ReplaceResponse{
		Resource:   request.Payload,
		Location:   request.Payload.Location(),
		OldVersion: ref.Version(),
		NewVersion: request.Payload.Version(),
	}, nil
}
