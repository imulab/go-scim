package event

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
)

// Return a composite publisher which invokes the given publishers one by one on events.
func Of(publishers ...Publisher) Publisher {
	return &composite{publishers: publishers}
}

type composite struct {
	publishers []Publisher
}

func (c *composite) ResourceCreated(ctx context.Context, created *prop.Resource) {
	for _, p := range c.publishers {
		p.ResourceCreated(ctx, created)
	}
}

func (c *composite) ResourceUpdated(ctx context.Context, updated *prop.Resource, original *prop.Resource) {
	for _, p := range c.publishers {
		p.ResourceUpdated(ctx, updated, original)
	}
}

func (c *composite) ResourceDeleted(ctx context.Context, deleted *prop.Resource) {
	for _, p := range c.publishers {
		p.ResourceDeleted(ctx, deleted)
	}
}
