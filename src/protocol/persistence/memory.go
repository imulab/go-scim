package persistence

import (
	"context"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	"sync"
)

// Return a new memory implementation of PersistenceProvider. This implementation saves resources in memory. Although
// it does allow for concurrent access through the use of RWMutex, it does not however allow for high throughput usage.
// It is intended for testing and showcasing purposes only.
func Memory() protocol.PersistenceProvider {
	return &memoryProvider{
		RWMutex: sync.RWMutex{},
		db:      make(map[string]*prop.Resource),
	}
}

type memoryProvider struct {
	sync.RWMutex
	db	map[string]*prop.Resource
}

func (m *memoryProvider) Insert(ctx context.Context, resource *prop.Resource) error {
	id := resource.ID()
	if len(id) == 0 {
		return errors.Internal("cannot save resource with empty id")
	}

	if _, ok := m.db[id]; ok {
		return errors.Internal("resource with id '%s' already exists", id)
	}

	m.Lock()
	defer m.Unlock()
	m.db[id] = resource

	return nil
}

func (m *memoryProvider) Get(ctx context.Context, id string) (*prop.Resource, error) {
	r, ok := m.db[id]
	if !ok {
		return nil, errors.NotFound("resource by id [%s] is not found", id)
	}
	return r, nil
}

func (m *memoryProvider) Count(ctx context.Context, filter string) (int, error) {
	// todo
	return 0, nil
}

func (m *memoryProvider) Replace(ctx context.Context, resource *prop.Resource) error {
	id := resource.ID()
	_, ok := m.db[id]
	if !ok {
		return errors.NotFound("resource by id [%s] is not found", id)
	}
	m.db[id] = resource
	return nil
}