package spec

import (
	"encoding/json"
	"sync"
)

// Reserved Id for core schema
const CoreSchemaId = "core"

// Schema models a SCIM schema. It is the collection of one or more attributes. Schema structure is read only
// after construction. Schema can be identified by its id, and can be cached in a schema registry.
//
// See also:
//
//	Schemas()
//
// Schema is currently being parsed to and from JSON via special adapters. This design is subject to change when we
// move to treat Schema as just another resource.
// See also:
//
//	issue https://github.com/imulab/go-scim/issues/40
type Schema struct {
	id          string
	name        string
	description string
	attributes  []*Attribute
}

// ID returns the id of the schema.
func (s *Schema) ID() string {
	return s.id
}

// Name returns the name of the schema.
func (s *Schema) Name() string {
	return s.name
}

// Description returns the human-readable text that describes the schema.
func (s *Schema) Description() string {
	return s.description
}

// ForEachAttribute iterate all attributes in this schema and invoke callback function.
func (s *Schema) ForEachAttribute(callback func(attr *Attribute) error) error {
	for _, attr := range s.attributes {
		if err := callback(attr); err != nil {
			return err
		}
	}
	return nil
}

// ResourceTypeName returns the resource type of the Schema resource. This value is formally defined and hence fixed.
func (s *Schema) ResourceTypeName() string {
	return "Schema"
}

// ResourceLocation returns the relative URI at which this Schema resource can be accessed. This value is formally
// defined in the specification and hence fixed.
func (s *Schema) ResourceLocation() string {
	return "/Schemas/" + s.ID()
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(schemaJsonAdapter{
		ID:          s.id,
		Name:        s.name,
		Description: s.description,
		Attributes:  s.attributes,
	})
}

func (s *Schema) UnmarshalJSON(raw []byte) error {
	var adapter schemaJsonAdapter
	if err := json.Unmarshal(raw, &adapter); err != nil {
		return err
	}

	s.id = adapter.ID
	s.name = adapter.Name
	s.description = adapter.Description
	s.attributes = adapter.Attributes
	return nil
}

type schemaJsonAdapter struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Attributes  []*Attribute `json:"attributes"`
}

var (
	schemaReg          *schemaRegistry
	schemaRegistryOnce sync.Once
)

type schemaRegistry struct {
	db map[string]*Schema
}

// Register relates the schema with its id in the registry. This method does not check existence of the id and may
// overwrite existing schemas if abused.
func (r *schemaRegistry) Register(schema *Schema) {
	r.db[schema.id] = schema
}

// Get returns the schema that is related to a schemaId, or nil, along with a boolean indicating if the schema exists.
func (r *schemaRegistry) Get(schemaId string) (schema *Schema, ok bool) {
	schema, ok = r.db[schemaId]
	return
}

// ForEachSchema invokes the callback function on each registered schema.
func (r *schemaRegistry) ForEachSchema(callback func(schema *Schema) error) error {
	for _, schema := range r.db {
		if err := callback(schema); err != nil {
			return err
		}
	}
	return nil
}

func (r *schemaRegistry) mustGet(schemaId string) *Schema {
	schema, ok := r.Get(schemaId)
	if !ok {
		panic("schema " + schemaId + " was not registered")
	}
	return schema
}

// Schemas return the schema registry that holds all registered schemas. Use Get and Register to operate the registry.
func Schemas() *schemaRegistry {
	schemaRegistryOnce.Do(func() {
		schemaReg = &schemaRegistry{db: map[string]*Schema{}}
	})
	return schemaReg
}
