package spec

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/spec/internal"
)

// Resource type models the SCIM resource type. It is a collection of one main schema and zero or more schema extensions
// to describe a single type of SCIM resource.
//
// To access the main schema of the resource type, call Schema method; to access the schema extensions, use ForEachExtension
// method.
//
// A resource type can be used to generate a super attribute. A super attribute is a single valued complex typed attribute
// that encapsulates all attributes from its main schema and schema extensions as sub attributes. Among these, the top-level
// attributes from the main schema will be added to the super attribute as top level sub attributes. The top-level attributes
// from each schema extension will be first added to a container complex attribute as top level sub attributes, and then that
// single container complex attribute is added to the super attribute as a top level sub attribute.
//
// For example, suppose we have a main schema containing attribute A1 and A2, a schema extension B containing attribute B1,
// and another schema extension C containing attribute C1 and C2. The super attribute will be in the structure of:
//	{
//		A1,
//		A2,
//		B {
//			B1
//		},
//		C {
//			C1,
//			C2
//		}
//	}
//
// ResourceType is currently being parsed to and from JSON using special adapters. This design is subject to change
// when we move to treat ResourceType as just another resource.
// See also:
//	issue https://github.com/imulab/go-scim/issues/40
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
func (t *ResourceType) ForEachExtension(callback func(extension *Schema, required bool) error) error {
	for _, ext := range t.extensions {
		if err := callback(ext, t.required[ext.id]); err != nil {
			return err
		}
	}
	return nil
}

// CountExtensions returns the total number of extensions
func (t *ResourceType) CountExtensions() int {
	return len(t.extensions)
}

// ResourceTypeName returns the resource type of the ResourceType resource. This value is formally defined and hence fixed.
func (t *ResourceType) ResourceTypeName() string {
	return "ResourceType"
}

// ResourceLocation returns the relative URI at which this ResourceType resource can be accessed. This value is formally
// defined in the specification and hence fixed.
func (t *ResourceType) ResourceLocation() string {
	return "/ResourceTypes/" + t.ID()
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
		super.subAttributes = append(super.subAttributes, Schemas().mustGet(CoreSchemaId).attributes...)
		super.annotations[annotation.SyncSchema] = map[string]interface{}{}
	}

	super.subAttributes = append(super.subAttributes, t.schema.attributes...)

	var i = len(super.subAttributes)
	_ = t.ForEachExtension(func(extension *Schema, required bool) error {
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
		return nil
	})

	return &super
}
