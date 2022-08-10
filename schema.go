package scim

import "encoding/json"

// BuildSchema builds and returns a new Schema.
func BuildSchema(builder func(d *schemaDsl)) *Schema {
	b := new(schemaDsl)
	builder(b)
	return b.build()
}

// Schema is a collection of Attribute.
type Schema struct {
	id          string
	name        string
	description string
	attrs       []*Attribute
	required    bool
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(schemaJSON{
		Id:          s.id,
		Name:        s.name,
		Description: s.description,
		Attributes:  s.attrs,
	})
}

type schemaJSON struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Attributes  []*Attribute `json:"attributes,omitempty"`
}

type schemaDsl Schema

func (d *schemaDsl) ID(id string) *schemaDsl {
	d.id = id
	return d
}

func (d *schemaDsl) Name(name string) *schemaDsl {
	d.name = name
	return d
}

func (d *schemaDsl) Describe(text string) *schemaDsl {
	d.description = text
	return d
}

func (d *schemaDsl) WithAttributes(fn func(d *attributeListDsl)) *schemaDsl {
	if len(d.id) == 0 {
		panic("set id first")
	}

	sd := &attributeListDsl{namespace: d.id}
	fn(sd)

	for _, it := range sd.list {
		d.attrs = append(d.attrs, it.build())
	}

	return d
}

func (d *schemaDsl) build() *Schema {
	return (*Schema)(d)
}

var (
	UserSchema = func(d *schemaDsl) {
		d.ID("urn:ietf:params:scim:schemas:core:2.0:User").
			Name("User").
			Describe("Defined attributes for the user schema").
			WithAttributes(func(d *attributeListDsl) {
				d.Add(func(d *attributeDsl) {
					d.Name("userName").String().Required().UniqueOnServer()
				}).Add(func(d *attributeDsl) {
					d.Name("name").Complex().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("formatted").String()
						}).Add(func(d *attributeDsl) {
							d.Name("familyName").String()
						}).Add(func(d *attributeDsl) {
							d.Name("givenName").String()
						}).Add(func(d *attributeDsl) {
							d.Name("middleName").String()
						}).Add(func(d *attributeDsl) {
							d.Name("honorificPrefix").String()
						}).Add(func(d *attributeDsl) {
							d.Name("honorificSuffix").String()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("displayName").String()
				}).Add(func(d *attributeDsl) {
					d.Name("nickName").String()
				}).Add(func(d *attributeDsl) {
					d.Name("profileUrl").Reference().ReferenceTypes("external")
				}).Add(func(d *attributeDsl) {
					d.Name("title").String()
				}).Add(func(d *attributeDsl) {
					d.Name("userType").String()
				}).Add(func(d *attributeDsl) {
					d.Name("preferredLanguage").String()
				}).Add(func(d *attributeDsl) {
					d.Name("locale").String()
				}).Add(func(d *attributeDsl) {
					d.Name("timezone").String()
				}).Add(func(d *attributeDsl) {
					d.Name("active").Boolean()
				}).Add(func(d *attributeDsl) {
					d.Name("password").String().WriteOnly().NeverReturn()
				}).Add(func(d *attributeDsl) {
					d.Name("emails").Complex().MultiValued().Required().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("phoneNumbers").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("ims").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("photos").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").Reference().ReferenceTypes("external").Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("addresses").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("formatted").String()
						}).Add(func(d *attributeDsl) {
							d.Name("streetAddress").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("locality").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("region").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("postalCode").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("country").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("groups").Complex().MultiValued().ReadOnly().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String().ReadOnly().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("$ref").Reference().ReadOnly().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().ReadOnly().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String().ReadOnly()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("entitlements").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("roles").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String()
						})
					})
				}).Add(func(d *attributeDsl) {
					d.Name("x509Certificates").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").Binary().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("type").String().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("primary").Boolean().Primary()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String()
						})
					})
				})
			})
	}

	UserEnterpriseSchemaExtension = func(d *schemaDsl) {
		d.ID("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User").
			Name("Enterprise User").
			Describe("Extension attributes for enterprise users").
			WithAttributes(func(d *attributeListDsl) {
				d.Add(func(d *attributeDsl) {
					d.Name("employeeNumber").String()
				}).Add(func(d *attributeDsl) {
					d.Name("costCenter").String()
				}).Add(func(d *attributeDsl) {
					d.Name("organization").String()
				}).Add(func(d *attributeDsl) {
					d.Name("division").String()
				}).Add(func(d *attributeDsl) {
					d.Name("department").String()
				}).Add(func(d *attributeDsl) {
					d.Name("manager").Complex().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String()
						}).Add(func(d *attributeDsl) {
							d.Name("$ref").Reference()
						}).Add(func(d *attributeDsl) {
							d.Name("displayName").String()
						})
					})
				})
			})
	}

	GroupSchema = func(d *schemaDsl) {
		d.ID("urn:ietf:params:scim:schemas:core:2.0:Group").
			Name("Group").
			Describe("Defined attributes for the group schema").
			WithAttributes(func(d *attributeListDsl) {
				d.Add(func(d *attributeDsl) {
					d.Name("displayName").String()
				}).Add(func(d *attributeDsl) {
					d.Name("members").Complex().MultiValued().SubAttributes(func(sd *attributeListDsl) {
						sd.Add(func(d *attributeDsl) {
							d.Name("value").String().Immutable().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("$ref").Reference().Immutable().Identity()
						}).Add(func(d *attributeDsl) {
							d.Name("display").String().Immutable()
						})
					})
				})
			})
	}
)
