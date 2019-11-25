package core

import (
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/pkg/core"
)

var (
	_ json.Marshaler   = (*Attribute)(nil)
	_ json.Unmarshaler = (*Attribute)(nil)
)

// The attribute model that is used through out SCIM to
// determine data type and constraint.
type Attribute struct {
	// attributes defined in SCIM
	name            string
	description     string
	typ             Type
	subAttributes   []*Attribute
	canonicalValues []string
	multiValued     bool
	required        bool
	caseExact       bool
	mutability      Mutability
	returned        Returned
	uniqueness      Uniqueness
	referenceTypes  []string

	// metadata of attribute id. an attribute id is the id of the schema to which
	// this attribute belongs to, appended by the full path of this attribute.
	// i.e. schemas, meta.version, urn:ietf:params:scim:schemas:core:2.0:Group:displayName
	id string
	// metadata for the relative index of this attribute
	// within its parent container
	index int
	// metadata for the full path of this attribute (without urn prefix)
	path string
	// metadata to indicate whether this attribute is the primary attribute.
	// primary attributes are boolean attribute within a multiValued complex
	// container. there can only be one true value among all elements.
	primary bool
	// metadata to indicate whether this attribute is the identity attribute.
	// identity attributes participates in equality comparison for its
	// complex container.
	identity bool
	// metadata of the list of annotations on this attribute.
	annotations []string
}

// Return the ID of the attribute.
// The ID can be used to universally identify an attribute.
func (attr *Attribute) ID() string {
	return attr.id
}

// Return the name of the attribute.
func (attr *Attribute) Name() string {
	return attr.name
}

// Return human-readable text to describe the attribute.
func (attr *Attribute) Description() string {
	return attr.description
}

// Return the data type of the attribute. Note that the type
// could indicate either singular or multiValued attribute.
func (attr *Attribute) Type() Type {
	return attr.typ
}

// Return true if the attribute is multiValued.
func (attr *Attribute) MultiValued() bool {
	return attr.multiValued
}

// Return true if the attribute is singular.
func (attr *Attribute) SingleValued() bool {
	return !attr.multiValued
}

// Return true if the attribute is required.
func (attr *Attribute) Required() bool {
	return attr.required
}

// Return true if the attribute is optional, which is the opposite of required.
func (attr *Attribute) Optional() bool {
	return !attr.required
}

// Return true if the attribute's value is case sensitive. Note that case sensitivity
// only applies to string type attributes.
//
// The API does not provide a version in reverse form (i.e. CaseInExact) because they
// are too similar, and may potentially confuse developers, leading to bugs. Too check
// for case insensitivity, just use !attribute.CaseExact()
func (attr *Attribute) CaseExact() bool {
	return attr.caseExact
}

// Return the mutability of the attribute.
func (attr *Attribute) Mutability() Mutability {
	return attr.mutability
}

// Return the return-ability of the attribute.
func (attr *Attribute) Returned() Returned {
	return attr.returned
}

// Return the uniqueness of the attribute.
func (attr *Attribute) Uniqueness() Uniqueness {
	return attr.uniqueness
}

// Return the total number of defined canonical values.
func (attr *Attribute) CountCanonicalValues() int {
	return len(attr.canonicalValues)
}

// Iterate through all canonical values and invoke the callback function on each value.
// This method is designed to preserve the SOLID principal. The callback SHALL NOT block
// the executing Goroutine.
func (attr *Attribute) ForEachCanonicalValue(callback func(canonicalValue string)) {
	for _, eachCanonicalValue := range attr.canonicalValues {
		callback(eachCanonicalValue)
	}
}

// Return the total number of reference types
func (attr *Attribute) CountReferenceTypes() int {
	return len(attr.referenceTypes)
}

// Iterate through all reference types and invoke the callback function on each value.
// This method is designed to preserve the SOLID principal. The callback SHALL NOT block
// the executing Goroutine.
func (attr *Attribute) ForEachReferenceType(callback func(referenceType string)) {
	for _, eachReferenceType := range attr.referenceTypes {
		callback(eachReferenceType)
	}
}

