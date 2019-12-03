package filters

import (
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	uuid "github.com/satori/go.uuid"
)

// Create a new resource filter that generates a new uuid on the id field.
func NewIDResourceFilter() protocol.ResourceFilter {
	return &idFilter{}
}

type idFilter struct {}

func (f *idFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource) error {
	idProp, err := resource.NewNavigator().FocusName("id")
	if err != nil {
		return err
	}

	err = idProp.Replace(uuid.NewV4().String())
	if err != nil {
		return err
	}

	return nil
}

func (f *idFilter) FilterRef(ctx *protocol.FilterContext, resource *prop.Resource, ref *prop.Resource) error {
	return nil
}
