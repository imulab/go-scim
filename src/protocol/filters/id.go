package filters

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	uuid "github.com/satori/go.uuid"
)

// Create a new field filter that handles the 'id' field. When Filter is called, this filter generates a new uuid and
// assigns it to the id field. When FilterRef is called, this filter copies the reference property value to the property,
// provided reference property is present and is not unassigned. It also puts the id in the filter context under key
// IDFieldFilterKey.
func NewIDFieldFilter() protocol.FieldFilter {
	return &idFieldFilter{}
}

type (
	idFieldFilter struct {}
	IDFieldFilterKey struct {}
)

func (f *idFieldFilter) Supports(attribute *core.Attribute) bool {
	return attribute.ID() == "id"
}

func (f *idFieldFilter) Order() int {
	panic("implement me")
}

func (f *idFieldFilter) Filter(ctx *protocol.FieldFilterContext, resource *prop.Resource, property core.Property) error {
	id := uuid.NewV4().String()

	_, err := property.Replace(id)
	if err != nil {
		return err
	}
	ctx.Put(IDFieldFilterKey{}, id)

	return nil
}

func (f *idFieldFilter) FieldRef(ctx *protocol.FieldFilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	if refProperty == nil || refProperty.IsUnassigned() {
		return nil
	}

	id := refProperty.Raw()
	_, err := property.Replace(id)
	ctx.Put(IDFieldFilterKey{}, id)

	return err
}

