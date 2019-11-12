package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// A SCIM schema or schema extension
type Schema struct {
	Id          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Attributes  []*Attribute     `json:"attributes"`
	Companion   *SchemaCompanion `json:"-"`
}

// Parse a schema from JSON bytes.
func ParseSchema(raw []byte) (schema *Schema, err error) {
	schema = new(Schema)

	err = json.Unmarshal(raw, &schema)
	if err != nil {
		err = Errors.InvalidSyntax("invalid schema JSON definition")
		return
	}

	for _, attr := range schema.Attributes {
		attr.setDefaults()
	}
	return
}

// Load the metadata of the schema companion onto the attribute. Panics if operation failed.
func (sch *Schema) mustLoadCompanion(sc *SchemaCompanion) {
	sch.Companion = sc

	dfs := make([]*Attribute, 0)
	prefix := make([]string, 0)

	// note: SCIM only allow two levels, hence we are only checking two levels.
	for _, attr := range sch.Attributes {
		dfs = append(dfs, attr)
		prefix = append(prefix, "")

		for _, subAttr := range attr.SubAttributes {
			dfs = append(dfs, subAttr)
			prefix = append(prefix, attr.Name)
		}
	}

	if len(dfs) != len(sc.Metadata) {
		panic(Errors.Internal(fmt.Sprintf(
			"failed to load schema companion for '%s': metadata size different with attributes count", sch.Id,
		)))
	}

	for i, attr := range dfs {
		path := attr.Name
		if len(prefix[i]) > 0 {
			path = prefix[i] + "." + path
		}

		metadata := sc.Metadata[i]
		if strings.ToLower(path) != strings.ToLower(metadata.Path) {
			panic(Errors.Internal(fmt.Sprintf(
				"failed to load schema companion for '%s': metadata path mismatch with attribute path", sch.Id,
			)))
		}

		attr.Metadata = metadata
	}
}

var (
	// Entry point to access schema repository
	Schemas = &schemaRepository{mem: make(map[string]*Schema)}
)

// An in memory repository to cache all schemas. It is not thread safe and is
// intended to used as read-only after initial setup.
type schemaRepository struct {
	mem map[string]*Schema
}

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

var (
	// core schema
	ScimCoreSchema = &Schema{
		Attributes: []*Attribute{
			{
				Name:        "schemas",
				Type:        TypeReference,
				MultiValued: true,
				Required:    true,
				CaseExact:   true,
				Mutability:  MutabilityReadWrite,
				Returned:    ReturnedAlways,
				Uniqueness:  UniquenessNone,
				Metadata: &Metadata{
					Path:        "schemas",
					Annotations: []string{AnnotationSchemas},
				},
			},
			{
				Name:       "id",
				Type:       TypeString,
				CaseExact:  true,
				Mutability: MutabilityReadOnly,
				Returned:   ReturnedAlways,
				Uniqueness: UniquenessGlobal,
				Metadata: &Metadata{
					Path:        "id",
					Annotations: []string{AnnotationId},
				},
			},
			{
				Name:       "externalId",
				Type:       TypeString,
				Mutability: MutabilityReadWrite,
				Returned:   ReturnedDefault,
				Uniqueness: UniquenessNone,
				Metadata: &Metadata{
					Path: "externalId",
				},
			},
			{
				Name:       "meta",
				Type:       TypeComplex,
				Mutability: MutabilityReadOnly,
				Returned:   ReturnedDefault,
				Uniqueness: UniquenessNone,
				Metadata: &Metadata{
					Path:        "meta",
					Annotations: []string{AnnotationMeta},
				},
				SubAttributes: []*Attribute{
					{
						Name:       "resourceType",
						Type:       TypeString,
						CaseExact:  true,
						Mutability: MutabilityReadOnly,
						Returned:   ReturnedDefault,
						Uniqueness: UniquenessNone,
						Metadata: &Metadata{
							Path: "meta.resourceType",
						},
					},
					{
						Name:       "created",
						Type:       TypeDateTime,
						Mutability: MutabilityReadOnly,
						Returned:   ReturnedDefault,
						Uniqueness: UniquenessNone,
						Metadata: &Metadata{
							Path: "meta.created",
						},
					},
					{
						Name:       "lastModified",
						Type:       TypeDateTime,
						Mutability: MutabilityReadOnly,
						Returned:   ReturnedDefault,
						Uniqueness: UniquenessNone,
						Metadata: &Metadata{
							Path: "meta.lastModified",
						},
					},
					{
						Name:       "location",
						Type:       TypeReference,
						CaseExact:  true,
						Mutability: MutabilityReadOnly,
						Returned:   ReturnedDefault,
						Uniqueness: UniquenessNone,
						Metadata: &Metadata{
							Path: "meta.location",
						},
					},
					{
						Name:       "version",
						Type:       TypeString,
						Mutability: MutabilityReadOnly,
						Returned:   ReturnedDefault,
						Uniqueness: UniquenessNone,
						Metadata: &Metadata{
							Path: "meta.version",
						},
					},
				},
			},
		},
	}
)
