package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
)

const (
	annotationSkipCanonical = "@canonical:skip"
)

// Create a canonical value filter. This filter is responsible for ensuring the provided values are among the defined
// canonicalValues in the attribute. The filter only works on string type where canonicalValues attribute is not empty.
// In addition, the filter does nothing if the property is unassigned, or if the attribute has been marked by the
// annotation '@canonical:skip'.
func NewCanonicalValueFilter() PropertyFilter {
	return &canonicalValueFilter{}
}

var (
	_ PropertyFilter = (*canonicalValueFilter)(nil)
)

type canonicalValueFilter struct {}

func (f *canonicalValueFilter) Supports(attribute *core.Attribute) bool {
	return !attribute.MultiValued &&
		attribute.Type == core.TypeString &&
		len(attribute.CanonicalValues) > 0 &&
		!containsAnnotation(attribute, annotationSkipCanonical)
}

func (f *canonicalValueFilter) Order(attribute *core.Attribute) int {
	return 600
}

func (f *canonicalValueFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.canonical(property)
}

func (f *canonicalValueFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return f.canonical(property)
}

func (f *canonicalValueFilter) canonical(property core.Property) error {
	if property.IsUnassigned() {
		return nil
	}
	var attribute = property.Attribute()
	for _, canonicalValue := range attribute.CanonicalValues {
		if property.(core.EqualAware).IsEqualTo(canonicalValue) {
			return nil
		}
	}
	return core.Errors.InvalidValue("'%s' does not satisfy canonical values constraint", attribute.DisplayName())
}