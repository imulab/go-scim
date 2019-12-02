package core

import "encoding/json"

var (
	_ json.Marshaler   = (*Schema)(nil)
	_ json.Unmarshaler = (*Schema)(nil)

	// Central repository for schemas
	SchemaHub = &schemaHub{schemaById: make(map[string]*Schema)}
)

type (
	// A SCIM schema is a collection of attributes, used to describe a whole or a part of a resource.
	Schema struct {
		id          string
		name        string
		description string
		attributes  []*Attribute
	}
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
				typ:         TypeReference,
				multiValued: true,
				required:    true,
				caseExact:   true,
				returned:    ReturnedAlways,
				index:       0,
				path:        "schemas",
			},
			{
				id:         "id",
				name:       "id",
				typ:        TypeString,
				caseExact:  true,
				mutability: MutabilityReadOnly,
				returned:   ReturnedAlways,
				uniqueness: UniquenessGlobal,
				index:      1,
				path:       "id",
				annotations: []string{
					"@CopyReadOnly",
				},
			},
			{
				id:    "externalId",
				name:  "externalId",
				typ:   TypeString,
				index: 2,
				path:  "externalId",
			},
			{
				id:         "meta",
				name:       "meta",
				typ:        TypeComplex,
				mutability: MutabilityReadOnly,
				index:      3,
				path:       "meta",
				subAttributes: []*Attribute{
					{
						id:         "meta.resourceType",
						name:       "resourceType",
						typ:        TypeString,
						caseExact:  true,
						mutability: MutabilityReadOnly,
						index:      0,
						path:       "meta.resourceType",
						annotations: []string{
							"@CopyReadOnly",
						},
					},
					{
						id:         "meta.created",
						name:       "created",
						typ:        TypeDateTime,
						mutability: MutabilityReadOnly,
						index:      1,
						path:       "meta.created",
						annotations: []string{
							"@CopyReadOnly",
						},
					},
					{
						id:         "meta.lastModified",
						name:       "lastModified",
						typ:        TypeDateTime,
						mutability: MutabilityReadOnly,
						index:      2,
						path:       "meta.lastModified",
						annotations: []string{
							"@CopyReadOnly",
						},
					},
					{
						id:         "meta.location",
						name:       "location",
						typ:        TypeReference,
						caseExact:  true,
						mutability: MutabilityReadOnly,
						index:      3,
						path:       "meta.location",
						annotations: []string{
							"@CopyReadOnly",
						},
					},
					{
						id:         "meta.version",
						name:       "version",
						typ:        TypeString,
						mutability: MutabilityReadOnly,
						index:      4,
						path:       "meta.version",
						annotations: []string{
							"@CopyReadOnly",
						},
					},
				},
			},
		},
	}
}
