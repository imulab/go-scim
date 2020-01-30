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

// Create returns a create resource service.
func CreateService(resourceType *spec.ResourceType, database db.DB, filters []filter.ByResource) Create {
	return &createService{
		resourceType: resourceType,
		filters:      filters,
		database:     database,
	}
}

type (
	// Create resource service
	Create interface {
		Do(ctx context.Context, req *CreateRequest) (resp *CreateResponse, err error)
	}
	// Create resource request
	CreateRequest struct {
		PayloadSource io.Reader // reader source to read resource payload from
	}
	// Create resource response
	CreateResponse struct {
		Resource *prop.Resource // the created resource
	}
)

type createService struct {
	resourceType *spec.ResourceType
	filters      []filter.ByResource
	database     db.DB
}

func (s *createService) Do(ctx context.Context, req *CreateRequest) (resp *CreateResponse, err error) {
	resource, err := s.parseResource(req)
	if err != nil {
		return
	}

	for _, f := range s.filters {
		if err = f.Filter(ctx, resource); err != nil {
			return
		}
	}

	if err = s.database.Insert(ctx, resource); err != nil {
		return
	}

	resp = &CreateResponse{Resource: resource}
	return
}

func (s *createService) parseResource(req *CreateRequest) (*prop.Resource, error) {
	if req == nil || req.PayloadSource == nil {
		return nil, fmt.Errorf("%w: no payload for create service", spec.ErrInternal)
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
