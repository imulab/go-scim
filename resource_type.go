package scim

import "encoding/json"

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

func (r *ResourceType[T]) MarshalJSON() ([]byte, error) {
	j := resourceTypeJSON{
		Id:         r.id,
		Name:       r.name,
		Endpoint:   r.endpoint,
		Schema:     r.schema.id,
		Extensions: []schemaExtensionJSON{},
	}

	for _, it := range r.extensions {
		j.Extensions = append(j.Extensions, schemaExtensionJSON{
			Schema:   it.id,
			Required: it.required,
		})
	}

	return json.Marshal(j)
}

type resourceTypeJSON struct {
	Id         string                `json:"id"`
	Name       string                `json:"name"`
	Endpoint   string                `json:"endpoint,omitempty"`
	Schema     string                `json:"schema"`
	Extensions []schemaExtensionJSON `json:"schemaExtensions,omitempty"`
}

type schemaExtensionJSON struct {
	Schema   string `json:"schema"`
	Required bool   `json:"required"`
}

type resourceTypeDsl[T any] ResourceType[T]

func NewResourceType[T any](id string) *resourceTypeDsl[T] {
	d := new(resourceTypeDsl[T])
	d.id = id
	return d
}

func (d *resourceTypeDsl[T]) Name(name string) *resourceTypeDsl[T] {
	d.name = name
	return d
}

func (d *resourceTypeDsl[T]) Describe(text string) *resourceTypeDsl[T] {
	d.description = text
	return d
}

func (d *resourceTypeDsl[T]) Location(baseUrl, endpoint string) *resourceTypeDsl[T] {
	d.endpoint = endpoint
	d.location = baseUrl + endpoint
	return d
}

func (d *resourceTypeDsl[T]) MainSchema(fn func(sd *schemaDsl)) *resourceTypeDsl[T] {
	d.schema = BuildSchema(fn)
	return d
}

func (d *resourceTypeDsl[T]) ExtendSchema(fn func(sd *schemaDsl), required bool) *resourceTypeDsl[T] {
	ext := BuildSchema(fn)
	ext.required = required
	d.extensions = append(d.extensions, ext)
	return d
}

func (d *resourceTypeDsl[T]) NewFunc(fn func() *T) *resourceTypeDsl[T] {
	d.newModel = fn
	return d
}

func (d *resourceTypeDsl[T]) AddMapping(fn func(md *mappingDsl[T])) *resourceTypeDsl[T] {
	md := new(mappingDsl[T])
	fn(md)

	d.mappings = append(d.mappings, md.build())

	return d
}

func (d *resourceTypeDsl[T]) Build() *ResourceType[T] {
	registerURN(d.schema.id)
	for _, it := range d.extensions {
		registerURN(it.id)
	}

	return (*ResourceType[T])(d)
}
