package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
)

// Create an required filter. The filter is responsible for checking any attribute whose required is set true that they
// are not unassigned.
func NewRequiredFilter() PropertyFilter {
	return &requiredFilter{}
}

var (
	_ PropertyFilter = (*requiredFilter)(nil)
)

type requiredFilter struct {}

func (f *requiredFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Required
}

func (f *requiredFilter) Order(attribute *core.Attribute) int {
	return 500
}

func (f *requiredFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.required(property)
}

func (f *requiredFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return f.required(property)
}

func (f *requiredFilter) required(property core.Property) error {
	if !property.IsUnassigned() {
		return nil
	}
	return core.Errors.InvalidValue("'%s' is required, but is unassigned", property.Attribute().DisplayName())
}