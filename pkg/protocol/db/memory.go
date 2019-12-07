package db

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"sync"
)

// Return a new memory implementation of Database. This implementation saves resources in memory. Although
// it does allow for concurrent access through the use of RWMutex, it does not allow for high throughput usage.
// It is intended for testing and showcasing purposes only.
func Memory() DB {
	return &memoryDB{
		RWMutex: sync.RWMutex{},
		db:      make(map[string]*prop.Resource),
	}
}

type memoryDB struct {
	sync.RWMutex
	db map[string]*prop.Resource
}

func (m *memoryDB) Insert(ctx context.Context, resource *prop.Resource) error {
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

func (m *memoryDB) Get(ctx context.Context, id string) (*prop.Resource, error) {
	r, ok := m.db[id]
	if !ok {
		return nil, errors.NotFound("resource by id [%s] is not found", id)
	}
	return r, nil
}

func (m *memoryDB) Count(ctx context.Context, filter string) (int, error) {
	// todo
	return 0, nil
}

func (m *memoryDB) Replace(ctx context.Context, resource *prop.Resource) error {
	id := resource.ID()
	_, ok := m.db[id]
	if !ok {
		return errors.NotFound("resource by id [%s] is not found", id)
	}
	m.db[id] = resource
	return nil
}

func (m *memoryDB) Delete(ctx context.Context, id string) error {
	delete(m.db, id)
	return nil
}
