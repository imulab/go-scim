package internal

import (
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// SerializableSchema is the json.Serializable wrapper for spec.Schema.
type SerializableSchema struct {
	Sch *spec.Schema
}

// MainSchemaId returns the main schema id for Schema as a resource.
func (s *SerializableSchema) MainSchemaId() string {
	return "urn:ietf:params:scim:schemas:core:2.0:Schema"
}

// Visit takes the visitor on a DFS tour of the structure of the Schema resource. This method takes control of what
// to visit and will not consult the ShouldVisit method of visitor.
func (s *SerializableSchema) Visit(visitor prop.Visitor) error {
	dummyContainer := prop.NewComplex(spec.MetaAttributes().SchemaAttributeNoSub())
	visitor.BeginChildren(dummyContainer)
	if err := s.visitCore(visitor); err != nil {
		return err
	}
	if err := s.visitSchema(visitor); err != nil {
		return err
	}
	visitor.EndChildren(dummyContainer)
	return nil
}

func (s *SerializableSchema) visitCore(visitor prop.Visitor) error {
	// schemas
	schemas := prop.NewMultiOf(spec.MetaAttributes().CoreSchemasAttribute(), []interface{}{s.MainSchemaId()})
	if err := visitor.Visit(schemas); err != nil {
		return err
	}
	visitor.BeginChildren(schemas)
	if err := schemas.ForEachChild(func(index int, child prop.Property) error {
		return visitor.Visit(child)
	}); err != nil {
		return err
	}
	visitor.EndChildren(schemas)

	// id
	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().CoreIdAttribute(), s.Sch.ID())); err != nil {
		return err
	}

	// meta
	meta := prop.NewComplexOf(spec.MetaAttributes().CoreMetaPartialAttribute(), map[string]interface{}{
		"resourceType": s.Sch.ResourceTypeName(),
		"location":     s.Sch.ResourceLocation(),
	})
	if err := visitor.Visit(meta); err != nil {
		return err
	}
	visitor.BeginChildren(meta)
	if err := meta.ForEachChild(func(_ int, child prop.Property) error {
		return visitor.Visit(child)
	}); err != nil {
		return err
	}
	visitor.EndChildren(meta)

	return nil
}

func (s *SerializableSchema) visitSchema(visitor prop.Visitor) error {
	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().SchemaNameAttribute(), s.Sch.Name())); err != nil {
		return err
	}

	if description := s.Sch.Description(); len(description) > 0 {
		if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().SchemaDescriptionAttribute(), description)); err != nil {
			return err
		}
	}

	dummyMulti := prop.NewMulti(spec.MetaAttributes().SchemaAttributesAttributeNoSub())
	if err := visitor.Visit(dummyMulti); err != nil {
		return err
	}
	visitor.BeginChildren(dummyMulti)
	if err := s.Sch.ForEachAttribute(func(attr *spec.Attribute) error {
		dummyComplex := prop.NewComplex(spec.MetaAttributes().SchemaAttributesAttributeNoSub().DeriveElementAttribute())
		if err := visitor.Visit(dummyComplex); err != nil {
			return err
		}
		visitor.BeginChildren(dummyComplex)
		if err := s.visitAttribute(attr, visitor); err != nil {
			return err
		}
		visitor.EndChildren(dummyComplex)
		return nil
	}); err != nil {
		return err
	}
	visitor.EndChildren(dummyMulti)

	return nil
}

func (s *SerializableSchema) visitAttribute(attr *spec.Attribute, visitor prop.Visitor) error {
	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().AttributeNameAttribute(), attr.Name())); err != nil {
		return err
	}

	if description := attr.Description(); len(description) > 0 {
		if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().AttributeDescriptionAttribute(), description)); err != nil {
			return err
		}
	}

	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().AttributeTypeAttribute(), attr.Type().String())); err != nil {
		return err
	}

	if err := visitor.Visit(prop.NewBooleanOf(spec.MetaAttributes().AttributeMultiValuedAttribute(), attr.MultiValued())); err != nil {
		return err
	}

	if err := visitor.Visit(prop.NewBooleanOf(spec.MetaAttributes().AttributeRequiredAttribute(), attr.Required())); err != nil {
		return err
	}

	if err := visitor.Visit(prop.NewBooleanOf(spec.MetaAttributes().AttributeCaseExactAttribute(), attr.CaseExact())); err != nil {
		return err
	}

	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().AttributeMutabilityAttribute(), attr.Mutability().String())); err != nil {
		return err
	}

	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().AttributeReturnedAttribute(), attr.Returned().String())); err != nil {
		return err
	}

	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().AttributeUniquenessAttribute(), attr.Uniqueness().String())); err != nil {
		return err
	}

	if attr.CountCanonicalValues() > 0 {
		var canonicalValues []interface{}
		attr.ForEachCanonicalValues(func(canonicalValue string) {
			canonicalValues = append(canonicalValues, canonicalValue)
		})

		cvp := prop.NewMultiOf(spec.MetaAttributes().AttributeCanonicalValuesAttribute(), canonicalValues)
		if err := visitor.Visit(cvp); err != nil {
			return err
		}
		visitor.BeginChildren(cvp)
		if err := cvp.ForEachChild(func(_ int, child prop.Property) error {
			return visitor.Visit(child)
		}); err != nil {
			return err
		}
		visitor.EndChildren(cvp)
	}

	if attr.CountReferenceTypes() > 0 {
		var referenceTypes []interface{}
		attr.ForEachReferenceTypes(func(referenceType string) {
			referenceTypes = append(referenceTypes, referenceType)
		})

		rtp := prop.NewMultiOf(spec.MetaAttributes().AttributeReferenceTypesAttribute(), referenceTypes)
		if err := visitor.Visit(rtp); err != nil {
			return err
		}
		visitor.BeginChildren(rtp)
		if err := rtp.ForEachChild(func(_ int, child prop.Property) error {
			return visitor.Visit(child)
		}); err != nil {
			return err
		}
		visitor.EndChildren(rtp)
	}

	if attr.CountSubAttributes() > 0 {
		dummyMulti := prop.NewMulti(spec.MetaAttributes().AttributeSubAttributesAttributeNoSub())
		if err := visitor.Visit(dummyMulti); err != nil {
			return err
		}
		visitor.BeginChildren(dummyMulti)
		if err := attr.ForEachSubAttribute(func(subAttribute *spec.Attribute) error {
			dummyComplex := prop.NewComplex(spec.MetaAttributes().AttributeSubAttributesAttributeNoSub().DeriveElementAttribute())
			if err := visitor.Visit(dummyComplex); err != nil {
				return err
			}
			visitor.BeginChildren(dummyComplex)
			if err := s.visitAttribute(subAttribute, visitor); err != nil {
				return err
			}
			visitor.EndChildren(dummyComplex)
			return nil
		}); err != nil {
			return err
		}
		visitor.EndChildren(dummyMulti)
	}

	return nil
}
