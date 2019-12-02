package filters

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	"strings"
)

func NewCopyReadOnlyResourceFilter(resourceType *core.ResourceType, order int) protocol.ResourceFilter {
	return NewResourceFieldFilterOf(resourceType, []protocol.FieldFilter{
		NewCopyReadOnlyFieldFilter(0),
	}, order)
}

func NewCopyReadOnlyFieldFilter(order int) protocol.FieldFilter {
	return &copyReadOnlyFieldFilter{order: order}
}

// This filter copies value from the reference property to the resource property when the mutability is readOnly and
// the attribute is marked with annotation '@CopyReadOnly'. It is suggested that: for simple singular attribute, mark
// on the attribute itself; for complex singular attribute, mark on the sub attributes instead; for multiValued attribute,
// mark on the attribute itself.
type copyReadOnlyFieldFilter struct {
	order int
}

func (f *copyReadOnlyFieldFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Mutability() == core.MutabilityReadOnly &&
		attribute.HasAnnotation("@CopyReadOnly") &&
		!strings.HasSuffix(attribute.ID(), "$elem") // skip over the derived element attribute.
}

func (f *copyReadOnlyFieldFilter) Order() int {
	return f.order
}

func (f *copyReadOnlyFieldFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property) error {
	return nil
}

func (f *copyReadOnlyFieldFilter) FieldRef(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	if refProperty == nil {
		return nil
	}
	return property.Replace(refProperty.Raw())
}
