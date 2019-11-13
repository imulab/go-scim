package core

import (
	"fmt"
	"strings"
)

// SCIM attribute contains metadata and rules for a field in SCIM resource.
type Attribute struct {
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Type            string       `json:"type"`
	SubAttributes   []*Attribute `json:"subAttributes"`
	CanonicalValues []string     `json:"canonicalValues"`
	MultiValued     bool         `json:"multiValued"`
	Required        bool         `json:"required"`
	CaseExact       bool         `json:"caseExact"`
	Mutability      string       `json:"mutability"`
	Returned        string       `json:"returned"`
	Uniqueness      string       `json:"uniqueness"`
	ReferenceTypes  []string     `json:"referenceTypes"`
	Metadata        *Metadata    `json:"-"`
}

// Set the default value in attributes.
func (attr *Attribute) setDefaults() {
	if attr == nil {
		return
	}

	if len(attr.Type) == 0 {
		attr.Type = TypeString
	}

	if len(attr.Mutability) == 0 {
		attr.Mutability = MutabilityReadWrite
	}

	if len(attr.Returned) == 0 {
		attr.Returned = ReturnedDefault
	}

	if len(attr.Uniqueness) == 0 {
		attr.Uniqueness = UniquenessNone
	}

	if attr.Metadata == nil {
		attr.Metadata = new(Metadata)
	}

	for _, subAttr := range attr.SubAttributes {
		subAttr.setDefaults()
	}
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
	name := attr.Name
	if attr.Metadata != nil && len(attr.Metadata.Path) > 0 {
		name = attr.Metadata.Path
	}
	return name
}

// Returns a descriptive name of the attribute's type.
func (attr *Attribute) DescribeType() string {
	desc := ""
	if attr.MultiValued {
		desc = "multiValued "
	}
	desc += attr.Type
	return desc
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
		Metadata:        attr.Metadata,
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
		Metadata:        attr.Metadata,
	}
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
		metadata        *Metadata    = nil
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

		if attr.Metadata != nil {
			metadata = attr.Metadata.copy()
		}
	}

	return &Attribute{
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
		Metadata:        metadata,
	}
}

// Return true if the attribute has a sub attribute of boolean type and is marked as exclusive
func (attr *Attribute) HasExclusiveSubAttribute() bool {
	for _, subAttr := range attr.SubAttributes {
		if subAttr.Type == TypeBoolean && subAttr.Metadata != nil && subAttr.Metadata.IsExclusive {
			return true
		}
	}
	return false
}

// Returns a noTarget error about the attribute, depending on the step's type.
func (attr *Attribute) errNoTarget(step *Step) error {
	if step.IsPath() {
		return Errors.noTarget(fmt.Sprintf("%s does not have the specified sub attributes.", attr.DisplayName()))
	} else if step.IsOperator() {
		return Errors.noTarget(fmt.Sprintf("%s cannot be filtered.", attr.DisplayName()))
	} else {
		return Errors.noTarget(fmt.Sprintf("%s cannot be processed by given step", attr.DisplayName()))
	}
}

// Returns an invalidValue error about the attribute.
func (attr *Attribute) errInvalidValue() error {
	return Errors.InvalidValue(fmt.Sprintf("value is invalid or incompatible for attribute %s", attr.DisplayName()))
}
