package db

import (
	"context"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/expr"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/crud"
	"sync"
)

// Return a new memory implementation of Database. This implementation saves resources in memory. Although
// it does allow for concurrent access through the use of RWMutex, it does not allow for high throughput usage.
// It is intended for testing and showcasing purposes only. This implementation also ignores all the field projection
// parameters that it always returned the full resource regardless of the request to include or exclude attributes.
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

func (m *memoryDB) Get(ctx context.Context, id string, _ *crud.Projection) (*prop.Resource, error) {
	r, ok := m.db[id]
	if !ok {
		return nil, errors.NotFound("resource by id [%s] is not found", id)
	}
	return r, nil
}

func (m *memoryDB) Count(ctx context.Context, filter string) (int, error) {
	if len(filter) == 0 {
		return len(m.db), nil
	}

	root, err := expr.CompileFilter(filter)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range m.db {
		ok, _ := crud.Evaluate(r.NewNavigator().Current(), root)
		if ok {
			n++
		}
	}
	return n, nil
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

func (m *memoryDB) Delete(ctx context.Context, resource *prop.Resource) error {
	delete(m.db, resource.ID())
	return nil
}

func (m *memoryDB) Query(ctx context.Context, filter string, sort *crud.Sort, pagination *crud.Pagination, _ *crud.Projection) ([]*prop.Resource, error) {
	root, err := expr.CompileFilter(filter)
	if err != nil {
		return nil, err
	}

	var candidates = make([]*prop.Resource, 0)
	for _, r := range m.db {
		if ok, _ := crud.Evaluate(r.NewNavigator().Current(), root); ok {
			candidates = append(candidates, r)
		}
	}
	if len(candidates) == 0 {
		return []*prop.Resource{}, nil
	}

	if sort != nil {
		if err := sort.Sort(candidates); err != nil {
			return nil, err
		}
	}

	if pagination != nil {
		lb := pagination.StartIndex - 1
		if lb < 0 {
			lb = 0
		}
		ub := pagination.StartIndex + pagination.Count - 1
		if ub > len(candidates) {
			ub = len(candidates)
		}
		candidates = candidates[lb:ub]
	}

	return candidates, nil
}
