package filters

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	uuid "github.com/satori/go.uuid"
)

// Create a new field filter that generates a new uuid on the id field.
func NewIDFieldFilter(order int) protocol.FieldFilter {
	return &idFilter{order: order}
}

type idFilter struct {
	order int
}

func (u *idFilter) Supports(attribute *core.Attribute) bool {
	return attribute.ID() == "id"
}

func (u *idFilter) Order() int {
	return u.order
}

func (u *idFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property) error {
	_, err := prop.Internal(property).Replace(uuid.NewV4().String())
	return err
}

func (u *idFilter) FieldRef(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	return nil
}
