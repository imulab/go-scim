package spec

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/annotations"
)

// Central repository for schemas
var SchemaHub = &schemaHub{schemaById: make(map[string]*Schema)}

// A SCIM schema is a collection of attributes, used to describe
// a whole or a part of a resource.
type Schema struct {
	id          string
	name        string
	description string
	attributes  []*Attribute
}

// Return the ID of the schema.
func (s *Schema) ID() string {
	return s.id
}

// Return the name of the schema.
func (s *Schema) Name() string {
	return s.name
}

// Return the description of the schema
func (s *Schema) Description() string {
	return s.description
}

// Return the number of top level attributes in this schema.
func (s *Schema) CountAttributes() int {
	return len(s.attributes)
}

// Iterate the top level attributes in this schema and invoke callback function.
// This method maintains SOLID principal. Callback function SHALL NOT block.
func (s *Schema) ForEachAttribute(callback func(attr *Attribute)) {
	for _, attr := range s.attributes {
		callback(attr)
	}
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	tmp := new(schemaJSONAdapter)
	tmp.extract(s)
	return json.Marshal(tmp)
}

func (s *Schema) UnmarshalJSON(raw []byte) error {
	var tmp *schemaJSONAdapter
	{
		tmp = new(schemaJSONAdapter)
		err := json.Unmarshal(raw, tmp)
		if err != nil {
			return nil
		}
	}
	tmp.fill(s)
	return nil
}

type (
	// Adapter for schema to serialize to and deserialize from JSON format, so that schema does not have
	// to expose internal pointers.
	schemaJSONAdapter struct {
		ID          string       `json:"id"`
		Name        string       `json:"name"`
		Description string       `json:"description"`
		Attributes  []*Attribute `json:"attributes"`
	}
	// A collection of schemas by their ID.
	schemaHub struct {
		schemaById map[string]*Schema
	}
)

// Extract values from schema into this adapter
func (d *schemaJSONAdapter) extract(s *Schema) {
	d.ID = s.id
	d.Name = s.name
	d.Description = s.description
	d.Attributes = s.attributes
}

// Fill values of this adapter in the schema
func (d *schemaJSONAdapter) fill(s *Schema) {
	s.id = d.ID
	s.name = d.Name
	s.description = d.Description
	s.attributes = d.Attributes
}

// Put schema into the hub.
func (h *schemaHub) Put(s *Schema) {
	h.schemaById[s.ID()] = s
}

// Get schema by id. This method panics if schema is not found by the id.
func (h *schemaHub) MustGet(id string) *Schema {
	s, ok := h.schemaById[id]
	if !ok || s == nil {
		panic("schema not found by id '" + id + "'")
	}
	return s
}

// Return the core schema, which includes the common attributes of schemas, id, externalId, meta.
func (h *schemaHub) CoreSchema() *Schema {
	return &Schema{
		attributes: []*Attribute{
			{
				id:          "schemas",
				name:        "schemas",
				path:        "schemas",
				typ:         TypeReference,
				multiValued: true,
				required:    true,
				caseExact:   true,
				returned:    ReturnedAlways,
				index:       0,
				annotationIndex: map[string]struct{}{
					annotations.AutoCompact: {},
				},
			},
			{
				id:         "id",
				name:       "id",
				path:       "id",
				typ:        TypeString,
				caseExact:  true,
				mutability: MutabilityReadOnly,
				returned:   ReturnedAlways,
				uniqueness: UniquenessGlobal,
				index:      1,
				annotationIndex: map[string]struct{}{
					annotations.CopyReadOnly: {},
				},
			},
			{
				id:    "externalId",
				name:  "externalId",
				path:  "externalId",
				typ:   TypeString,
				index: 2,
			},
			{
				id:         "meta",
				name:       "meta",
				path:       "meta",
				typ:        TypeComplex,
				mutability: MutabilityReadOnly,
				index:      3,
				subAttributes: []*Attribute{
					{
						id:         "meta.resourceType",
						name:       "resourceType",
						path:       "meta.resourceType",
						typ:        TypeString,
						caseExact:  true,
						mutability: MutabilityReadOnly,
						index:      0,
						annotationIndex: map[string]struct{}{
							annotations.CopyReadOnly: {},
						},
					},
					{
						id:         "meta.created",
						name:       "created",
						path:       "meta.created",
						typ:        TypeDateTime,
						mutability: MutabilityReadOnly,
						index:      1,
						annotationIndex: map[string]struct{}{
							annotations.CopyReadOnly: {},
						},
					},
					{
						id:         "meta.lastModified",
						name:       "lastModified",
						path:       "meta.lastModified",
						typ:        TypeDateTime,
						mutability: MutabilityReadOnly,
						index:      2,
						annotationIndex: map[string]struct{}{
							annotations.CopyReadOnly: {},
						},
					},
					{
						id:         "meta.location",
						name:       "location",
						path:       "meta.location",
						typ:        TypeReference,
						caseExact:  true,
						mutability: MutabilityReadOnly,
						index:      3,
						annotationIndex: map[string]struct{}{
							annotations.CopyReadOnly: {},
						},
					},
					{
						id:         "meta.version",
						name:       "version",
						path:       "meta.version",
						typ:        TypeString,
						mutability: MutabilityReadOnly,
						index:      4,
						annotationIndex: map[string]struct{}{
							annotations.CopyReadOnly: {},
						},
					},
				},
			},
		},
	}
}

var (
	_ json.Marshaler   = (*Schema)(nil)
	_ json.Unmarshaler = (*Schema)(nil)
)
