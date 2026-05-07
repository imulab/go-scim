package service

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/json"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// QueryService returns a query resource service. This service is only capable of performing querying on a single type
// of resource. This does not handle root query.
func QueryService(config *spec.ServiceProviderConfig, database db.DB) Query {
	return &queryService{
		database: database,
		config:   config,
	}
}

type (
	// Query resource service
	Query interface {
		Do(ctx context.Context, req *QueryRequest) (resp *QueryResponse, err error)
	}
	// Query resource request
	QueryRequest struct {
		Filter     string
		Sort       *crud.Sort
		Pagination *crud.Pagination
		Projection *crud.Projection
	}
	// Query resource response
	QueryResponse struct {
		TotalResults int
		StartIndex   int
		ItemsPerPage int
		Resources    []json.Serializable
		Projection   *crud.Projection // included so that caller may render properly
	}
)

type queryService struct {
	database db.DB
	config   *spec.ServiceProviderConfig
}

func (s *queryService) Do(ctx context.Context, req *QueryRequest) (resp *QueryResponse, err error) {
	if err = s.checkSupport(req); err != nil {
		return
	}

	if err = req.ValidateAndDefault(); err != nil {
		return
	}

	resp = new(QueryResponse)
	resp.Projection = req.Projection

	if req.Pagination != nil {
		resp.StartIndex = req.Pagination.StartIndex
	}

	if resp.TotalResults, err = s.database.Count(ctx, req.Filter); err != nil {
		return
	}
	if req.Pagination != nil && req.Pagination.Count == 0 {
		return
	}

	if s.config.Filter.MaxResults > 0 {
		if (req.Pagination == nil && resp.TotalResults > s.config.Filter.MaxResults) ||
			(req.Pagination != nil && req.Pagination.Count > s.config.Filter.MaxResults) {
			err = spec.ErrTooMany
			return
		}
	}

	resources, err := s.database.Query(ctx, req.Filter, req.Sort, req.Pagination, req.Projection)
	if err != nil {
		return
	}
	for _, r := range resources {
		resp.Resources = append(resp.Resources, r)
	}

	resp.ItemsPerPage = len(resp.Resources)
	return
}

func (s *queryService) checkSupport(request *QueryRequest) error {
	if !s.config.Filter.Supported {
		if len(request.Filter) > 0 {
			return fmt.Errorf("%w: filter is not supported", spec.ErrInvalidSyntax)
		}
	}

	if !s.config.Sort.Supported {
		if request.Sort != nil && len(request.Sort.By) > 0 {
			return fmt.Errorf("%w: sorting is not supported", spec.ErrInvalidSyntax)
		}
	}

	return nil
}

func (q *QueryRequest) ValidateAndDefault() error {
	if len(q.Filter) == 0 {
		q.Filter = "id pr"
	} else {
		if _, err := expr.CompileFilter(q.Filter); err != nil {
			return err
		}
	}
	if q.Pagination != nil {
		if q.Pagination.StartIndex <= 0 {
			q.Pagination.StartIndex = 1
		}
	}
	if q.Sort != nil {
		if len(q.Sort.By) == 0 {
			q.Sort.By = "id"
		} else {
			if _, err := expr.CompilePath(q.Sort.By); err != nil {
				return err
			}
		}
		switch q.Sort.Order {
		case "", crud.SortAsc, crud.SortDesc:
		default:
			return fmt.Errorf("%w: invalid sortOrder", spec.ErrInvalidSyntax)
		}
	}
	if q.Projection != nil {
		if len(q.Projection.Attributes) > 0 && len(q.Projection.ExcludedAttributes) > 0 {
			return fmt.Errorf("%w: only one of attributes and excludedAttributes may be used", spec.ErrInvalidSyntax)
		}
		if len(q.Projection.Attributes) > 0 {
			for _, p := range q.Projection.Attributes {
				if _, err := expr.CompilePath(p); err != nil {
					return err
				}
			}
		}
		if len(q.Projection.ExcludedAttributes) > 0 {
			for _, p := range q.Projection.ExcludedAttributes {
				if _, err := expr.CompilePath(p); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
