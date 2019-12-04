package core

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core/annotations"
)

var (
	_ json.Marshaler   = (*ResourceType)(nil)
	_ json.Unmarshaler = (*ResourceType)(nil)
)

type (
	// A SCIM resource type is a collection of schemas to describe a resource.
	ResourceType struct {
		id          string
		name        string
		description string
		endpoint    string
		schema      *Schema
		extensions  []*Schema
		// A map of schema id to a bool indicating
		// whether the schema by this id is required
		required map[string]bool
	}
	// JSON adapter to resource type to serialize to and deserialize from JSON format.
	resourceTypeJSONAdapter struct {
		ID          string             `json:"id"`
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Endpoint    string             `json:"endpoint"`
		Schema      string             `json:"schema"`
		Extensions  []*schemaExtension `json:"schemaExtensions"`
	}
	schemaExtension struct {
		Schema   string `json:"schema"`
		Required bool   `json:"required"`
	}
)

// Return the id of the resource type
func (t *ResourceType) ID() string {
	return t.id
}

// Return the name of the resource type
func (t *ResourceType) Name() string {
	return t.name
}

// Return the description of the resource type
func (t *ResourceType) Description() string {
	return t.description
}

func (t *ResourceType) Endpoint() string {
	return t.endpoint
}

// Return the main schema of the resource type
func (t *ResourceType) Schema() *Schema {
	return t.schema
}

// Return the number of schema extensions in this resource type
func (t *ResourceType) CountExtensions() int {
	return len(t.extensions)
}

// Iterate through all schema extensions and invoke the callback. The required parameter is the callback
// function indicates whether this schema extension is required.
// This method maintains SOLID principal. The callback SHALL NOT block.
func (t *ResourceType) ForEachExtension(callback func(extension *Schema, required bool)) {
	for _, ext := range t.extensions {
		callback(ext, t.required[ext.id])
	}
}

// Return a complex attribute that contains all schema attributes as its sub attributes.
func (t *ResourceType) SuperAttribute(includeCore bool) *Attribute {
	super := &Attribute{
		id:            t.schema.id,
		typ:           TypeComplex,
		subAttributes: []*Attribute{},
		mutability:    MutabilityReadWrite,
		returned:      ReturnedDefault,
		uniqueness:    UniquenessNone,
	}

	if includeCore {
		super.subAttributes = append(super.subAttributes, SchemaHub.CoreSchema().attributes...)
		super.annotations = append(super.annotations, annotations.SyncSchema)
	}
	super.subAttributes = append(super.subAttributes, t.schema.attributes...)

	var i = len(super.subAttributes)
	t.ForEachExtension(func(extension *Schema, required bool) {
		super.subAttributes = append(super.subAttributes, &Attribute{
			id:            extension.id,
			name:          extension.id,
			description:   extension.description,
			typ:           TypeComplex,
			subAttributes: extension.attributes,
			required:      required,
			mutability:    MutabilityReadWrite,
			returned:      ReturnedDefault,
			uniqueness:    UniquenessNone,
			index:         i,
			path:          extension.id,
			annotations:   []string{annotations.StateSummary, annotations.SchemaExtensionRoot},
		})
		i++
	})

	return super
}

func (t *ResourceType) MarshalJSON() ([]byte, error) {
	tmp := new(resourceTypeJSONAdapter)
	tmp.extract(t)
	return json.Marshal(tmp)
}

func (t *ResourceType) UnmarshalJSON(raw []byte) error {
	var tmp *resourceTypeJSONAdapter
	{
		tmp = new(resourceTypeJSONAdapter)
		err := json.Unmarshal(raw, tmp)
		if err != nil {
			return err
		}
	}
	tmp.fill(t)
	return nil
}

// Extract values from resource type to this adapter.
func (d *resourceTypeJSONAdapter) extract(t *ResourceType) {
	d.ID = t.id
	d.Name = t.name
	d.Description = t.description
	d.Endpoint = t.endpoint
	d.Schema = t.schema.id
	d.Extensions = make([]*schemaExtension, len(t.extensions), len(t.extensions))
	for i, ext := range t.extensions {
		d.Extensions[i] = &schemaExtension{
			Schema:   ext.id,
			Required: t.required[ext.id],
		}
	}
}

// Fill values from this adapter into the resource type
func (d *resourceTypeJSONAdapter) fill(t *ResourceType) {
	t.id = d.ID
	t.name = d.Name
	t.description = d.Description
	t.endpoint = d.Endpoint
	t.schema = SchemaHub.MustGet(d.Schema)
	if len(d.Extensions) > 0 {
		t.extensions = make([]*Schema, len(d.Extensions), len(d.Extensions))
		t.required = make(map[string]bool)
		for i, ext := range d.Extensions {
			t.extensions[i] = SchemaHub.MustGet(ext.Schema)
			t.required[ext.Schema] = ext.Required
		}
	}
}
