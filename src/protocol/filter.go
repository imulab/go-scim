package protocol

import (
	"context"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
)

type (
	// ResourceFilter is responsible of carrying out operations on a resource like validation, modification, etc.
	ResourceFilter interface {
		// Return an integer to determine this filter's place among other filter.
		// Multiple filters will be sorted based on the order value in ascending
		// order and executed sequentially.
		Order() int
		// Filter the resource and return any error. If the error returned is not nil,
		// the caller should immediately abort the operation and avoid executing the
		// following filters.
		Filter(ctx *FilterContext, resource *prop.Resource) error
		// Filter the resource with reference to a reference resource and return any error.
		// The reference resource may serve as a guidance for the expected state of the resource.
		// If the error returned is not nil, the caller should immediately abort the operation
		// and avoid executing the following filters.
		FilterRef(ctx *FilterContext, resource *prop.Resource, ref *prop.Resource) error
	}
	// FieldFilter is responsible of carrying out operations on a single property field. It provides more
	// granular control than a resource filter, and is the core of the default resource filter, which simply
	// traverses the resource and invokes the field filter.
	FieldFilter interface {
		// Returns true if this filter supports the supplied attribute. The Filter method
		// will only be called when this method returns true. This method is expected to be
		// called at setup time.
		Supports(attribute *core.Attribute) bool
		// Returns an integer to determine this filter's place among others. Multiple filters
		// on the same attribute will be sorted based on the order value in ascending order.
		// This method is expected to be called at setup time.
		Order() int
		// Filter the given property with reference to the resource that contains this property.
		// Any error returned shall cause the caller to abort subsequent operations.
		Filter(ctx *FilterContext, resource *prop.Resource, property core.Property) error
		// Filter the given property with reference to the resource that contains this property, another reference resource which
		// potentially holds a reference property. The reference resource and property may serve as a guidance for the expected
		// state of the property. The reference resource shall never be nil, whereas the reference property may be nil.
		// Any error returned shall cause the caller to abort subsequent operations.
		FieldRef(ctx *FilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error
	}
	// A shared context among filters.
	FilterContext struct {
		requestContext context.Context
		data           map[interface{}]interface{}
	}
)

// Create a new filter context.
func NewFilterContext(ctx context.Context) *FilterContext {
	return &FilterContext{
		requestContext: ctx,
		data:           make(map[interface{}]interface{}),
	}
}

func (c *FilterContext) RequestContext() context.Context {
	return c.requestContext
}

func (c *FilterContext) Get(key interface{}) (value interface{}, ok bool) {
	value, ok = c.data[key]
	return
}

func (c *FilterContext) Put(key, value interface{}) {
	c.data[key] = value
}
