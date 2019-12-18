package filter

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"strings"
)

// Create a ForResource filter that deletes the values in readOnly properties from a resource.
func ClearReadOnly() ForResource {
	return FromForProperty(&readOnlyClearFieldFilter{})
}

type readOnlyClearFieldFilter struct{}

func (f *readOnlyClearFieldFilter) Supports(attribute *spec.Attribute) bool {
	return attribute.Mutability() == spec.MutabilityReadOnly &&
		!strings.HasSuffix(attribute.ID(), "$elem")
}

func (f *readOnlyClearFieldFilter) Filter(ctx context.Context, resource *prop.Resource, property prop.Property) error {
	return property.Delete()
}

func (f *readOnlyClearFieldFilter) FieldRef(ctx context.Context, resource *prop.Resource, property prop.Property,
	refResource *prop.Resource, refProperty prop.Property) error {
	return property.Delete()
}

var (
	_ ForProperty = (*readOnlyClearFieldFilter)(nil)
)