// Return the total number of sub attributes.
func (attr *Attribute) CountSubAttributes() int {
	return len(attr.subAttributes)
}

// Iterate through all sub attributes and invoke the callback function on each value.
// This method is designed to preserve the SOLID principal. The callback SHALL NOT block
// the executing Goroutine.
func (attr *Attribute) ForEachSubAttribute(callback func(subAttribute *Attribute)) {
	for _, eachSubAttribute := range attr.subAttributes {
		callback(eachSubAttribute)
	}
}

// Return the metadata about the relative index of this attribute among its parent container.
// This index is used to sort the attribute in order to maintain a stable and ideal iteration order.
func (attr *Attribute) Index() int {
	return attr.index
}

// Return the full path of this attribute.
func (attr *Attribute) Path() string {
	return attr.path
}

// Return true if this attribute is a primary attribute.
func (attr *Attribute) IsPrimary() bool {
	return attr.primary
}

// Return true if this attribute is an identity attribute.
func (attr *Attribute) IsIdentity() bool {
	return attr.identity
}

// Return the number of total annotations on this attribute.
func (attr *Attribute) CountAnnotations() int {
	return len(attr.annotations)
}

// Iterate through all annotations on this attribute and invoke callback function on each.
// This method maintains SOLID principal. The callback function SHALL NOT block.
func (attr *Attribute) ForEachAnnotation(callback func(annotation string)) {
	for _, annotation := range attr.annotations {
		callback(annotation)
	}
}

// Return an exact shallow copy of the attribute, but with the multiValued field set to false, thus
// effectively converting any multiValued attribute to singular attribute.
func (attr *Attribute) AsSingleValued() *Attribute {
	return &Attribute{
		name:            attr.name,
		description:     attr.description,
		typ:             attr.typ,
		subAttributes:   attr.subAttributes,
		canonicalValues: attr.canonicalValues,
		multiValued:     false,
		required:        attr.required,
		caseExact:       attr.caseExact,
		mutability:      attr.mutability,
		returned:        attr.returned,
		uniqueness:      attr.uniqueness,
		referenceTypes:  attr.referenceTypes,
		id:              attr.id,
		index:           attr.index,
		path:            attr.path,
		primary:         attr.primary,
		identity:        attr.identity,
		annotations:     attr.annotations,
	}
}

// Validate the attribute and panic if the required fields are not present.
func (attr *Attribute) MustValidate() {
	if len(attr.id) == 0 {
		panic("attribute id is required")
	}
	if len(attr.name) == 0 {
		panic("attribute name is required")
	}
	if len(attr.path) == 0 {
		panic("attribute path is required")
	}

	if !attr.multiValued {
		switch attr.typ {
		case core.TypeReference:
			if !attr.caseExact {
				panic("reference attribute must be case exact")
			}
		case core.TypeDateTime:
			if !attr.caseExact {
				panic("dateTime attribute must be case exact")
			}
		case core.TypeBinary:
			if !attr.caseExact {
				panic("binary attribute must be case exact")
			}
			if attr.uniqueness != UniquenessNone {
				panic("binary attribute must not have uniqueness")
			}
		}
	}
}

// Returns true if the two attributes are equal. This method checks pointer equality first.
// If does not match, goes on to check whether id and multiValued matches: these two properties
// can effectively dictate if two attributes refer to the same one.
func (attr *Attribute) Equals(other *Attribute) bool {
	return (attr == other) || (attr.id == other.id && attr.multiValued == other.multiValued)
}

// Return a string representation of the attribute. This method is intended to assist
// debugger printing, and does not intend to display all data.
func (attr *Attribute) String() string {
	if attr.multiValued {
		return fmt.Sprintf("%s (%s[])", attr.id, attr.typ.String())
	} else {
		return fmt.Sprintf("%s (%s)", attr.id, attr.typ.String())
	}
}

