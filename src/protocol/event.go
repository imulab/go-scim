package protocol

import (
	"context"
	"github.com/imulab/go-scim/src/core/prop"
)

// EventPublisher provides hooks for implementations to react to certain events during protocol execution.
// Implementations MUST NOT panic and MUST NOT block the caller's Goroutine.
type EventPublisher interface {
	// Notify that a new resource has been created.
	ResourceCreated(ctx context.Context, created *prop.Resource)
}
