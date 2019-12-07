package lock

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/prop"
)

// Return a no-op implementation of Lock.
func None() Lock {
	return noOpLocker{}
}

type noOpLocker struct{}

func (n noOpLocker) Lock(ctx context.Context, resource *prop.Resource) error {
	return nil
}

func (n noOpLocker) Unlock(ctx context.Context, resource *prop.Resource) {}
