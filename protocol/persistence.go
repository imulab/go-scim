package protocol

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/query"
	"sync"
)

type (
	// Interface for providing persistence services. This interface is the central point of all
	// state persistence and management for the endpoints.
	PersistenceProvider interface {
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

		// Returns true if this provider supports the resource type.
		IsResourceTypeSupported(resourceType *core.ResourceType) bool

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

	// In memory implementation of the persistence provider, intended for testing and demonstration purposes, not for
	// production use.
	MemoryPersistenceProvider struct {
		sync.RWMutex
		// Resources are indexed by first their resource type id, and then their id.
		database map[string]map[string]*core.Resource
	}
)

func (p *MemoryPersistenceProvider) IsFilterSupported() bool {
	return true
}

func (p *MemoryPersistenceProvider) IsPaginationSupported() bool {
	return false
}

func (p *MemoryPersistenceProvider) IsSortSupported() bool {
	return false
}

func (p *MemoryPersistenceProvider) IsResourceTypeSupported(resourceType *core.ResourceType) bool {
	_, ok := p.database[resourceType.Id]
	return ok
}

func (p *MemoryPersistenceProvider) Total(ctx context.Context) (n int64, err error) {
	for _, resources := range p.database {
		n += int64(len(resources))
	}
	return
}

func (p *MemoryPersistenceProvider) Count(ctx context.Context, scimFilter string) (n int64, err error) {
	var root *core.Step
	{
		root, err = query.CompileFilter(scimFilter)
		if err != nil {
			return
		}
	}

	for _, resources := range p.database {
		for _, resource := range resources {
			if ok, err := resource.Evaluate(root); err != nil {
				return 0, err
			} else if ok {
				n += 1
			}
		}
	}

	return
}

func (p *MemoryPersistenceProvider) InsertOne(ctx context.Context, resource *core.Resource) error {
	id, err := resource.GetID()
	if err != nil {
		return err
	}

	var m map[string]*core.Resource
	m = p.database[resource.GetResourceType().Id]
	if m == nil {
		m = make(map[string]*core.Resource)
	}
	m[id] = resource
	p.database[resource.GetResourceType().Id] = m

	return nil
}

func (p *MemoryPersistenceProvider) GetById(ctx context.Context, id string) (*core.Resource, error) {
	for _, resources := range p.database {
		resource, ok := resources[id]
		if ok {
			return resource, nil
		}
	}
	return nil, core.Errors.NotFound("resource by id %s is not found", id)
}

// implementation check
var (
	_ PersistenceProvider = (*MemoryPersistenceProvider)(nil)
)