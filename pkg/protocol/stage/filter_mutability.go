package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
)

const annotationSkipMutability = "@mutability:skip"

// Create a mutability filter. The filter is responsible for scanning properties with readOnly or immutable attributes.
// However, readOnly attributes can be skipped by marking it with annotation '@skipReadOnly'. In general, property values
// with readOnly attribute is deleted in absence of a reference property, while they are overwritten when in presence of
// a reference property; on the other hand, property values with an immutable attribute is ignored in absence of a reference
// property, while in the presence of a reference property, they are matched with the reference property value to ensure
// values have not changed.
func NewMutabilityFilter(order int) PropertyFilter {
	return &mutabilityFilter{order: order}
}

var _ PropertyFilter = (*mutabilityFilter)(nil)

type mutabilityFilter struct{ order int }

func (f *mutabilityFilter) Supports(attribute *core.Attribute) bool {
	// We do not have to worry about read only attributes because server is allowed
	// to modify those values. The part where modified values didn't come from the
	// user is ensured by a preceding readOnlyFilter.
	return attribute.Mutability == core.MutabilityImmutable &&
		!containsAnnotation(attribute, annotationSkipMutability)
}

func (f *mutabilityFilter) Order() int {
	return f.order
}

func (f *mutabilityFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return nil
}

func (f *mutabilityFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	if ref == nil || refProp == nil || refProp.IsUnassigned() {
		// no reference or reference does not contain a value, then
		// property is free to have any value, by definition of immutability
		return nil
	}

	if property.(core.EqualAware).Matches(refProp) {
		return nil
	}

	return core.Errors.Mutability("'%s' is immutable, but value has changed", property.Attribute().DisplayName())
}
