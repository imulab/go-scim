package filters

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
)

func NewReadOnlyClearFieldFilter(order int) protocol.FieldFilter {
	return &readOnlyClearFieldFilter{order: order}
}

type readOnlyClearFieldFilter struct {
	order 	int
}

func (f *readOnlyClearFieldFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Mutability() == core.MutabilityReadOnly
}

func (f *readOnlyClearFieldFilter) Order() int {
	return f.order
}

func (f *readOnlyClearFieldFilter) Filter(ctx *protocol.FieldFilterContext, resource *prop.Resource, property core.Property) error {
	_, err := property.Delete()
	return err
}

func (f *readOnlyClearFieldFilter) FieldRef(ctx *protocol.FieldFilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	_, err := property.Delete()
	return err
}

