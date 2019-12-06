package protocol

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"sync"
)

// Return a default implementation of LockProvider.
func DefaultLock() LockProvider {
	return &defaultLocker{
		m: &sync.Map{},
	}
}

// Return a no-op implementation of LockProvider.
func NoOpLock() LockProvider {
	return noOpLocker{}
}

type (
	// Lock provider provides a set of methods to allow callers obtain exclusive modification rights
	// to a resource. This protects the underlying resource from concurrent modification failures.
	// The provider does not specify lock fairness, it is up to the implementations. The implementations
	// are also encouraged to implement an auto-unlock mechanism which kicks in after a lock has been
	// held on for long periods of time, so that to avoid the situation where the process with a lock
	// crashed and failed to unlock.
	LockProvider interface {
		// Try to obtain a lock on the given resource represented by the id. The method blocks indefinitely
		// if some other process is currently holding the lock and context does not specify a timeout period.
		// If lock failed to obtain within the contextual time limits, an error will be returned.
		Lock(ctx context.Context, resource *prop.Resource) error
		// Return the held lock. In any situation, the method should return immediately.
		Unlock(ctx context.Context, resource *prop.Resource)
	}
	// A default lock implementation that uses channels to pass tokens among requesting Goroutines in order to
	// synchronize access. The lock's fairness solely depends on that of a channel, and is generally not fair.
	// The synchronizing channels are stored in a sync.Map which is indexed by the resource's ID. Hence, the resource
	// must have its ID set before calling Lock, otherwise, the lock will return error.
	defaultLocker struct {
		m *sync.Map
	}
	// no op locker
	noOpLocker struct{}
)

func (p *defaultLocker) Lock(ctx context.Context, resource *prop.Resource) error {
	id := resource.ID()
	if len(id) == 0 {
		return errors.Internal("Cannot obtain lock for resource without ID")
	}

	var sig chan struct{}
	{
		// try a clean load first
		v, ok := p.m.Load(id)
		if ok {
			sig = v.(chan struct{})
		} else {
			// load or store, if clean load didn't have a hit
			l := make(chan struct{}, 1)
			l <- struct{}{}
			v, _ = p.m.LoadOrStore(id, l)
			sig = v.(chan struct{})
		}
	}

	select {
	case <-sig:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *defaultLocker) Unlock(ctx context.Context, resource *prop.Resource) {
	id := resource.ID()
	if len(id) == 0 {
		return
	}

	var sig chan struct{}
	{
		v, ok := p.m.Load(id)
		if !ok {
			return
		}
		sig = v.(chan struct{})
	}

	if len(sig) == 0 {
		// only return when no tokens in channel
		sig <- struct{}{}
	}
}

func (n noOpLocker) Lock(ctx context.Context, resource *prop.Resource) error {
	return nil
}

func (n noOpLocker) Unlock(ctx context.Context, resource *prop.Resource) {}
