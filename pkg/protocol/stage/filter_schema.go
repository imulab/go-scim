package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
)

// Return a new schema filter. The filter is responsible of handling the schema attribute. The ids of the main schema
// and the required schema extension must exist in the schema property.
func NewSchemaFilter(order int) PropertyFilter {
	return &schemaFilter{order: order}
}

var _ PropertyFilter = (*schemaFilter)(nil)

type schemaFilter struct {order int}

func (f *schemaFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "schemas" && attribute.MultiValued
}

func (f *schemaFilter) Order() int {
	return f.order
}

func (f *schemaFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.checkSchema(resource, property)
}

func (f *schemaFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return f.checkSchema(resource, property)
}

func (f *schemaFilter) checkSchema(resource *core.Resource, property core.Property) error {
	resourceType := resource.GetResourceType()

	var requiredSchemas []string
	{
		requiredSchemas = []string{resourceType.Schema}
		for _, extension := range resourceType.SchemaExtensions {
			if extension.Required {
				requiredSchemas = append(requiredSchemas, extension.Schema)
			}
		}
	}

	var schemas []string
	{
		schemas = make([]string, 0)
		for _, elem := range property.Children() {
			if !elem.IsUnassigned() {
				schemas = append(schemas, elem.Raw().(string))
			}
		}
	}

	for _, required := range requiredSchemas {
		var found = false
		for _, schema := range schemas {
			if required == schema {
				found = true
				break
			}
		}
		if !found {
			return core.Errors.InvalidValue("schema '%s' is required, but is not found", required)
		}
	}

	return nil
}