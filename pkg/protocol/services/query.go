package services

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
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
		Logger           log.Logger
		Database         db.DB
		TooManyThreshold int
	}
)

func (s *QueryService) QueryResource(ctx context.Context, request *QueryRequest) (resp *QueryResponse, err error) {
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

	if (request.Pagination == nil && resp.TotalResults > s.TooManyThreshold) ||
		(request.Pagination != nil && request.Pagination.Count > s.TooManyThreshold) {
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
