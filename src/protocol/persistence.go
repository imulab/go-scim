package protocol

import (
	"context"
	"github.com/imulab/go-scim/src/core/prop"
)

type PersistenceProvider interface {
	// Insert the given resource into the database, or return any error.
	Insert(ctx context.Context, resource *prop.Resource) error
	// Count the number of resources that meets the given SCIM filter.
	Count(ctx context.Context, filter string) (int, error)
	// Get a resource by its id
	Get(ctx context.Context, id string) (*prop.Resource, error)
	// Overwrite the existing resource with same ID with the new resource
	Replace(ctx context.Context, resource *prop.Resource) error
	// Delete a resource by its id
	Delete(ctx context.Context, id string) error
}
