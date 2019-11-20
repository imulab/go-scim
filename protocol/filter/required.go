package filter

import (
	"context"
	"github.com/imulab/go-scim/core"
)

type (
	// An implementation of PropertyFilter to check the property satisfies the required attribute. The filter
	// only supports properties with an actual required attribute, and it is suggested to be placed towards the
	// back of all filters to account for any copy and value generation on otherwise unassigned properties.
	requiredFilter struct {}
)

func (f *requiredFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Required
}

func (f *requiredFilter) Order(attribute *core.Attribute) int {
	return 500
}

func (f *requiredFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource) error {
	if !property.IsUnassigned() {
		return nil
	}
	return core.Errors.InvalidValue("'%s' is required, but is unassigned", property.Attribute().DisplayName())
}

//// Create a new required filter
//func NewRequired() PropertyFilter {
//	return &requiredFilter{}
//}