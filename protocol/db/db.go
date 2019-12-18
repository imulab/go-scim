package db

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/crud"
)

// Abstraction for the database that provides the main persistence and look up capabilities.
type DB interface {
	// Insert the given resource into the database, or return any error.
	Insert(ctx context.Context, resource *prop.Resource) error
	// Count the number of resources that meets the given SCIM filter.
	Count(ctx context.Context, filter string) (int, error)
	// Get a resource by its id. The projection parameter specifies the attributes to be included or excluded from the
	// response. Implementations may elect to ignore this parameter in case caller services need all the attributes for
	// additional processing.
	Get(ctx context.Context, id string, projection *crud.Projection) (*prop.Resource, error)
	// Overwrite the existing resource with same ID with the new resource
	Replace(ctx context.Context, resource *prop.Resource) error
	// Delete a resource by its id
	Delete(ctx context.Context, id string) error
	// Query resources. The projection parameter specifies the attributes to be included or excluded from the
	// response. Implementations may elect to ignore this parameter in case caller services need all the attributes for
	// additional processing.
	Query(ctx context.Context, filter string, sort *crud.Sort, pagination *crud.Pagination, projection *crud.Projection) ([]*prop.Resource, error)
}
