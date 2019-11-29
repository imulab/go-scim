package locks

import (
	"context"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
)

// Return a no-op implementation of LockProvider.
func NoOp() protocol.LockProvider {
	return noLock{}
}

type noLock struct {}

func (n noLock) Lock(ctx context.Context, resource *prop.Resource) error {
	return nil
}

func (n noLock) Unlock(ctx context.Context, resource *prop.Resource) {}

