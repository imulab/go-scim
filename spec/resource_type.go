package spec

import (
	"encoding/json"
	"github.com/elvsn/scim.go/annotation"
	"github.com/elvsn/scim.go/spec/internal"
)

// Resource type is a collection of one or more schemas to describe a SCIM resource.
type ResourceType struct {
	id          string
	name        string
	description string
	endpoint    string
	schema      *Schema
	extensions  []*Schema
	required    map[string]bool // schema id to boolean to indicate whether schema extension is required
}

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

// ForEachExtension iterates through all schema extensions and invoke the callback.
func (t *ResourceType) ForEachExtension(callback func(extension *Schema, required bool)) {
	for _, ext := range t.extensions {
		callback(ext, t.required[ext.id])
	}
}

func (t *ResourceType) MarshalJSON() ([]byte, error) {
	adapter := internal.ResourceTypeJsonAdapter{}
	t.convertToAdapter(&adapter)
	return json.Marshal(adapter)
}

func (t *ResourceType) convertToAdapter(p *internal.ResourceTypeJsonAdapter) {
	p.ID = t.id
	p.Name = t.name
	p.Description = t.description
	p.Endpoint = t.endpoint
	p.Schema = t.schema.id
	p.Extensions = []*internal.SchemaExtension{}
	for _, ext := range t.extensions {
		p.Extensions = append(p.Extensions, &internal.SchemaExtension{
			Schema:   ext.id,
			Required: t.required[ext.id],
		})
	}
}

func (t *ResourceType) UnmarshalJSON(raw []byte) error {
	var adapter internal.ResourceTypeJsonAdapter
	if err := json.Unmarshal(raw, &adapter); err != nil {
		return err
	}
	t.convertFromAdapter(&adapter)
	return nil
}

func (t *ResourceType) convertFromAdapter(p *internal.ResourceTypeJsonAdapter) {
	t.id = p.ID
	t.name = p.Name
	t.description = p.Description
	t.endpoint = p.Endpoint
	t.schema = Schemas().mustGet(p.Schema)
	t.extensions = []*Schema{}
	t.required = map[string]bool{}
	for _, ext := range p.Extensions {
		t.extensions = append(t.extensions, Schemas().mustGet(ext.Schema))
		t.required[ext.Schema] = ext.Required
	}
}

// SuperAttribute return a virtual complex attribute that contains all schema attributes as its sub attributes.
func (t *ResourceType) SuperAttribute(includeCore bool) *Attribute {
	super := Attribute{
		id:            t.schema.id,
		typ:           TypeComplex,
		subAttributes: []*Attribute{},
		mutability:    MutabilityReadWrite,
		returned:      ReturnedDefault,
		uniqueness:    UniquenessNone,
		annotations: map[string]map[string]interface{}{
			annotation.Root: {},
		},
	}

	if includeCore {
		super.subAttributes = append(super.subAttributes, Schemas().mustGet(internal.CoreSchemaId).attributes...)
		super.annotations[annotation.SyncSchema] = map[string]interface{}{}
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
			annotations: map[string]map[string]interface{}{
				annotation.StateSummary:        {},
				annotation.SchemaExtensionRoot: {},
			},
		})
		i++
	})

	return &super
}
