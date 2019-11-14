package protocol

import (
	"context"
	"github.com/imulab/go-scim/core"
)

// Interface for providing persistence services. This interface is the central point of all
// state persistence and management for the endpoints.
type PersistenceProvider interface {
	// Returns true if this provider is capable of performing resource filter. For providers capable
	// of performing filter, the filter will be done at the provider, and is expected to return filtered
	// results; for providers incapable of performing filter, the filtering operation will be attempted
	// to be done in process. Note that the service may refuse to process a large number of results and
	// return a 'tooMany' error instead.
	IsFilterSupported() bool

	// Returns true if this provider is capable of performing pagination. For providers capable of performing
	// pagination, the pagination will be done at the provider, and is expected to return only results within
	// the page; for providers incapable of performing pagination, the service will attempt to perform pagination
	// in process. However, the service may refuse to process a large number of results and return 'tooMany' error
	// instead.
	IsPaginationSupported() bool

	// Returns true if this provider is capable of return sorted results. For providers capable of sorting, the sort
	// will be done at the provider, and it is expected for the provider to return sorted results; for providers
	// incapable of sorting, the sort will be carried out in process.
	IsSortSupported() bool

	// Returns the resource type that this provider serves. The resource type of the persistence provider must match,
	// or partially match, those the endpoint serves. For instance, a PersistenceProvider serving the User resource
	// should be loaded onto an Endpoint serving the User resource, or an Endpoint serving at the root of all resources.
	ResourceType() *core.ResourceType

	// Returns the number of total resources managed by this provider. Caller may use this to estimate the amount
	// of workload and decide whether to proceed with the request. Caution that this result does not necessarily reflect
	// the actual status of the underlying data store, due to concurrency.
	Total(ctx context.Context) (int64, error)

	// Returns the number of managed resources that satisfies the given SCIM filter. Caller may use this to estimate
	// the amount of workload and decide whether to proceed with the request. Caution that this result does not necessarily
	// reflect the actual status of the underlying data store, due to concurrency. If the provider does not support
	// filter (as in IsFilterSupported() == false), it can default to return the result of Total().
	Count(ctx context.Context, scimFilter string) (int64, error)

	// Insert a single resource into the underling data store and returns any error. The data store is discouraged to
	// make any alterations to the resource. Although SCIM protocol allows alterations be done to user submitted resources,
	// such alterations are usually done by the service provider, not the persistence provider.
	InsertOne(ctx context.Context, resource *core.Resource) error

	// Retrieve a resource by its id from the provider. The id refers to the SCIM resource id as is defined with the core
	// schema. If error is non-nil, resource is nil.
	GetById(ctx context.Context, id string) (*core.Resource, error)

	// more to come...
}