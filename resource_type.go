package scim

import (
	"encoding/json"
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

// New creates a new Resource with this ResourceType.
func (r *ResourceType[T]) New() *Resource[T] {
	return &Resource[T]{
		resourceType: r,
		root:         r.archAttribute().createProperty().(*complexProperty),
	}
}

// archAttribute returns an ad-hoc attribute to model the entire resource structure. The returned attribute will
// include all coreAttributes, all attributes from the main schema, and optionally all attributes from any extensions
// under a complex attribute named after the extension's id.
func (r *ResourceType[T]) archAttribute() *Attribute {
	arch := &Attribute{
		name:     r.name,
		typ:      TypeComplex,
		sub:      []*Attribute{},
		required: true,
	}

	arch.sub = append(arch.sub, coreAttributes...)
	arch.sub = append(arch.sub, r.schema.attrs...)

	for _, it := range r.extensions {
		arch.sub = append(arch.sub, &Attribute{
			name:     it.id,
			typ:      TypeComplex,
			sub:      append([]*Attribute{}, it.attrs...),
			required: it.required,
			id:       it.id,
			path:     it.id,
		})
	}

	return arch
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

	for _, it := range d.mappings {
		head, err := compilePath(it.path)
		if err != nil {
			panic(err)
		}
		it.compiledPath = head
	}

	return (*ResourceType[T])(d)
}
