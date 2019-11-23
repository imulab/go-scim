package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/persistence"
)

type referenceTypeFilter struct {
	// A map from resource type's name to the persistence provider capable
	// of handling of the resource type
	providers map[string]persistence.Provider
}

func (f *referenceTypeFilter) Supports(attribute *core.Attribute) bool {
	panic("implement me")
}

func (f *referenceTypeFilter) Order(attribute *core.Attribute) int {
	panic("implement me")
}

func (f *referenceTypeFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	panic("implement me")
}

func (f *referenceTypeFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	panic("implement me")
}

