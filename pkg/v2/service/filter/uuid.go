package filter

import (
	"context"
	"github.com/google/uuid"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// UUIDFilter returns a ByProperty filter that generates a UUID for string property that is annotated with @UUID. The
// generation only happens when the target property is currently unassigned. The generated value will trigger event
// propagation.
func UUIDFilter() ByProperty {
	return uuidPropertyFilter{}
}

type uuidPropertyFilter struct{}

func (f uuidPropertyFilter) Supports(attribute *spec.Attribute) bool {
	_, ok := attribute.Annotation(annotation.UUID)
	if !ok {
		return false
	}
	return !attribute.MultiValued() && attribute.Type() == spec.TypeString
}

func (f uuidPropertyFilter) Filter(_ context.Context, _ *spec.ResourceType, nav prop.Navigator) error {
	if nav.HasError() {
		return nav.Error()
	}

	if !nav.Current().IsUnassigned() {
		return nil
	}

	return nav.Replace(uuid.New().String()).Error()
}

func (f uuidPropertyFilter) FilterRef(_ context.Context, _ *spec.ResourceType, _ prop.Navigator, _ prop.Navigator) error {
	return nil
}
