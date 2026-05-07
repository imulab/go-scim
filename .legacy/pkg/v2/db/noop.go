package db

import (
	"context"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/prop"
)

// NoOp return an no op implementation of DB. This implementation does nothing and always returns nil error. For Count
// method, it returns 0 as count; for Get method, it returns nil resource; For Query method, it returns empty slice as
// results. This implementation might be useful when implementing use cases where resource does not require persistence.
func NoOp() DB {
	return noOpDB{}
}

type noOpDB struct{}

func (_ noOpDB) Insert(_ context.Context, _ *prop.Resource) error {
	return nil
}

func (_ noOpDB) Count(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (_ noOpDB) Get(_ context.Context, _ string, _ *crud.Projection) (*prop.Resource, error) {
	return nil, nil
}

func (_ noOpDB) Replace(_ context.Context, _ *prop.Resource, _ *prop.Resource) error {
	return nil
}

func (_ noOpDB) Delete(_ context.Context, _ *prop.Resource) error {
	return nil
}

func (_ noOpDB) Query(_ context.Context, _ string, _ *crud.Sort, _ *crud.Pagination, _ *crud.Projection) ([]*prop.Resource, error) {
	return []*prop.Resource{}, nil
}
