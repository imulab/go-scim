package services

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/expr"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/crud"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/log"
)

type (
	QueryRequest struct {
		Filter     string
		Sort       *crud.Sort
		Pagination *crud.Pagination
		Projection *crud.Projection
	}
	QueryResponse struct {
		TotalResults int
		StartIndex   int
		ItemsPerPage int
		Resources    []*prop.Resource
	}
	QueryService struct {
		Logger                log.Logger
		Database              db.DB
		ServiceProviderConfig *spec.ServiceProviderConfig
	}
)

func (s *QueryService) checkSupport(request *QueryRequest) error {
	if !s.ServiceProviderConfig.Filter.Supported {
		if len(request.Filter) > 0 {
			return errors.InvalidRequest("filter is not supported")
		}
	}
	if !s.ServiceProviderConfig.Sort.Supported {
		if request.Sort != nil && len(request.Sort.By) > 0 {
			return errors.InvalidRequest("sort is not supported")
		}
	}
	return nil
}

func (s *QueryService) QueryResource(ctx context.Context, request *QueryRequest) (resp *QueryResponse, err error) {
	err = s.checkSupport(request)
	if err != nil {
		return
	}

	err = request.ValidateAndDefault()
	if err != nil {
		return
	}

	resp = new(QueryResponse)
	if request.Pagination != nil {
		resp.StartIndex = request.Pagination.StartIndex
	}

	resp.TotalResults, err = s.Database.Count(ctx, request.Filter)
	if err != nil {
		return
	} else if request.Pagination != nil && request.Pagination.Count == 0 {
		return
	}

	if (request.Pagination == nil && resp.TotalResults > s.ServiceProviderConfig.Filter.MaxResults) ||
		(request.Pagination != nil && request.Pagination.Count > s.ServiceProviderConfig.Filter.MaxResults) {
		err = errors.TooMany("request would return too many results")
		return
	}

	resp.Resources, err = s.Database.Query(ctx, request.Filter, request.Sort, request.Pagination, request.Projection)
	if err != nil {
		s.Logger.Error("failed to query resource: %s", err.Error())
		return
	}
	resp.ItemsPerPage = len(resp.Resources)

	return
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
			return errors.InvalidValue("sortOrder is invalid")
		}
	}
	if q.Projection != nil {
		if len(q.Projection.Attributes) > 0 && len(q.Projection.ExcludedAttributes) > 0 {
			return errors.InvalidValue("only one of attributes and excludedAttributes may be used")
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
