package db

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"sync"
)

// Memory return a new memory implementation of DB. This implementation saves resources in memory. Although
// it does allow for concurrent access through the use of RWMutex, it does not support high throughput usage.
// Hence, it is only intended for testing and showcasing purposes. This implementation also ignores all the field projection
// parameters that it always returned the full resource regardless of the request to include or exclude attributes.
func Memory() DB {
	db := memoryDB{
		RWMutex: sync.RWMutex{},
		db:      make(map[string]*prop.Resource),
	}
	return &db
}

type memoryDB struct {
	sync.RWMutex
	db map[string]*prop.Resource
}

func (m *memoryDB) Insert(_ context.Context, resource *prop.Resource) error {
	id := resource.IdOrEmpty()
	if len(id) == 0 {
		return fmt.Errorf("%w: empty id", spec.ErrInternal)
	}

	if _, ok := m.db[id]; ok {
		return fmt.Errorf("%w: id exists", spec.ErrInvalidValue)
	}

	m.Lock()
	defer m.Unlock()
	m.db[id] = resource

	return nil
}

func (m *memoryDB) Get(_ context.Context, id string, _ *crud.Projection) (*prop.Resource, error) {
	r, ok := m.db[id]
	if !ok {
		return nil, fmt.Errorf("%w: resource not found by id", spec.ErrNotFound)
	}
	return r, nil
}

func (m *memoryDB) Count(_ context.Context, filter string) (int, error) {
	if len(filter) == 0 {
		return len(m.db), nil
	}

	n := 0
	for _, r := range m.db {
		ok, _ := crud.Evaluate(r, filter)
		if ok {
			n++
		}
	}
	return n, nil
}

func (m *memoryDB) Replace(_ context.Context, ref *prop.Resource, replacement *prop.Resource) error {
	id := ref.IdOrEmpty()
	_, ok := m.db[id]
	if !ok {
		return fmt.Errorf("%w: resource not found by id", spec.ErrNotFound)
	}

	version := ref.MetaVersionOrEmpty()
	if len(version) > 0 && m.db[id].MetaVersionOrEmpty() != version {
		return spec.ErrConflict
	}

	m.db[id] = replacement
	return nil
}

func (m *memoryDB) Delete(_ context.Context, resource *prop.Resource) error {
	delete(m.db, resource.IdOrEmpty())
	return nil
}

func (m *memoryDB) Query(_ context.Context, filter string, sort *crud.Sort, pagination *crud.Pagination, _ *crud.Projection) ([]*prop.Resource, error) {
	var candidates = make([]*prop.Resource, 0)
	for _, r := range m.db {
		if ok, _ := crud.Evaluate(r, filter); ok {
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
