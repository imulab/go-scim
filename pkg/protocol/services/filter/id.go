package filter

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/prop"
	uuid "github.com/satori/go.uuid"
)

// Create a new resource filter that generates a new uuid on the id field.
func ID() ForResource {
	return &idFilter{}
}

type idFilter struct{}

func (f *idFilter) Filter(ctx context.Context, resource *prop.Resource) error {
	if idProp, err := resource.NewNavigator().FocusName("id"); err != nil {
		return err
	} else if err := idProp.Replace(uuid.NewV4().String()); err != nil {
		return err
	}
	return nil
}

func (f *idFilter) FilterRef(ctx context.Context, resource *prop.Resource, ref *prop.Resource) error {
	return nil
}
