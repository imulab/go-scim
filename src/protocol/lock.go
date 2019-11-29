package protocol

import (
	"context"
	"github.com/imulab/go-scim/src/core/prop"
)

// Lock provider provides a set of methods to allow callers obtain exclusive modification rights
// to a resource. This protects the underlying resource from concurrent modification failures.
// The provider does not specify lock fairness, it is up to the implementations. The implementations
// are also encouraged to implement an auto-unlock mechanism which kicks in after a lock has been
// held on for long periods of time, so that to avoid the situation where the process with a lock
// crashed and failed to unlock.
type LockProvider interface {
	// Try to obtain a lock on the given resource represented by the id. The method blocks indefinitely
	// if some other process is currently holding the lock and context does not specify a timeout period.
	// If lock failed to obtain within the contextual time limits, an error will be returned.
	Lock(ctx context.Context, resource *prop.Resource) error
	// Return the held lock. In any situation, the method should return immediately.
	Unlock(ctx context.Context, resource *prop.Resource)
}