func (attr *Attribute) MarshalJSON() ([]byte, error) {
	tmp := new(attributeMarshaler)
	tmp.extract(attr)
	return json.Marshal(tmp)
}

func (attr *Attribute) UnmarshalJSON(raw []byte) error {
	var tmp *attributeUnmarshaler
	{
		tmp = new(attributeUnmarshaler)
		err := json.Unmarshal(raw, tmp)
		if err != nil {
			return err
		}
	}
	tmp.fill(attr)
	return nil
}

type (
	// adapter to marshal the attribute
	attributeMarshaler struct {
		Name            string       `json:"name"`
		Description     string       `json:"description,omitempty"`
		Type            Type         `json:"type"`
		SubAttributes   []*Attribute `json:"subAttributes,omitempty"`
		CanonicalValues []string     `json:"canonicalValues,omitempty"`
		MultiValued     bool         `json:"multiValued"`
		Required        bool         `json:"required"`
		CaseExact       bool         `json:"caseExact"`
		Mutability      Mutability   `json:"mutability"`
		Returned        Returned     `json:"returned"`
		Uniqueness      Uniqueness   `json:"uniqueness"`
		ReferenceTypes  []string     `json:"referenceTypes,omitempty"`
	}
	// adapter to unmarshal the attribute
	attributeUnmarshaler struct {
		ID              string                  `json:"id"`
		Name            string                  `json:"name"`
		Description     string                  `json:"description"`
		Type            string                  `json:"type"`
		SubAttributes   []*attributeUnmarshaler `json:"subAttributes"`
		CanonicalValues []string                `json:"canonicalValues"`
		MultiValued     bool                    `json:"multiValued"`
		Required        bool                    `json:"required"`
		CaseExact       bool                    `json:"caseExact"`
		Mutability      string                  `json:"mutability"`
		Returned        string                  `json:"returned"`
		Uniqueness      string                  `json:"uniqueness"`
		ReferenceTypes  []string                `json:"referenceTypes"`
		Index           int                     `json:"_index"`
		Path            string                  `json:"_path"`
		Primary         bool                    `json:"_primary"`
		Identity        bool                    `json:"_identity"`
		Annotations     []string                `json:"_annotations"`
	}
)

// Extract all values from the given attribute into this marshaler
func (m *attributeMarshaler) extract(attr *Attribute) {
	m.Name = attr.name
	m.Description = attr.description
	m.Type = attr.typ
	m.SubAttributes = attr.subAttributes
	m.CanonicalValues = attr.canonicalValues
	m.MultiValued = attr.multiValued
	m.Required = attr.required
	m.CaseExact = attr.caseExact
	m.Mutability = attr.mutability
	m.Returned = attr.returned
	m.Uniqueness = attr.uniqueness
	m.ReferenceTypes = attr.referenceTypes
}

// Fill all values from the unmarshaler into the given attribute.
func (u *attributeUnmarshaler) fill(attr *Attribute) {
	attr.id = u.ID
	attr.name = u.Name
	attr.description = u.Description
	attr.typ = MustParseType(u.Type)
	attr.canonicalValues = u.CanonicalValues
	attr.multiValued = u.MultiValued
	attr.required = u.Required
	attr.caseExact = u.CaseExact
	attr.mutability = MustParseMutability(u.Mutability)
	attr.returned = MustParseReturned(u.Returned)
	attr.uniqueness = MustParseUniqueness(u.Uniqueness)
	attr.referenceTypes = u.ReferenceTypes
	attr.index = u.Index
	attr.path = u.Path
	attr.primary = u.Primary
	attr.identity = u.Identity
	attr.annotations = u.Annotations

	if len(u.SubAttributes) > 0 {
		attr.subAttributes = make([]*Attribute, len(u.SubAttributes), len(u.SubAttributes))
		for i, sub := range u.SubAttributes {
			subAttribute := new(Attribute)
			sub.fill(subAttribute)
			attr.subAttributes[i] = subAttribute
		}
	}
}
