package internal

import (
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// SerializableResourceType is the json.Serializable wrapper for spec.ResourceType.
type SerializableResourceType struct {
	ResourceType *spec.ResourceType
}

// MainSchemaId returns the main schema id for ResourceType as a resource.
func (s *SerializableResourceType) MainSchemaId() string {
	return "urn:ietf:params:scim:schemas:core:2.0:ResourceType"
}

// Visit takes the visitor on a DFS tour of the structure of the ResourceType resource. This method takes control of what
// to visit and will not consult the ShouldVisit method of visitor.
func (s *SerializableResourceType) Visit(visitor prop.Visitor) error {
	dummyContainer := prop.NewComplex(spec.MetaAttributes().ResourceTypeAttributeNoSub())
	visitor.BeginChildren(dummyContainer)
	if err := s.visitCore(visitor); err != nil {
		return err
	}
	if err := s.visitResourceType(visitor); err != nil {
		return err
	}
	visitor.EndChildren(dummyContainer)
	return nil
}

func (s *SerializableResourceType) visitCore(visitor prop.Visitor) error {
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
	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().CoreIdAttribute(), s.ResourceType.ID())); err != nil {
		return err
	}

	// meta
	meta := prop.NewComplexOf(spec.MetaAttributes().CoreMetaPartialAttribute(), map[string]interface{}{
		"resourceType": s.ResourceType.ResourceTypeName(),
		"location":     s.ResourceType.ResourceLocation(),
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

func (s *SerializableResourceType) visitResourceType(visitor prop.Visitor) error {
	if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().ResourceTypeNameAttribute(), s.ResourceType.Name())); err != nil {
		return err
	}

	if description := s.ResourceType.Description(); len(description) > 0 {
		if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().ResourceTypeDescriptionAttribute(), description)); err != nil {
			return err
		}
	}

	if endpoint := s.ResourceType.Endpoint(); len(endpoint) > 0 {
		if err := visitor.Visit(prop.NewStringOf(spec.MetaAttributes().ResourceTypeEndpointAttribute(), endpoint)); err != nil {
			return err
		}
	}

	if err := visitor.Visit(prop.NewReferenceOf(spec.MetaAttributes().ResourceTypeSchemaAttribute(), s.ResourceType.Schema().ID())); err != nil {
		return err
	}

	if s.ResourceType.CountExtensions() > 0 {
		dummyMulti := prop.NewMulti(spec.MetaAttributes().ResourceTypeSchemaExtensionsAttributeNoSub())
		if err := visitor.Visit(dummyMulti); err != nil {
			return err
		}
		visitor.BeginChildren(dummyMulti)
		if err := s.ResourceType.ForEachExtension(func(extension *spec.Schema, required bool) error {
			dummyComplex := prop.NewComplex(spec.MetaAttributes().ResourceTypeSchemaExtensionsAttributeNoSub().DeriveElementAttribute())
			if err := visitor.Visit(dummyComplex); err != nil {
				return err
			}
			visitor.BeginChildren(dummyComplex)
			if err := visitor.Visit(prop.NewReferenceOf(spec.MetaAttributes().ResourceTypeSchemaExtensionSchemaAttribute(), extension.ID())); err != nil {
				return err
			}
			if err := visitor.Visit(prop.NewBooleanOf(spec.MetaAttributes().ResourceTypeSchemaExtensionRequiredAttribute(), required)); err != nil {
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
