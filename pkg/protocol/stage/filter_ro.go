package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
)

const (
	annotationSkipReadOnlyClear = "@readOnly:noClean"
	annotationSkipReadOnlyCopy  = "@readOnly:noCopy"
)

// Create a new read only filter. The filter is responsible of handling any attribute whose mutability is set to read only.
// The filter has two major responsibilities: clear the value present in the read only property, and copy the reference value
// to the read only property. Both responsibilities may be skipped using annotation '@readOnly:noClean' and '@readOnly:noCopy'
// respectively. In FilterOnCreate, the filter attempts to clear the property value. In FilterOnUpdate, the filter attempts
// to clear the property value and then copy the reference value, if one exists.
//
// It is sensible to skip this filter on the complex property and let the sub properties do the work, provided they are all
// read only.
func NewReadOnlyFilter(order int) PropertyFilter {
	return &readOnlyFilter{order: order}
}

var _ PropertyFilter = (*readOnlyFilter)(nil)

type readOnlyFilter struct{ order int }

func (f *readOnlyFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Mutability == core.MutabilityReadOnly
}

func (f *readOnlyFilter) Order() int {
	return f.order
}

func (f *readOnlyFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	if !containsAnnotation(property.Attribute(), annotationSkipReadOnlyClear) {
		return property.(core.Crud).Delete(nil)
	}

	return nil
}

func (f *readOnlyFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	if !containsAnnotation(property.Attribute(), annotationSkipReadOnlyClear) {
		err := property.(core.Crud).Delete(nil)
		if err != nil {
			return err
		}
	}

	if !containsAnnotation(property.Attribute(), annotationSkipReadOnlyCopy) {
		if refProp != nil {
			err := property.(core.Crud).Replace(nil, refProp.Raw())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
