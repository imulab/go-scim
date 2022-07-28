package scim

import (
	"encoding/json"
	"github.com/imulab/go-scim/internal/expr"
)

const (
	resourceTypeURN = "urn:ietf:params:scim:schemas:core:2.0:ResourceType"
)

type ResourceType[T any] struct {
	id          string
	name        string
	description string
	endpoint    string
	location    string
	schema      *Schema
	extensions  []*Schema
	newModel    func() *T
	mappings    []*Mapping[T]
}

// archetypeAttribute returns an ad-hoc Attribute which contains all main Schema attribute on the top level, and all
// extension Schema attributes as sub attributes of a complex attribute named the id of the extension, placed also on
// the top level. This method is useful to create the root complex property of a Resource.
func (t *ResourceType[T]) archetypeAttribute() *Attribute {
	a := &Attribute{
		typ:      TypeComplex,
		subAttrs: []*Attribute{},
		required: true,
	}

	a.subAttrs = append(a.subAttrs, coreSchema.attrs...)
	a.subAttrs = append(a.subAttrs, t.schema.attrs...)

	for _, ext := range t.extensions {
		a.subAttrs = append(a.subAttrs, &Attribute{
			name:     ext.id,
			typ:      TypeComplex,
			subAttrs: ext.attrs,
			required: ext.required,
		})
	}

	return a
}

func (t *ResourceType[T]) registerSchemaURNs() {
	if t.schema != nil {
		expr.RegisterURN(t.schema.id)
	}

	for _, ext := range t.extensions {
		expr.RegisterURN(ext.id)
	}
}

func (t *ResourceType[T]) MarshalJSON() ([]byte, error) {
	j := resourceTypeJSON{
		Schemas: []string{resourceTypeURN},
		Meta: &metaJSON{
			Location:     t.location,
			ResourceType: "ResourceType",
		},
		Id:          t.id,
		Name:        t.name,
		Endpoint:    t.endpoint,
		Description: t.description,
		Schema:      t.schema.id,
	}

	for _, each := range t.extensions {
		j.Extensions = append(j.Extensions, &schemaExtensionJSON{
			Schema:   each.id,
			Required: each.required,
		})
	}

	return json.Marshal(j)
}

// moved outside because type declarations inside generic functions are not currently supported in Go1.18
type (
	metaJSON struct {
		Location     string `json:"location,omitempty"`
		ResourceType string `json:"resourceType"`
	}
	schemaExtensionJSON struct {
		Schema   string `json:"schema"`
		Required bool   `json:"required"`
	}
	resourceTypeJSON struct {
		Schemas     []string               `json:"schemas"`
		Meta        *metaJSON              `json:"meta"`
		Id          string                 `json:"id"`
		Name        string                 `json:"name"`
		Endpoint    string                 `json:"endpoint,omitempty"`
		Description string                 `json:"description,omitempty"`
		Schema      string                 `json:"schema"`
		Extensions  []*schemaExtensionJSON `json:"extensions,omitempty"`
	}
)

func BuildResourceType[T any](id string) *resourceTypeDsl[T] {
	return &resourceTypeDsl[T]{
		ResourceType: &ResourceType[T]{
			id:         id,
			extensions: []*Schema{},
		},
	}
}

type resourceTypeDsl[T any] struct {
	*ResourceType[T]
}

func (d *resourceTypeDsl[T]) Name(name string) *resourceTypeDsl[T] {
	d.ResourceType.name = name
	return d
}

func (d *resourceTypeDsl[T]) Endpoint(endpoint string) *resourceTypeDsl[T] {
	d.ResourceType.endpoint = endpoint
	return d
}

func (d *resourceTypeDsl[T]) SelfLocation(location string) *resourceTypeDsl[T] {
	d.ResourceType.location = location
	return d
}

func (d *resourceTypeDsl[T]) Description(description string) *resourceTypeDsl[T] {
	d.ResourceType.description = description
	return d
}

func (d *resourceTypeDsl[T]) MainSchema(schema *Schema) *resourceTypeDsl[T] {
	d.ResourceType.schema = schema
	return d
}

func (d *resourceTypeDsl[T]) AddExtensionSchema(schemas ...*Schema) *resourceTypeDsl[T] {
	d.ResourceType.extensions = append(d.ResourceType.extensions, schemas...)
	return d
}

func (d *resourceTypeDsl[T]) NewFunc(fn func() *T) *resourceTypeDsl[T] {
	d.ResourceType.newModel = fn
	return d
}

func (d *resourceTypeDsl[T]) AddMapping(mappings ...*Mapping[T]) *resourceTypeDsl[T] {
	d.ResourceType.mappings = append(d.ResourceType.mappings, mappings...)
	return d
}

func (d *resourceTypeDsl[T]) Build() *ResourceType[T] {
	switch {
	case len(d.ResourceType.id) == 0:
		panic("id is required")
	case len(d.ResourceType.name) == 0:
		panic("name is required")
	case d.ResourceType.schema == nil:
		panic("main schema is required")
	default:
		d.ResourceType.registerSchemaURNs()
		return d.ResourceType
	}
}
