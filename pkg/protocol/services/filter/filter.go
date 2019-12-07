package filter

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
)

// Responsible of filtering a resource, like carrying out operations on a resource such as
// validation, modification, etc.
type ForResource interface {
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
type ForProperty interface {
	// Returns true if this filter supports the supplied attribute. The Filter method
	// will only be called when this method returns true. This method is expected to be
	// called at setup time.
	Supports(attribute *spec.Attribute) bool
	// Filter the given property with reference to the resource that contains this property.
	// Any error returned shall cause the caller to abort subsequent operations.
	Filter(ctx context.Context, resource *prop.Resource, property prop.Property) error
	// Filter the given property with reference to the resource that contains this property, another reference resource which
	// potentially holds a reference property. The reference resource and property may serve as a guidance for the expected
	// state of the property. The reference resource shall never be nil, whereas the reference property may be nil.
	// Any error returned shall cause the caller to abort subsequent operations.
	FieldRef(ctx context.Context, resource *prop.Resource, property prop.Property, refResource *prop.Resource, refProperty prop.Property) error
}

// Create a ForResource filter using a list of ForProperty filters. The filtering resource will be traversed
// by a special prop.Visitor implementation which invokes each ForProperty filter in sequence on every visited
// property.
func FromForProperty(filters ...ForProperty) ForResource {
	return &traversingResourceFilter{filters: filters}
}

// Special resource filter to adapt the use of ForProperty to ForResource.
type traversingResourceFilter struct {
	filters []ForProperty
}

func (f *traversingResourceFilter) Filter(ctx context.Context, resource *prop.Resource) error {
	v := &syncPropertyVisitor{
		ctx:      ctx,
		resource: resource,
		filters:  f.filters,
		stack:    make([]*frame, 0),
	}
	return resource.Visit(v)
}

func (f *traversingResourceFilter) FilterRef(ctx context.Context, resource *prop.Resource, ref *prop.Resource) error {
	if ref == nil {
		return f.Filter(ctx, resource)
	}
	v := &syncPropertyVisitor{
		ctx:      ctx,
		resource: resource,
		ref:      ref,
		refNav:   ref.NewNavigator(),
		filters:  f.filters,
		stack:    make([]*frame, 0),
	}
	return resource.Visit(v)
}

var (
	_ ForResource = (*traversingResourceFilter)(nil)
)