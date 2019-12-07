package lock

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"sync"
)

// Return a default implementation of Lock.
func Default() Lock {
	return &defaultLocker{
		m: &sync.Map{},
	}
}

// A default lock implementation that uses channels to pass tokens among requesting Goroutines in order to
// synchronize access. The lock's fairness solely depends on that of a channel, and is generally not fair.
// The synchronizing channels are stored in a sync.Map which is indexed by the resource's ID. Hence, the resource
// must have its ID set before calling Lock, otherwise, the lock will return error.
type defaultLocker struct {
	m *sync.Map
}

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
