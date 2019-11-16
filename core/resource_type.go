package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	// An in memory repository to cache all resource types. It is not thread safe and is
	// intended to used as read-only after initial setup.
	resourceTypeRepository struct {
		mem map[string]*ResourceType
	}
)

var (
	// Entry point to access resource type repository
	ResourceTypes = &resourceTypeRepository{mem: make(map[string]*ResourceType)}
)

// Get all attributes available in this resource type. The attributes from the base schema are simply included.
// All attributes from schema extensions are nested under a complex attribute named with the schema extension's id.
func (rt *ResourceType) DerivedAttributes() []*Attribute {
	if len(rt.attrs) == 0 {
		rt.attrs = make([]*Attribute, 0)
		rt.attrs = append(rt.attrs, ScimCoreSchema.Attributes...)
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
		panic(Errors.Internal(fmt.Sprintf("no such schema by id '%s'", rt.Schema)))
	}
	return
}

// Get the schema object from repository by the schema id defined in
// SchemaExtension.Schema, or panic if such schema does not exist.
func (ext *SchemaExtension) MustSchema() (schema *Schema) {
	schema = Schemas.Get(ext.Schema)
	if schema == nil {
		panic(Errors.Internal(fmt.Sprintf("no such schema by id '%s'", ext.Schema)))
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
		Name:          schema.Id,
		Type:          TypeComplex,
		Required:      ext.Required,
		Mutability:    MutabilityReadWrite,
		Returned:      ReturnedDefault,
		Uniqueness:    UniquenessNone,
		SubAttributes: schema.Attributes,
	}
}

// Load a resource type from file, or panic.
func (r *resourceTypeRepository) MustLoad(filePath string) *ResourceType {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	resourceType, err := ParseResourceType(raw)
	if err != nil {
		panic(err)
	}

	return resourceType
}

// Add a resource type to repository.
func (r *resourceTypeRepository) Add(resourceType *ResourceType) {
	if resourceType != nil && len(resourceType.Id) > 0 {
		r.mem[resourceType.Id] = resourceType
	}
}

// Get resource type from repository by its id, or nil if it does not exist.
func (r *resourceTypeRepository) Get(resourceTypeId string) *ResourceType {
	return r.mem[resourceTypeId]
}

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
