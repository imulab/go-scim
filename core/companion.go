package core

import (
	"encoding/json"
	"fmt"
)

// Schema companion is the container for all metadata for one schema. It maintains a 1-to-1
// relation with a schema.
type SchemaCompanion struct {
	// Id of the schema that this companion is for.
	Schema string `json:"schema"`

	// Alias for the schema id in the database in case it is used as a container
	// for its attributes. For instance, schema extensions will use the extension
	// id as a complex property container for all its defined properties. Some database
	// does not play well with the '.' (dot) in the URN namespace. If not specified, defaults
	// to schema id.
	DbAlias string `json:"dbAlias"`

	// All metadata for this schema companion. The ordering of the metadata definition matters!
	// It must follow a depth-first-search order of the attributes and sub-attributes defined
	// in the schema. This is crucial for the metadata to be successfully loaded onto the attributes
	// without resorting to hash maps indexed with full path of attributes.
	Metadata []*Metadata `json:"metadata"`
}

// Parse the schema companion from JSON bytes and validate it.
func ParseSchemaCompanion(raw []byte) (sc *SchemaCompanion, err error) {
	sc = new(SchemaCompanion)

	err = json.Unmarshal(raw, &sc)
	if err != nil {
		err = Errors.InvalidSyntax("invalid schema companion JSON definition")
		return
	}

	err = sc.validate()
	return
}

// validate the schema companion
func (sc *SchemaCompanion) validate() error {
	if len(sc.Schema) == 0 {
		return Errors.internal("schema field is required for schema companion")
	}

	if len(sc.Metadata) == 0 {
		return Errors.internal("no metadata defined for schema companion")
	}

	for _, md := range sc.Metadata {
		if err := md.validate(); err != nil {
			return err
		}
	}

	return nil
}

// Entry point load schema companion onto schemas. Panics if operation fails.
func (sc *SchemaCompanion) MustLoadOntoSchema() {
	schema := Schemas.Get(sc.Schema)
	if schema == nil {
		panic(Errors.internal(fmt.Sprintf("no such schema by id '%s'", sc.Schema)))
	}
	schema.mustLoadCompanion(sc)
}

// Metadata is a remodelling of the former assist fields in v1. It is now an independent
// entity will can be loaded onto attributes during startup. This enables cleaner definition
// of schemas and no longer needs to distinguish between internal and external schema.
type Metadata struct {
	// path is the name that can be used to address
	// the property from the root of the resource. (i.e. userName, emails.value)
	Path string `json:"path"`

	// DB alias field is an alternative name to the attribute's name
	// when persisting to database. Some database does not welcome
	// certain SCIM attribute name rules, hence the need for an
	// alternative name. If not specified, defaults to attribute's name
	// https://github.com/imulab/go-scim/issues/7
	DbAlias string `json:"dbAlias"`

	// True if the property shall participate in identity comparison
	// Two complex properties shall be regarded as equal when all their
	// respective identity sub properties equal. For instance, when
	// comparing emails, two emails are equals when their value and type
	// equal.
	IsIdentity bool `json:"isIdentity"`

	// True if the property is an exclusive property. Exclusive properties
	// are boolean properties within a multi-valued complex property that
	// has true value. Among all the values in this multi-valued complex
	// property, at most one true value can exist at the same time. This
	// field controls this invariant, as to ensure when new value with a
	// true exclusive property is added, the existing true exclusive property
	// is turned off (set to false).
	IsExclusive bool `json:"isExclusive"`

	// Annotations mark the property for more flexible, extensible transformations. It
	// is to be used in conjunction with annotation processors. This is the extension
	// point for the capabilities of this project. For instance, a property could be
	// marked as '@uuid', and an recognizing annotation processor will generate a uuid
	// for this field when necessary.
	Annotations []string `json:"annotations"`
}

// validate the metadata
func (md *Metadata) validate() error {
	if len(md.Path) == 0 {
		return Errors.internal("path is required for metadata")
	}

	return nil
}

// Create a (deep) copy of the metadata
func (md *Metadata) copy() *Metadata {
	var annotations []string
	{
		if len(md.Annotations) > 0 {
			annotations = make([]string, 0)
			for _, v := range md.Annotations {
				annotations = append(annotations, v)
			}
		}
	}

	return &Metadata{
		Path:        md.Path,
		DbAlias:     md.DbAlias,
		IsIdentity:  md.IsIdentity,
		IsExclusive: md.IsExclusive,
		Annotations: annotations,
	}
}
