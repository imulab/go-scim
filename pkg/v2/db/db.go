package db

import (
	"context"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/prop"
)

// DB is the abstraction for the database that provides the persistence and look up capabilities.
type DB interface {
	// Insert the given resource into the database, or return any error.
	Insert(ctx context.Context, resource *prop.Resource) error
	// Count the number of resources that meets the given SCIM filter.
	Count(ctx context.Context, filter string) (int, error)
	// Get a resource by its id. The projection parameter specifies the attributes to be included or excluded from the
	// response. Implementations may elect to ignore this parameter in case caller services need all the attributes for
	// additional processing.
	Get(ctx context.Context, id string, projection *crud.Projection) (*prop.Resource, error)
	// Replace overwrites an existing reference resource with the content of the replacement resource. The reference
	// and the replacement resource are supposed to have the same id.
	Replace(ctx context.Context, ref *prop.Resource, replacement *prop.Resource) error
	// Delete a resource
	Delete(ctx context.Context, resource *prop.Resource) error
	// Query resources. The projection parameter specifies the attributes to be included or excluded from the
	// response. Implementations may elect to ignore this parameter in case caller services need all the attributes for
	// additional processing.
	Query(ctx context.Context, filter string, sort *crud.Sort, pagination *crud.Pagination, projection *crud.Projection) ([]*prop.Resource, error)
}
