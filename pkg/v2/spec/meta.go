package spec

var (
	metaAttributes = &metaAttr{}
)

// MetaAttributes returns a structure to access individual attributes about
// fields in Schema, Attribute and ResourceType. These attributes are known
// as meta attributes, because they describe things that are used to describe
// other resources.
func MetaAttributes() *metaAttr {
	return metaAttributes
}

type metaAttr struct {
	coreSchemas *Attribute
	coreId      *Attribute
	coreMeta    *Attribute

	schema            *Attribute
	schemaName        *Attribute
	schemaDescription *Attribute
	schemaAttributes  *Attribute

	attrName            *Attribute
	attrDescription     *Attribute
	attrType            *Attribute
	attrMultiValued     *Attribute
	attrRequired        *Attribute
	attrCaseExact       *Attribute
	attrMutability      *Attribute
	attrReturned        *Attribute
	attrUniqueness      *Attribute
	attrCanonicalValues *Attribute
	attrReferenceTypes  *Attribute
	attrSubAttributes   *Attribute

	resourceType                        *Attribute
	resourceTypeName                    *Attribute
	resourceTypeDescription             *Attribute
	resourceTypeEndpoint                *Attribute
	resourceTypeSchema                  *Attribute
	resourceTypeSchemaExtensions        *Attribute
	resourceTypeSchemaExtensionSchema   *Attribute
	resourceTypeSchemaExtensionRequired *Attribute
}

// CoreSchemasAttribute returns an attribute to describe the "schemas" field.
func (m *metaAttr) CoreSchemasAttribute() *Attribute {
	if m.coreSchemas == nil {
		m.coreSchemas = &Attribute{
			id:          "schemas",
			name:        "schemas",
			typ:         TypeString,
			multiValued: true,
			index:       0,
			path:        "schemas",
		}
	}
	return m.coreSchemas
}

// CoreIdAttribute returns an attribute to describe the "id" field.
func (m *metaAttr) CoreIdAttribute() *Attribute {
	if m.coreId == nil {
		m.coreId = &Attribute{
			id:    "id",
			name:  "id",
			typ:   TypeString,
			index: 1,
			path:  "id",
		}
	}
	return m.coreId
}

// CoreMetaPartialAttribute returns an attribute to describe the "meta" field. Only the "resourceType" and "location"
// sub attributes are included as subAttributes.
func (m *metaAttr) CoreMetaPartialAttribute() *Attribute {
	if m.coreMeta == nil {
		m.coreMeta = &Attribute{
			id:    "meta",
			name:  "meta",
			typ:   TypeComplex,
			index: 2,
			path:  "meta",
			subAttributes: []*Attribute{
				{
					id:    "meta.resourceType",
					name:  "resourceType",
					typ:   TypeString,
					index: 0,
					path:  "meta.resourceType",
				},
				{
					id:    "meta.location",
					name:  "location",
					typ:   TypeString,
					index: 1,
					path:  "meta.location",
				},
			},
		}
	}
	return m.coreMeta
}

// SchemaAttributeNoSub returns an attribute to act as the container attribute for schema resources, but it does not
// actually contain any sub attributes.
func (m *metaAttr) SchemaAttributeNoSub() *Attribute {
	if m.schema == nil {
		m.schema = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema",
			name:  "",
			typ:   TypeComplex,
			index: 0,
			path:  "",
		}
	}
	return m.schema
}

// SchemaNameAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:name" field.
func (m *metaAttr) SchemaNameAttribute() *Attribute {
	if m.schemaName == nil {
		m.schemaName = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:name",
			name:  "name",
			typ:   TypeString,
			index: 100,
			path:  "name",
		}
	}
	return m.schemaName
}

// SchemaDescriptionAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:description" field.
func (m *metaAttr) SchemaDescriptionAttribute() *Attribute {
	if m.schemaDescription == nil {
		m.schemaDescription = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:description",
			name:  "description",
			typ:   TypeString,
			index: 101,
			path:  "description",
		}
	}
	return m.schemaDescription
}

// SchemaAttributesAttributeNoSub returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes".
// However, this attribute does not contain the subAttributes definition.
func (m *metaAttr) SchemaAttributesAttributeNoSub() *Attribute {
	if m.schemaAttributes == nil {
		m.schemaAttributes = &Attribute{
			id:          "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes",
			name:        "attributes",
			typ:         TypeComplex,
			multiValued: true,
			index:       102,
			path:        "attributes",
		}
	}
	return m.schemaAttributes
}

// AttributeNameAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.name".
func (m *metaAttr) AttributeNameAttribute() *Attribute {
	if m.attrName == nil {
		m.attrName = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.name",
			name:  "name",
			typ:   TypeString,
			index: 0,
			path:  "attributes.name",
		}
	}
	return m.attrName
}

// AttributeDescriptionAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.description".
func (m *metaAttr) AttributeDescriptionAttribute() *Attribute {
	if m.attrDescription == nil {
		m.attrDescription = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.description",
			name:  "description",
			typ:   TypeString,
			index: 1,
			path:  "attributes.description",
		}
	}
	return m.attrDescription
}

// AttributeTypeAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.type".
func (m *metaAttr) AttributeTypeAttribute() *Attribute {
	if m.attrType == nil {
		m.attrType = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.type",
			name:  "type",
			typ:   TypeString,
			index: 2,
			path:  "attributes.type",
			canonicalValues: []string{
				TypeString.String(),
				TypeInteger.String(),
				TypeDecimal.String(),
				TypeBoolean.String(),
				TypeDateTime.String(),
				TypeBinary.String(),
				TypeReference.String(),
				TypeComplex.String(),
			},
		}
	}
	return m.attrType
}

// AttributeMultiValuedAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.multiValued".
func (m *metaAttr) AttributeMultiValuedAttribute() *Attribute {
	if m.attrMultiValued == nil {
		m.attrMultiValued = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.multiValued",
			name:  "multiValued",
			typ:   TypeBoolean,
			index: 3,
			path:  "attributes.multiValued",
		}
	}
	return m.attrMultiValued
}

// AttributeRequiredAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.required".
func (m *metaAttr) AttributeRequiredAttribute() *Attribute {
	if m.attrRequired == nil {
		m.attrRequired = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.required",
			name:  "required",
			typ:   TypeBoolean,
			index: 4,
			path:  "attributes.required",
		}
	}
	return m.attrRequired
}

// AttributeCaseExactAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.caseExact".
func (m *metaAttr) AttributeCaseExactAttribute() *Attribute {
	if m.attrCaseExact == nil {
		m.attrCaseExact = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.caseExact",
			name:  "caseExact",
			typ:   TypeBoolean,
			index: 5,
			path:  "attributes.caseExact",
		}
	}
	return m.attrCaseExact
}

// AttributeMutabilityAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.mutability".
func (m *metaAttr) AttributeMutabilityAttribute() *Attribute {
	if m.attrMutability == nil {
		m.attrMutability = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.mutability",
			name:  "mutability",
			typ:   TypeString,
			index: 6,
			path:  "attributes.mutability",
			canonicalValues: []string{
				MutabilityReadWrite.String(),
				MutabilityReadOnly.String(),
				MutabilityImmutable.String(),
				MutabilityWriteOnly.String(),
			},
		}
	}
	return m.attrMutability
}

// AttributeReturnedAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.returned".
func (m *metaAttr) AttributeReturnedAttribute() *Attribute {
	if m.attrReturned == nil {
		m.attrReturned = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.returned",
			name:  "returned",
			typ:   TypeString,
			index: 7,
			path:  "attributes.returned",
			canonicalValues: []string{
				ReturnedDefault.String(),
				ReturnedAlways.String(),
				ReturnedNever.String(),
				ReturnedRequest.String(),
			},
		}
	}
	return m.attrReturned
}

// AttributeUniquenessAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.uniqueness".
func (m *metaAttr) AttributeUniquenessAttribute() *Attribute {
	if m.attrUniqueness == nil {
		m.attrUniqueness = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.uniqueness",
			name:  "uniqueness",
			typ:   TypeString,
			index: 8,
			path:  "attributes.uniqueness",
			canonicalValues: []string{
				UniquenessNone.String(),
				UniquenessServer.String(),
				UniquenessGlobal.String(),
			},
		}
	}
	return m.attrUniqueness
}

// AttributeCanonicalValuesAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.canonicalValues".
func (m *metaAttr) AttributeCanonicalValuesAttribute() *Attribute {
	if m.attrCanonicalValues == nil {
		m.attrCanonicalValues = &Attribute{
			id:          "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.canonicalValues",
			name:        "canonicalValues",
			typ:         TypeString,
			multiValued: true,
			index:       9,
			path:        "attributes.canonicalValues",
		}
	}
	return m.attrCanonicalValues
}

// AttributeReferenceTypesAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.referenceTypes".
func (m *metaAttr) AttributeReferenceTypesAttribute() *Attribute {
	if m.attrReferenceTypes == nil {
		m.attrReferenceTypes = &Attribute{
			id:          "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.referenceTypes",
			name:        "referenceTypes",
			typ:         TypeReference,
			multiValued: true,
			index:       10,
			path:        "attributes.referenceTypes",
		}
	}
	return m.attrReferenceTypes
}

