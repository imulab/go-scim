package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

type (
	// SCIM attribute contains metadata and rules for a field in SCIM resource.
	Attribute struct {
		Id              string       `json:"id"`
		Name            string       `json:"name"`
		Description     string       `json:"description"`
		Type            Type         `json:"type"`
		SubAttributes   []*Attribute `json:"subAttributes"`
		CanonicalValues []string     `json:"canonicalValues"`
		MultiValued     bool         `json:"multiValued"`
		Required        bool         `json:"required"`
		CaseExact       bool         `json:"caseExact"`
		Mutability      Mutability   `json:"mutability"`
		Returned        Returned     `json:"returned"`
		Uniqueness      Uniqueness   `json:"uniqueness"`
		ReferenceTypes  []string     `json:"referenceTypes"`
	}

	// An internal JSON representation of Attribute
	tmpAttr struct {
		Id              string     `json:"id"`
		Name            string     `json:"name"`
		Description     string     `json:"description"`
		Type            string     `json:"type"`
		SubAttributes   []*tmpAttr `json:"subAttributes"`
		CanonicalValues []string   `json:"canonicalValues"`
		MultiValued     bool       `json:"multiValued"`
		Required        bool       `json:"required"`
		CaseExact       bool       `json:"caseExact"`
		Mutability      string     `json:"mutability"`
		Returned        string     `json:"returned"`
		Uniqueness      string     `json:"uniqueness"`
		ReferenceTypes  []string   `json:"referenceTypes"`
	}
)

// Convert the temporary attr representation to the official attribute
func (attr *tmpAttr) convert() *Attribute {
	converted := &Attribute{
		Id:              attr.Id,
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            NewType(attr.Type),
		CanonicalValues: attr.CanonicalValues,
		MultiValued:     attr.MultiValued,
		Required:        attr.Required,
		CaseExact:       attr.CaseExact,
		Mutability:      NewMutability(attr.Mutability),
		Returned:        NewReturned(attr.Returned),
		Uniqueness:      NewUniqueness(attr.Uniqueness),
		ReferenceTypes:  attr.ReferenceTypes,
	}

	if len(attr.SubAttributes) > 0 {
		converted.SubAttributes = make([]*Attribute, 0)
		for _, subAttr := range attr.SubAttributes {
			converted.SubAttributes = append(converted.SubAttributes, subAttr.convert())
		}
	}

	return converted
}

// Implementation of json.Unmarshaler
func (attr *Attribute) UnmarshalJSON(raw []byte) error {
	tmp := new(tmpAttr)
	err := json.Unmarshal(raw, tmp)
	if err != nil {
		return err
	}

	converted := tmp.convert()

	attr.Id = converted.Id
	attr.Name = converted.Name
	attr.Description = converted.Description
	attr.Type = converted.Type
	attr.CanonicalValues = converted.CanonicalValues
	attr.MultiValued = converted.MultiValued
	attr.Required = converted.Required
	attr.CaseExact = converted.CaseExact
	attr.Mutability = converted.Mutability
	attr.Returned = converted.Returned
	attr.Uniqueness = converted.Uniqueness
	attr.ReferenceTypes = converted.ReferenceTypes
	attr.SubAttributes = converted.SubAttributes

	return nil
}

// Implementation of json.Marshaler
func (attr *Attribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name            string       `json:"name"`
		Description     string       `json:"description,omitempty"`
		Type            string       `json:"type"`
		SubAttributes   []*Attribute `json:"subAttributes,omitempty"`
		CanonicalValues []string     `json:"canonicalValues,omitempty"`
		MultiValued     bool         `json:"multiValued"`
		Required        bool         `json:"required"`
		CaseExact       bool         `json:"caseExact"`
		Mutability      string       `json:"mutability"`
		Returned        string       `json:"returned"`
		Uniqueness      string       `json:"uniqueness"`
		ReferenceTypes  []string     `json:"referenceTypes,omitempty"`
	}{
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            attr.Type.String(),
		SubAttributes:   attr.SubAttributes,
		CanonicalValues: attr.CanonicalValues,
		MultiValued:     attr.MultiValued,
		Required:        attr.Required,
		CaseExact:       attr.CaseExact,
		Mutability:      attr.Mutability.String(),
		Returned:        attr.Returned.String(),
		Uniqueness:      attr.Uniqueness.String(),
		ReferenceTypes:  attr.ReferenceTypes,
	})
}

// Returns true if the property that this attribute represents can be addressed
// by the name. According to SCIM spec, comparison is made against the name of the
// attribute in case insensitive fashion.
func (attr *Attribute) GoesBy(name string) bool {
	return attr != nil && strings.ToLower(name) == strings.ToLower(attr.Name)
}

// Returns a proper name for this attribute suitable for display purposes. Defaults
// to the attribute's name, and will use the metadata's path value, if available.
func (attr *Attribute) DisplayName() string {
	return attr.MustDefaultMetadata().Path
}

// Returns a descriptive name of the attribute's type.
func (attr *Attribute) DescribeType() string {
	if attr.MultiValued {
		return "multiValued " + attr.Type.String()
	}
	return attr.Type.String()
}

// Returns true if this attribute's type is string based, namely string, reference, binary and dateTime
func (attr *Attribute) IsStringBasedType() bool {
	return attr.Type == TypeString ||
		attr.Type == TypeReference ||
		attr.Type == TypeBinary ||
		attr.Type == TypeDateTime
}

