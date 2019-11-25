package core

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
		ReferenceTypes  []string                `json:"referenceTypes,omitempty"`
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

	if len(u.SubAttributes) > 0 {
		attr.subAttributes = make([]*Attribute, len(u.SubAttributes), len(u.SubAttributes))
		for i, sub := range u.SubAttributes {
			subAttribute := new(Attribute)
			sub.fill(subAttribute)
			attr.subAttributes[i] = subAttribute
		}
	}
}
