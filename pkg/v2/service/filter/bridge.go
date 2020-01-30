package filter

import (
	"context"
	"github.com/imulab/go-scim/pkg/v2/prop"
)

// ByPropertyToByResource returns a ByResource that iterates each property in the resource using a DFS visitor
// and sequentially invoke the list of ByProperty filters on each visited property. It effectively bridges ByProperty
// filters to a ByResource filters.
func ByPropertyToByResource(filters ...ByProperty) ByResource {
	return bridgeResourceFilter{byPropertyFilters: filters}
}

type bridgeResourceFilter struct {
	byPropertyFilters []ByProperty
}

func (f bridgeResourceFilter) Filter(ctx context.Context, resource *prop.Resource) error {
	return Visit(ctx, resource, f.byPropertyFilters...)
}

func (f bridgeResourceFilter) FilterRef(ctx context.Context, resource *prop.Resource, ref *prop.Resource) error {
	return VisitWithRef(ctx, resource, ref, f.byPropertyFilters...)
}
