package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
)

type (
	// Hook to be invoked immediately after endpoint has successfully parsed a resource from request payload.
	PostParseHook interface {
		ResourceHasBeenParsed(ctx context.Context, parsed *core.Resource)
	}
	// Hook to be invoked immediately before endpoint will attempt to persist the resource.
	PrePersistHook interface {
		ResourceWillBePersisted(ctx context.Context, resource *core.Resource)
	}
	// Hook to be invoked immediately after endpoint has successfully persisted a resource to database.
	PostPersistHook interface {
		ResourceHasBeenPersisted(ctx context.Context, resource *core.Resource)
	}
	// Do nothing hook that implements all the hooks above.
	noOpHook struct{}
)

// Create a new no op hook
func NewNoOpHook() *noOpHook {
	return &noOpHook{}
}

func (n noOpHook) ResourceHasBeenParsed(ctx context.Context, parsed *core.Resource) {}

func (n noOpHook) ResourceWillBePersisted(ctx context.Context, resource *core.Resource) {}

func (n noOpHook) ResourceHasBeenPersisted(ctx context.Context, resource *core.Resource) {}