// AttributeSubAttributesAttributeNoSub returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.subAttributes".
// However, this attribute does not contain the subAttributes definition.
func (m *metaAttr) AttributeSubAttributesAttributeNoSub() *Attribute {
	if m.attrSubAttributes == nil {
		m.attrSubAttributes = &Attribute{
			id:          "urn:ietf:params:scim:schemas:core:2.0:Schema:attributes.subAttributes",
			name:        "subAttributes",
			typ:         TypeComplex,
			multiValued: true,
			index:       11,
			path:        "attributes.subAttributes",
		}
	}
	return m.attrSubAttributes
}

// ResourceTypeAttributeNoSub returns an attribute to act as the container attribute for ResourceType resources, but it does not
// actually contain any sub attributes.
func (m *metaAttr) ResourceTypeAttributeNoSub() *Attribute {
	if m.resourceType == nil {
		m.resourceType = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:ResourceType",
			name:  "",
			typ:   TypeComplex,
			index: 0,
			path:  "",
		}
	}
	return m.resourceType
}

// ResourceTypeNameAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:ResourceType:name"
func (m *metaAttr) ResourceTypeNameAttribute() *Attribute {
	if m.resourceTypeName == nil {
		m.resourceTypeName = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:ResourceType:name",
			name:  "name",
			typ:   TypeString,
			index: 100,
			path:  "name",
		}
	}
	return m.resourceTypeName
}

// ResourceTypeDescriptionAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:ResourceType:description".
func (m *metaAttr) ResourceTypeDescriptionAttribute() *Attribute {
	if m.resourceTypeDescription == nil {
		m.resourceTypeDescription = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:ResourceType:description",
			name:  "description",
			typ:   TypeString,
			index: 101,
			path:  "description",
		}
	}
	return m.resourceTypeDescription
}

// ResourceTypeEndpointAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:ResourceType:endpoint".
func (m *metaAttr) ResourceTypeEndpointAttribute() *Attribute {
	if m.resourceTypeEndpoint == nil {
		m.resourceTypeEndpoint = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:ResourceType:endpoint",
			name:  "endpoint",
			typ:   TypeString,
			index: 102,
			path:  "endpoint",
		}
	}
	return m.resourceTypeEndpoint
}

// ResourceTypeSchemaAttribute returns an attribute to describe "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schema".
func (m *metaAttr) ResourceTypeSchemaAttribute() *Attribute {
	if m.resourceTypeSchema == nil {
		m.resourceTypeSchema = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schema",
			name:  "schema",
			typ:   TypeReference,
			index: 103,
			path:  "schema",
		}
	}
	return m.resourceTypeSchema
}

// ResourceTypeSchemaExtensionsAttributeNoSub returns an attribute to describe
// "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schemaExtensions", but the returned attribute
// does not have any subAttributes defined.
func (m *metaAttr) ResourceTypeSchemaExtensionsAttributeNoSub() *Attribute {
	if m.resourceTypeSchemaExtensions == nil {
		m.resourceTypeSchemaExtensions = &Attribute{
			id:          "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schemaExtensions",
			name:        "schemaExtensions",
			typ:         TypeComplex,
			multiValued: true,
			index:       104,
			path:        "schemaExtensions",
		}
	}
	return m.resourceTypeSchemaExtensions
}

// ResourceTypeSchemaExtensionSchemaAttribute returns an attribute to describe
// "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schemaExtensions.schema".
func (m *metaAttr) ResourceTypeSchemaExtensionSchemaAttribute() *Attribute {
	if m.resourceTypeSchemaExtensionSchema == nil {
		m.resourceTypeSchemaExtensionSchema = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schemaExtensions.schema",
			name:  "schema",
			typ:   TypeReference,
			index: 0,
			path:  "schemaExtensions.schema",
		}
	}
	return m.resourceTypeSchemaExtensionSchema
}

// ResourceTypeSchemaExtensionRequiredAttribute returns an attribute to describe
// "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schemaExtensions.required".
func (m *metaAttr) ResourceTypeSchemaExtensionRequiredAttribute() *Attribute {
	if m.resourceTypeSchemaExtensionRequired == nil {
		m.resourceTypeSchemaExtensionRequired = &Attribute{
			id:    "urn:ietf:params:scim:schemas:core:2.0:ResourceType:schemaExtensions.required",
			name:  "required",
			typ:   TypeBoolean,
			index: 1,
			path:  "schemaExtensions.required",
		}
	}
	return m.resourceTypeSchemaExtensionRequired
}
