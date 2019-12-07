package filter

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/annotations"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"strings"
)

// Create a ForResource filter that copies value from the reference property to the resource property when the
// mutability is readOnly and the attribute is marked with annotation '@CopyReadOnly'. It is suggested that:
// - for simple singular attribute, mark on the attribute itself
// - for complex singular attribute, mark on the sub attributes instead
// - for multiValued attribute, mark on the attribute itself.
func CopyReadOnly() ForResource {
	return FromForProperty(&copyReadOnlyFieldFilter{})
}

type copyReadOnlyFieldFilter struct{}

func (f *copyReadOnlyFieldFilter) Supports(attribute *spec.Attribute) bool {
	return attribute.Mutability() == spec.MutabilityReadOnly &&
		attribute.HasAnnotation(annotations.CopyReadOnly) &&
		!strings.HasSuffix(attribute.ID(), "$elem") // skip over the derived element attribute.
}

func (f *copyReadOnlyFieldFilter) Filter(ctx context.Context, resource *prop.Resource, property prop.Property) error {
	return nil
}

func (f *copyReadOnlyFieldFilter) FieldRef(ctx context.Context, resource *prop.Resource, property prop.Property,
	refResource *prop.Resource, refProperty prop.Property) error {
	if refProperty == nil {
		return nil
	}
	return property.Replace(refProperty.Raw())
}

var (
	_ ForProperty = (*copyReadOnlyFieldFilter)(nil)
)