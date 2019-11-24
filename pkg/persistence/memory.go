package persistence

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
	"github.com/imulab/go-scim/pkg/query"
	"sync"
)

// Create a new in memory persistence provider
func NewMemoryProvider(resourceType *core.ResourceType) Provider {
	p := new(memoryProvider)
	p.resourceType = resourceType
	p.database = make(map[string]*core.Resource)
	return p
}

// In memory implementation of the persistence provider, intended for testing and demonstration purposes, not for
// production use.
type memoryProvider struct {
	sync.RWMutex
	resourceType *core.ResourceType
	// Resources are indexed by id
	database map[string]*core.Resource
}

func (p *memoryProvider) SupportsFilter() bool {
	return true
}

func (p *memoryProvider) SupportsPagination() bool {
	return false
}

func (p *memoryProvider) SupportsSort() bool {
	return false
}

func (p *memoryProvider) ResourceType() *core.ResourceType {
	return p.resourceType
}

func (p *memoryProvider) Total(ctx context.Context) (int64, error) {
	p.RLock()
	defer p.RUnlock()
	return int64(len(p.database)), nil
}

func (p *memoryProvider) Count(ctx context.Context, scimFilter string) (n int64, err error) {
	p.RLock()
	defer p.RUnlock()

	var root *core.Step
	{
		root, err = query.CompileFilter(scimFilter)
		if err != nil {
			return
		}
	}

	for _, resource := range p.database {
		if ok, err := resource.Evaluate(root); err != nil {
			return 0, err
		} else if ok {
			n += 1
		}
	}

	return
}

func (p *memoryProvider) InsertOne(ctx context.Context, resource *core.Resource) error {
	p.Lock()
	defer p.Unlock()

	id, err := resource.GetID()
	if err != nil {
		return err
	}
	p.database[id] = resource

	return nil
}

func (p *memoryProvider) ReplaceOne(ctx context.Context, replacement *core.Resource) error {
	p.Lock()
	defer p.Unlock()

	id, err := replacement.GetID()
	if err != nil {
		return err
	}

	if _, ok := p.database[id]; !ok {
		return core.Errors.NotFound("cannot replace resource with id '%s': does not exist", id)
	}
	p.database[id] = replacement

	return nil
}

func (p *memoryProvider) GetById(ctx context.Context, id string) (*core.Resource, error) {
	p.RLock()
	defer p.RUnlock()

	resource, ok := p.database[id]
	if !ok {
		return nil, core.Errors.NotFound("resource by id %s is not found", id)
	}
	return resource, nil
}

// implementation check
var (
	_ Provider = (*memoryProvider)(nil)
)