// Returns nil if this attribute is compatible with the operation, otherwise a descriptive error
func (attr *Attribute) CheckOpCompatibility(op string) error {
	switch op {
	case And, Or, Not:
		return nil
	case Eq:
		if attr.Type == TypeComplex {
			return fmt.Errorf("%s cannot be applied to complex properties", op)
		}
		return nil
	case Ne:
		if attr.MultiValued || attr.Type == TypeComplex {
			return fmt.Errorf("%s cannot be applied to multiValued or complex properties", op)
		}
		return nil
	case Sw, Ew, Co:
		if attr.MultiValued {
			return fmt.Errorf("%s cannot be applied to multiValued properties", op)
		}
		switch attr.Type {
		case TypeString, TypeReference:
			return nil
		default:
			return fmt.Errorf("%s can only be applied to string or reference properties", op)
		}
	case Gt, Ge, Lt, Le:
		if attr.MultiValued {
			return fmt.Errorf("%s cannot be applied to multiValued properties", op)
		}
		switch attr.Type {
		case TypeInteger, TypeDecimal, TypeDateTime, TypeString:
			return nil
		default:
			return fmt.Errorf("%s can only be applied to integer, decimal, dateTime and string properties", op)
		}
	case Pr:
		return nil
	default:
		panic("invalid operator")
	}
}

// Return an attribute same to this attribute, but with multiValued set to false.
// If this attribute is not multiValued, this is returned.
func (attr *Attribute) ToSingleValued() *Attribute {
	if attr == nil {
		return nil
	}

	if !attr.MultiValued {
		return attr
	}

	return &Attribute{
		Id:              attr.Id,
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            attr.Type,
		SubAttributes:   attr.SubAttributes,
		CanonicalValues: attr.CanonicalValues,
		MultiValued:     false,
		Required:        attr.Required,
		CaseExact:       attr.CaseExact,
		Mutability:      attr.Mutability,
		Returned:        attr.Returned,
		Uniqueness:      attr.Uniqueness,
		ReferenceTypes:  attr.ReferenceTypes,
	}
}

// Return an attribute same to this attribute, but with required set to false.
// If this attribute is not required, this is returned.
func (attr *Attribute) ToOptional() *Attribute {
	if attr == nil {
		return nil
	}

	if !attr.Required {
		return attr
	}

	return &Attribute{
		Id:              attr.Id,
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            attr.Type,
		SubAttributes:   attr.SubAttributes,
		CanonicalValues: attr.CanonicalValues,
		MultiValued:     attr.MultiValued,
		Required:        false,
		CaseExact:       attr.CaseExact,
		Mutability:      attr.Mutability,
		Returned:        attr.Returned,
		Uniqueness:      attr.Uniqueness,
		ReferenceTypes:  attr.ReferenceTypes,
	}
}

// Return true if two attribute are the same.
func (attr *Attribute) Equals(another *Attribute) bool {
	if attr.Id != another.Id {
		// Definitely not the same if ids differ.
		return false
	}
	// Check if one of them is derived from multiValued
	return attr.MultiValued == another.MultiValued
}

// Make a deep copy of the attribute.
func (attr *Attribute) Copy() *Attribute {
	if attr == nil {
		return nil
	}

	var (
		subAttributes   []*Attribute = nil
		canonicalValues []string     = nil
		referenceTypes  []string     = nil
	)
	{
		if len(attr.SubAttributes) > 0 {
			subAttributes = make([]*Attribute, 0)
			for _, subAttr := range attr.SubAttributes {
				subAttributes = append(subAttributes, subAttr.Copy())
			}
		}

		if len(attr.CanonicalValues) > 0 {
			canonicalValues = make([]string, 0)
			for _, cv := range attr.CanonicalValues {
				canonicalValues = append(canonicalValues, cv)
			}
		}

		if len(attr.ReferenceTypes) > 0 {
			referenceTypes = make([]string, 0)
			for _, ref := range attr.ReferenceTypes {
				referenceTypes = append(referenceTypes, ref)
			}
		}
	}

	return &Attribute{
		Id:              attr.Id,
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            attr.Type,
		SubAttributes:   subAttributes,
		CanonicalValues: canonicalValues,
		MultiValued:     attr.MultiValued,
		Required:        attr.Required,
		CaseExact:       attr.CaseExact,
		Mutability:      attr.Mutability,
		Returned:        attr.Returned,
		Uniqueness:      attr.Uniqueness,
		ReferenceTypes:  referenceTypes,
	}
}

// Get the metadata corresponding to this attribute. The method assumes such metadata exists, otherwise would panic.
func (attr *Attribute) MustDefaultMetadata() *DefaultMetadata {
	return Meta.Get(attr.Id, DefaultMetadataId).(*DefaultMetadata)
}

// Return true if the attribute has a sub attribute of boolean type and is marked as exclusive
func (attr *Attribute) HasExclusiveSubAttribute() bool {
	for _, subAttr := range attr.SubAttributes {
		if subAttr.Type == TypeBoolean && subAttr.MustDefaultMetadata().IsExclusive {
			return true
		}
	}
	return false
}

// Returns a noTarget error about the attribute, depending on the step's type.
func (attr *Attribute) errNoTarget(step *Step) error {
	if step.IsPath() {
		return Errors.noTarget(fmt.Sprintf("'%s' does not have the specified sub attributes.", attr.DisplayName()))
	} else if step.IsOperator() {
		return Errors.noTarget(fmt.Sprintf("'%s' cannot be filtered.", attr.DisplayName()))
	} else {
		return Errors.noTarget(fmt.Sprintf("'%s' cannot be processed by given step", attr.DisplayName()))
	}
}

// Returns an invalidValue error about the attribute.
func (attr *Attribute) errInvalidValue() error {
	return Errors.InvalidValue(fmt.Sprintf("value is invalid or incompatible for attribute '%s'", attr.DisplayName()))
}
