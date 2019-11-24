package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
)

const (
	annotationSkipMutability = "@mutability:skip"
)

// Create a mutability filter. The filter is responsible for scanning properties with readOnly or immutable attributes.
// However, readOnly attributes can be skipped by marking it with annotation '@skipReadOnly'. In general, property values
// with readOnly attribute is deleted in absence of a reference property, while they are overwritten when in presence of
// a reference property; on the other hand, property values with an immutable attribute is ignored in absence of a reference
// property, while in the presence of a reference property, they are matched with the reference property value to ensure
// values have not changed.
func NewMutabilityFilter() PropertyFilter {
	return &mutabilityFilter{}
}

var (
	_ PropertyFilter = (*mutabilityFilter)(nil)
)

type mutabilityFilter struct {}

func (f *mutabilityFilter) Supports(attribute *core.Attribute) bool {
	if containsAnnotation(attribute, annotationSkipMutability) {
		// Because this filter copies reference property values to resource property when mutability is readOnly,
		// the property handled by this filter may have already been handled before by the same filter running against
		// a container attribute. For example, the 'meta' attribute is normally readOnly, one does not wish to double
		// process all its sub attributes when the 'meta' attribute has already been copied. Hence, we suggest to
		// mark any readOnly sub properties whose container property is also readOnly with '@mutability:skip'. This way,
		// we avoid the double processing.
		return false
	}
	return attribute.Mutability == core.MutabilityReadOnly || attribute.Mutability == core.MutabilityImmutable
}

func (f *mutabilityFilter) Order() int {
	return 200
}

func (f *mutabilityFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	if property.Attribute().Mutability == core.MutabilityReadOnly {
		return property.(core.Crud).Delete(nil)
	}
	return nil
}

func (f *mutabilityFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	switch property.Attribute().Mutability {
	case core.MutabilityReadOnly:
		if ref == nil || refProp == nil {
			return property.(core.Crud).Delete(nil)
		}
		return property.(core.Crud).Replace(nil, refProp.Raw())
	case core.MutabilityImmutable:
		if ref == nil || refProp == nil || refProp.IsUnassigned() {
			// no reference or reference does not contain a value, then
			// property is free to have any value, given immutable.
			return nil
		}
		if !property.(core.EqualAware).Matches(refProp) {
			return core.Errors.Mutability("'%s' is immutable, but value has changed", property.Attribute().DisplayName())
		}
		return nil
	default:
		return nil
	}
}