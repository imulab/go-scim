package scratch

import "encoding/json"

type Schema struct {
	id          string
	name        string
	description string
	attrs       []*Attribute
	required    bool
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	type schemaJSON struct {
		Id          string       `json:"id"`
		Name        string       `json:"name"`
		Description string       `json:"description,omitempty"`
		Attributes  []*Attribute `json:"attributes,omitempty"`
	}

	return json.Marshal(schemaJSON{
		Id:          s.id,
		Name:        s.name,
		Description: s.description,
		Attributes:  s.attrs,
	})
}

func SchemaBuilder(id string) *schemaDsl {
	return &schemaDsl{
		Schema: &Schema{id: id},
	}
}

type schemaDsl struct {
	*Schema
}

func (d *schemaDsl) Name(name string) *schemaDsl {
	d.name = name
	return d
}

func (d *schemaDsl) Description(description string) *schemaDsl {
	d.description = description
	return d
}

func (d *schemaDsl) Attributes(attrs ...*Attribute) *schemaDsl {
	d.attrs = append(d.attrs, attrs...)
	return d
}

func (d *schemaDsl) RequiredAsExtension() *schemaDsl {
	d.required = true
	return d
}

func (d *schemaDsl) Build() *Schema {
	switch {
	case len(d.id) == 0:
		panic("id is required")
	case len(d.name) == 0:
		panic("name is required")
	case len(d.attrs) == 0:
		panic("at least one attribute is required")
	default:
		return d.Schema
	}
}

var (
	coreSchema = SchemaBuilder("core").
			Name("Core").
			Description("Shared attributes for all SCIM resources").
			Attributes(
			ReferenceAttribute("schemas").MultiValued().Required().CaseExact().AlwaysReturn().Build(),
			StringAttribute("id").CaseExact().AlwaysReturn().ReadOnly().UniqueGlobally().Build(),
			StringAttribute("externalId").Build(),
			ComplexAttribute("meta").ReadOnly().WithSubAttributes(
				StringAttribute("resourceType").CaseExact().ReadOnly().Build(),
				DateTimeAttribute("created").ReadOnly().Build(),
				DateTimeAttribute("lastModified").ReadOnly().Build(),
				ReferenceAttribute("location").CaseExact().ReadOnly().Build(),
				StringAttribute("version").ReadOnly().Build(),
			).Build(),
		).Build()

	UserSchema = SchemaBuilder("urn:ietf:params:scim:schemas:core:2.0:User").
			Name("User").
			Description("Defined attributes for the user schema").
			Attributes(
			StringAttribute("userName").Required().UniqueOnServer().Build(),
			ComplexAttribute("name").WithSubAttributes(
				StringAttribute("formatted").Build(),
				StringAttribute("familyName").Build(),
				StringAttribute("givenName").Build(),
				StringAttribute("middleName").Build(),
				StringAttribute("honorificPrefix").Build(),
				StringAttribute("honorificSuffix").Build(),
			).Build(),
			StringAttribute("displayName").Build(),
			StringAttribute("nickName").Build(),
			ReferenceAttribute("profileUrl").ReferenceTypes("external").Build(),
			StringAttribute("title").Build(),
			StringAttribute("userType").Build(),
			StringAttribute("preferredLanguage").Build(),
			StringAttribute("locale").Build(),
			StringAttribute("timezone").Build(),
			BooleanAttribute("active").Build(),
			StringAttribute("password").WriteOnly().NeverReturn().Build(),
			ComplexAttribute("emails").MultiValued().Required().WithSubAttributes(
				StringAttribute("value").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
				StringAttribute("display").Build(),
			).Build(),
			ComplexAttribute("phoneNumbers").MultiValued().WithSubAttributes(
				StringAttribute("value").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
				StringAttribute("display").Build(),
			).Build(),
			ComplexAttribute("ims").MultiValued().WithSubAttributes(
				StringAttribute("value").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
				StringAttribute("display").Build(),
			).Build(),
			ComplexAttribute("photos").MultiValued().WithSubAttributes(
				ReferenceAttribute("value").ReferenceTypes("external").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
			).Build(),
			ComplexAttribute("addresses").MultiValued().WithSubAttributes(
				StringAttribute("formatted").MarkAsIdentity().Build(),
				StringAttribute("streetAddress").MarkAsIdentity().Build(),
				StringAttribute("locality").MarkAsIdentity().Build(),
				StringAttribute("region").MarkAsIdentity().Build(),
				StringAttribute("postalCode").MarkAsIdentity().Build(),
				StringAttribute("country").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
			).Build(),
			ComplexAttribute("groups").MultiValued().ReadOnly().WithSubAttributes(
				StringAttribute("value").ReadOnly().MarkAsIdentity().Build(),
				ReferenceAttribute("$ref").ReadOnly().Build(),
				StringAttribute("type").CanonicalValues("direct", "indirect").ReadOnly().Build(),
				StringAttribute("display").ReadOnly().Build(),
			).Build(),
			ComplexAttribute("entitlements").MultiValued().WithSubAttributes(
				StringAttribute("value").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
				StringAttribute("display").Build(),
			).Build(),
			ComplexAttribute("roles").MultiValued().WithSubAttributes(
				StringAttribute("value").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
				StringAttribute("display").Build(),
			).Build(),
			ComplexAttribute("x509Certificates").MultiValued().WithSubAttributes(
				StringAttribute("value").MarkAsIdentity().Build(),
				StringAttribute("type").MarkAsIdentity().Build(),
				BooleanAttribute("primary").MarkAsPrimary().Build(),
				StringAttribute("display").Build(),
			).Build(),
		).Build()

	UserEnterpriseSchemaExtension = SchemaBuilder("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User").
					Name("Enterprise User").
					Description("Extension attributes for enterprise users").
					Attributes(
			StringAttribute("employeeNumber").Build(),
			StringAttribute("costCenter").Build(),
			StringAttribute("organization").Build(),
			StringAttribute("division").Build(),
			StringAttribute("department").Build(),
			ComplexAttribute("manager").WithSubAttributes(
				StringAttribute("value").MarkAsIdentity().Build(),
				ReferenceAttribute("$ref").Build(),
				StringAttribute("displayName").Build(),
			).Build(),
		).Build()

	GroupSchema = SchemaBuilder("urn:ietf:params:scim:schemas:core:2.0:Group").
			Name("Group").
			Description("Defined attributes for the group schema").
			Attributes(
			StringAttribute("displayName").Build(),
			ComplexAttribute("members").MultiValued().WithSubAttributes(
				StringAttribute("value").Immutable().MarkAsIdentity().Build(),
				ReferenceAttribute("$ref").Immutable().MarkAsIdentity().Build(),
				StringAttribute("display").Build(),
			).Build(),
		).Build()
)
