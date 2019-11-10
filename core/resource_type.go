package core

import (
	"encoding/json"
	"fmt"
)

type (
	// A SCIM resource type. The schema field is just an id.
	ResourceType struct {
		Id               string             `json:"id"`
		Name             string             `json:"name"`
		Description      string             `json:"description"`
		Endpoint         string             `json:"endpoint"`
		Schema           string             `json:"schema"`
		SchemaExtensions []*SchemaExtension `json:"schemaExtensions"`

		// A transient cache of all the attributes derived from the base schema and schema extensions.
		// It is cached here because deriveAttributes() is potentially expensive and the result
		// does not change.
		attrs []*Attribute `json:"-"`
	}
	// A SCIM schema extension.
	SchemaExtension struct {
		Schema   string `json:"schema"`
		Required bool   `json:"required"`
	}
)

// Parse resource type from JSON. This method will attempt to derive attributes from schemas in this resource type.
// Hence, it relies on the relevant schemas being parsed and saved in repository already.
func ParseResourceType(raw []byte) (rt *ResourceType, err error) {
	rt = new(ResourceType)
	err = json.Unmarshal(raw, &rt)
	if err != nil {
		err = Errors.InvalidSyntax("invalid resource type JSON definition")
		return
	}

	_ = rt.DerivedAttributes()
	return
}

// Get all attributes available in this resource type. The attributes from the base schema are simply included.
// All attributes from schema extensions are nested under a complex attribute named with the schema extension's id.
func (rt *ResourceType) DerivedAttributes() []*Attribute {
	if len(rt.attrs) == 0 {
		rt.attrs = make([]*Attribute, 0)
		rt.attrs = append(rt.attrs, rt.MustSchema().Attributes...)
		for _, ext := range rt.SchemaExtensions {
			rt.attrs = append(rt.attrs, ext.AsAttributeContainer())
		}
	}
	return rt.attrs
}

// Get the schema object from repository by the schema id defined in
// ResourceType.Schema, or panic if such schema does not exist.
func (rt *ResourceType) MustSchema() (schema *Schema) {
	schema = Schemas.Get(rt.Schema)
	if schema == nil {
		panic(Errors.internal(fmt.Sprintf("no such schema by id '%s'", rt.Schema)))
	}
	return
}

// Get the schema object from repository by the schema id defined in
// SchemaExtension.Schema, or panic if such schema does not exist.
func (ext *SchemaExtension) MustSchema() (schema *Schema) {
	schema = Schemas.Get(ext.Schema)
	if schema == nil {
		panic(Errors.internal(fmt.Sprintf("no such schema by id '%s'", ext.Schema)))
	}
	return
}

// Create a new complex attribute which hosts all schema attributes as its sub-attributes.
// After inserting the attributes down a level, their path metadata is supposed to be prepended
// with the name of the container (i.e. employeeNumber -> urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber).
// However, since we are not actively using the path metadata here, it seemly unnecessary to do so. Hence, just keep
// in mind this little inconsistency.
func (ext *SchemaExtension) AsAttributeContainer() *Attribute {
	schema := ext.MustSchema()
	return &Attribute{
		Name:       schema.Id,
		Type:       TypeComplex,
		Required:   ext.Required,
		Mutability: MutabilityReadWrite,
		Returned:   ReturnedDefault,
		Uniqueness: UniquenessNone,
		Metadata: &Metadata{
			Path:    schema.Id,
			DbAlias: schema.Companion.DbAlias,
		},
		SubAttributes: schema.Attributes,
	}
}
