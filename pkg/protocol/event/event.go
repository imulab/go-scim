package event

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/prop"
)

// Publisher provides hooks for implementations to react to certain events during protocol execution.
// Implementations MUST NOT panic and MUST NOT block the caller's Goroutine.
type Publisher interface {
	// Notify that a new resource has been created.
	ResourceCreated(ctx context.Context, created *prop.Resource)
	// Notify that an old resource has been updated
	ResourceUpdated(ctx context.Context, updated *prop.Resource, original *prop.Resource)
	// Notify that a resource has been deleted
	ResourceDeleted(ctx context.Context, deleted *prop.Resource)
}
