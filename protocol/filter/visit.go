package filter

import (
	"context"
	"github.com/imulab/go-scim/core"
)

type (
	// Function signature that runs the given resource through all necessary filters, with the help of a reference
	// resource, and return any error
	ExecutorWithRef func(ctx context.Context, resource *core.Resource, ref *core.Resource) error

	// Implementation of core.Visitor that visits all properties within a resource, trying to invoke all property
	// filters assigned to it. This visitor can also be equipped with a reference navigator, which navigates through
	// a reference resource, to keep a reference property in async with the currently visited property for reference
	// and comparison purposes.
	filterVisitor struct {
		// The context to use for all filters. If nil, a context.Background() will be used.
		context context.Context
		// The resource currently being visited. This field must be set.
		resource *core.Resource
		// The reference currently serving as a reference to the resource being visited. This
		// field is optional. If nil, it means the resource is being visited without reference.
		// This will impact the behaviour of the filter.
		ref *core.Resource
		// Map of attribute ids to all property filters handling the attribute. The filters for an attribute
		// will be invoked in order when the property with that attribute is visited, along with the reference
		// property, if necessary. Any error from the filter will results in the visitor's premature exit.
		filters map[string][]PropertyFilter
	}
)

// Construct and return a filter executor with reference to run during the filter stage.
func NewExecutorWithRef(resourceTypes []*core.ResourceType, filters []PropertyFilter) ExecutorWithRef {
	index := BuildIndex(resourceTypes, filters)
	return func(ctx context.Context, resource *core.Resource, ref *core.Resource) error {
		visitor := &filterVisitor{
			context:  ctx,
			resource: resource,
			ref:      ref,
			filters:  index,
		}
		return resource.Visit(visitor)
	}
}

func (v *filterVisitor) ShouldVisit(property core.Property) bool {
	return true
}

func (v *filterVisitor) Visit(property core.Property) error {
	filters, ok := v.filters[property.Attribute().Id]
	if !ok || len(filters) == 0 {
		return nil
	}

	var ctx context.Context
	{
		if v.context != nil {
			ctx = v.context
		} else {
			ctx = context.Background()
		}
	}

	for _, filter := range filters {
		err := filter.Filter(ctx, v.resource, property, v.ref)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *filterVisitor) BeginComplex(complex core.Property) {}

func (v *filterVisitor) EndComplex(complex core.Property) {}

func (v *filterVisitor) BeginMulti(multi core.Property) {}

func (v *filterVisitor) EndMulti(multi core.Property) {}
