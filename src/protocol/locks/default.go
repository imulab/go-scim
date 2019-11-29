package locks

import (
	"context"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	"sync"
)

// Return a default implementation of LockProvider. The implementation utilizes Golang's sync.Map or channel to sync
// goroutine access to the resource represented by its id.
func Default() protocol.LockProvider {
	return &defaultProvider{
		m: &sync.Map{},
	}
}

type defaultProvider struct {
	m *sync.Map
}

func (p *defaultProvider) Lock(ctx context.Context, resource *prop.Resource) error {
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

func (p *defaultProvider) Unlock(ctx context.Context, resource *prop.Resource) {
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
