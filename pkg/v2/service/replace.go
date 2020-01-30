package service

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/service/filter"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"io"
	"io/ioutil"
)

// ReplaceService returns a replace service.
func ReplaceService(
	config *spec.ServiceProviderConfig,
	resourceType *spec.ResourceType,
	database db.DB,
	filters []filter.ByResource,
) Replace {
	return &replaceService{
		resourceType: resourceType,
		filters:      filters,
		database:     database,
		config:       config,
	}
}

type (
	// Replace resource service
	Replace interface {
		Do(ctx context.Context, req *ReplaceRequest) (resp *ReplaceResponse, err error)
	}
	// Replace resource request
	ReplaceRequest struct {
		ResourceID    string                             // id of the resource to be replaced
		PayloadSource io.Reader                          // source to read replacement payload from
		MatchCriteria func(resource *prop.Resource) bool // extra criteria to meet in order to be replaced
	}
	// Replace resource response
	ReplaceResponse struct {
		Replaced bool           // true if resource was replaced; false if resource was not replaced, but has no error
		Ref      *prop.Resource // reference resource (before state)
		Resource *prop.Resource // replaced resource (after state)
	}
)

type replaceService struct {
	resourceType *spec.ResourceType
	filters      []filter.ByResource
	database     db.DB
	config       *spec.ServiceProviderConfig
}

func (s *replaceService) Do(ctx context.Context, req *ReplaceRequest) (resp *ReplaceResponse, err error) {
	ref, err := s.database.Get(ctx, req.ResourceID, nil)
	if err != nil {
		return
	}

	if s.config.ETag.Supported && req.MatchCriteria != nil {
		if !req.MatchCriteria(ref) {
			err = fmt.Errorf("%w: resource does not meet pre condition", spec.ErrConflict)
			return
		}
	}

	replacement, err := s.parseResource(req)
	if err != nil {
		return
	}

	for _, f := range s.filters {
		if err = f.FilterRef(ctx, replacement, ref); err != nil {
			return
		}
	}

	var (
		newVersion = replacement.MetaVersionOrEmpty()
		oldVersion = ref.MetaVersionOrEmpty()
	)
	if newVersion == oldVersion {
		resp = &ReplaceResponse{
			Replaced: false,
			Ref:      ref,
		}
		return
	}

	if err = s.database.Replace(ctx, ref, replacement); err != nil {
		return
	}

	resp = &ReplaceResponse{
		Replaced: true,
		Resource: replacement,
		Ref:      ref,
	}
	return
}

func (s *replaceService) parseResource(req *ReplaceRequest) (*prop.Resource, error) {
	if req == nil || req.PayloadSource == nil {
		return nil, fmt.Errorf("%w: no payload for replace service", spec.ErrInternal)
	}

	raw, err := ioutil.ReadAll(req.PayloadSource)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read request body", spec.ErrInternal)
	}

	resource := prop.NewResource(s.resourceType)
	if err := json.Deserialize(raw, resource); err != nil {
		return nil, err
	}

	return resource, nil
}
