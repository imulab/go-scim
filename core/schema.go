package core

import (
	"encoding/json"
	"io/ioutil"
)

type (
	// A SCIM schema or schema extension
	Schema struct {
		Id          string       `json:"id"`
		Name        string       `json:"name"`
		Description string       `json:"description"`
		Attributes  []*Attribute `json:"attributes"`
	}

	// An in memory repository to cache all schemas. It is not thread safe and is
	// intended to used as read-only after initial setup.
	schemaRepository struct {
		mem map[string]*Schema
	}
)

var (
	// Entry point to access schema repository
	Schemas = &schemaRepository{mem: make(map[string]*Schema)}

	// core schema
	ScimCoreSchema = &Schema{
		Attributes: []*Attribute{
			{
				Id:          "schemas",
				Name:        "schemas",
				Type:        TypeReference,
				MultiValued: true,
				Required:    true,
				CaseExact:   true,
				Returned:    ReturnedAlways,
			},
			{
				Id:         "id",
				Name:       "id",
				Type:       TypeString,
				CaseExact:  true,
				Mutability: MutabilityReadOnly,
				Returned:   ReturnedAlways,
				Uniqueness: UniquenessGlobal,
			},
			{
				Id:   "externalId",
				Name: "externalId",
				Type: TypeString,
			},
			{
				Id:         "meta",
				Name:       "meta",
				Type:       TypeComplex,
				Mutability: MutabilityReadOnly,
				SubAttributes: []*Attribute{
					{
						Id:         "meta.resourceType",
						Name:       "resourceType",
						Type:       TypeString,
						CaseExact:  true,
						Mutability: MutabilityReadOnly,
					},
					{
						Id:         "meta.created",
						Name:       "created",
						Type:       TypeDateTime,
						Mutability: MutabilityReadOnly,
					},
					{
						Id:         "meta.lastModified",
						Name:       "lastModified",
						Type:       TypeDateTime,
						Mutability: MutabilityReadOnly,
					},
					{
						Id:         "meta.location",
						Name:       "location",
						Type:       TypeReference,
						CaseExact:  true,
						Mutability: MutabilityReadOnly,
					},
					{
						Id:         "meta.version",
						Name:       "version",
						Type:       TypeString,
						Mutability: MutabilityReadOnly,
					},
				},
			},
		},
	}
)

// Add a schema to repository.
func (r *schemaRepository) Add(schema *Schema) {
	if schema != nil && len(schema.Id) > 0 {
		r.mem[schema.Id] = schema
	}
}

// Get schema from repository by its id, or nil if it does not exist.
func (r *schemaRepository) Get(schemaId string) *Schema {
	return r.mem[schemaId]
}

// Load the schema from a file, or panic
func (r *schemaRepository) MustLoad(filePath string) *Schema {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	schema, err := ParseSchema(raw)
	if err != nil {
		panic(err)
	}

	r.Add(schema)
	return schema
}

// Parse a schema from JSON bytes.
func ParseSchema(raw []byte) (schema *Schema, err error) {
	schema = new(Schema)

	err = json.Unmarshal(raw, &schema)
	if err != nil {
		err = Errors.InvalidSyntax("invalid schema JSON definition")
		return
	}

	return
}
