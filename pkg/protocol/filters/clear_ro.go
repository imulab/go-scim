package filters

import (
	"github.com/imulab/go-scim/pkg/core"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol"
	"strings"
)

// Create a resource filter that deletes the values in readOnly properties from a resource.
func NewClearReadOnlyResourceFilter() protocol.ResourceFilter {
	return NewResourceFieldFilterOf(NewClearReadOnlyFieldFilter())
}

func NewClearReadOnlyFieldFilter() protocol.FieldFilter {
	return &readOnlyClearFieldFilter{}
}

type readOnlyClearFieldFilter struct{}

func (f *readOnlyClearFieldFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Mutability() == core.MutabilityReadOnly &&
		!strings.HasSuffix(attribute.ID(), "$elem")
}

func (f *readOnlyClearFieldFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property) error {
	return property.Delete()
}

func (f *readOnlyClearFieldFilter) FieldRef(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	return property.Delete()
}
