package persistence

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/query"
	"sync"
)

// Create a new in memory persistence provider
func NewMemoryProvider() Provider {
	p := new(memoryProvider)
	p.database = make(map[string]map[string]*core.Resource)
	return p
}

// In memory implementation of the persistence provider, intended for testing and demonstration purposes, not for
// production use.
type memoryProvider struct {
	sync.RWMutex
	// Resources are indexed by first their resource type id, and then their id.
	database map[string]map[string]*core.Resource
}

func (p *memoryProvider) IsFilterSupported() bool {
	return true
}

func (p *memoryProvider) IsPaginationSupported() bool {
	return false
}

func (p *memoryProvider) IsSortSupported() bool {
	return false
}

func (p *memoryProvider) IsResourceTypeSupported(resourceType *core.ResourceType) bool {
	_, ok := p.database[resourceType.Id]
	return ok
}

func (p *memoryProvider) Total(ctx context.Context) (n int64, err error) {
	for _, resources := range p.database {
		n += int64(len(resources))
	}
	return
}

func (p *memoryProvider) Count(ctx context.Context, scimFilter string) (n int64, err error) {
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

func (p *memoryProvider) InsertOne(ctx context.Context, resource *core.Resource) error {
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

func (p *memoryProvider) GetById(ctx context.Context, id string) (*core.Resource, error) {
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
	_ Provider = (*memoryProvider)(nil)
)