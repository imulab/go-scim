package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
)

// A property filter is the main processing stage that the resource go through after being parsed and before being
// sent to a persistence provider. The implementations can carry out works like annotation processing, validation,
// value generation, etc.
type PropertyFilter interface {
	// Returns true if this filter supports processing the given attribute.
	Supports(attribute *core.Attribute) bool
	// Returns an integer based order value, so that filters can be sorted to be visited sequentially.
	Order() int
	// Process the given property during resource creation, with access to the owning resource.
	FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error
	// Process the given property during resource update, with access to the owning and reference resource, and property.
	FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error
}

// Return true if the attribute's metadata contains the queried annotation. The annotation is case sensitive.
func containsAnnotation(attr *core.Attribute, annotation string) bool {
	metadata := core.Meta.Get(attr.Id, core.DefaultMetadataId)
	if metadata == nil {
		return false
	}
	annotations := metadata.(*core.DefaultMetadata).Annotations
	for _, each := range annotations {
		if each == annotation {
			return true
		}
	}
	return false
}
