package filter

import (
	"context"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
)

// Responsible of filtering a resource, like carrying out operations on a resource such as validation, modification, etc.
type ByResource interface {
	// Filter the resource and return any error. If the error returned is not nil,
	// the caller should immediately abort the operation and avoid executing the
	// following filters.
	Filter(ctx context.Context, resource *prop.Resource) error
	// Filter the resource with reference to a reference resource and return any error.
	// The reference resource may serve as a guidance for the expected state of the resource.
	// If the error returned is not nil, the caller should immediately abort the operation
	// and avoid executing the following filters.
	FilterRef(ctx context.Context, resource *prop.Resource, ref *prop.Resource) error
}

// Responsible of filtering on the level of a single property field. It provides more granular control than
// a resource filter.
type ByProperty interface {
	// Returns true if this filter supports the supplied attribute. The Filter method
	// will only be called when this method returns true. This method is expected to be
	// called at setup time.
	Supports(attribute *spec.Attribute) bool
	// Filter the given property with reference to the resource that contains this property. The property being filter
	// is the Current property on the navigator. Any error returned shall cause the caller to abort subsequent operations.
	Filter(ctx context.Context, resourceType *spec.ResourceType, nav prop.Navigator) error
	// Filter the given property with reference to the resource that contains this property, another reference resource which
	// potentially holds a reference property. The reference resource and property may serve as a guidance for the expected
	// state of the property. The reference resource shall never be nil, whereas the reference property may be nil.
	// Any error returned shall cause the caller to abort subsequent operations.
	FilterRef(ctx context.Context, resourceType *spec.ResourceType, nav prop.Navigator, refNav prop.Navigator) error
}
