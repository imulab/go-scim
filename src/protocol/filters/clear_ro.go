package filters

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
)

// Create a resource filter that deletes the values in readOnly properties from a resource.
func NewClearReadOnlyResourceFilter(resourceType *core.ResourceType, order int) protocol.ResourceFilter {
	return NewResourceFieldFilterOf(resourceType, []protocol.FieldFilter{
		&readOnlyClearFieldFilter{order: 0},
	}, order)
}

type readOnlyClearFieldFilter struct {
	order int
}

func (f *readOnlyClearFieldFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Mutability() == core.MutabilityReadOnly
}

func (f *readOnlyClearFieldFilter) Order() int {
	return f.order
}

func (f *readOnlyClearFieldFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property) error {
	return property.Delete()
}

func (f *readOnlyClearFieldFilter) FieldRef(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	return property.Delete()
}
