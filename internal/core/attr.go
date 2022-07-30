package core

import "strings"

// Attribute describes the data type of Property.
type Attribute struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Type        Type         `json:"type"`
	Sub         []*Attribute `json:"subAttributes,omitempty"`
	Canonical   []string     `json:"canonicalValues,omitempty"`
	MultiValued bool         `json:"multiValued"`
	Required    bool         `json:"required"`
	CaseExact   bool         `json:"caseExact"`
	Mutability  Mutability   `json:"mutability"`
	Returned    Returned     `json:"returned"`
	Uniqueness  Uniqueness   `json:"uniqueness"`
	RefTypes    []string     `json:"referenceTypes,omitempty"`

	// Primary is a custom setting of the Attribute, applicable to a single-valued boolean attribute who is the sub
	// attribute of a multivalued complex attribute. When set to true, only one of the properties carrying this Attribute
	// may have a true value. When a new carrier of true value emerges, the original true value carrier is set to false.
	//
	// For example, the Attribute emails.primary can have its Primary set to true. Then the following structure:
	//
	//	{"emails": [{"value": "foo@example.com", "primary": true}]}
	//
	// added with a value of:
	//
	//	{"value": "bar@example.com", "primary": true}
	//
	// becomes:
	//
	//	{"emails": [{"value": "foo@example.com", "primary": false}, {"value": "bar@example.com", "primary": true}]}
	//
	// This setting is only applicable in the above-mentioned context. When used out of this context, this setting has
	// no effect.
	Primary bool `json:"primary,omitempty"`

	// Identity is a custom setting of the Attribute, applicable to any sub attributes of a multivalued complex attribute.
	// When set true, it indicates that the property holding this Attribute wish to participate in the identity comparison.
	// These properties will contribute to the hash of their container complex property. When two complex properties
	// have the same Attribute and the same hash, they are deemed as duplicate of each other. If these duplicated
	// properties are elements of a multivalued property, they may be subject to deduplication.
	Identity bool `json:"identity,omitempty"`
}

// NewProperty constructs a new Property instance with this Attribute. The returned Property is always unassigned.
func (t *Attribute) NewProperty() Property {
	if t.MultiValued {
		return &multiProperty{attr: t, elem: []Property{}}
	}

	if t.Type == TypeComplex {
		p := &complexProperty{attr: t, nameIndex: map[string]int{}}
		for i, sub := range p.attr.Sub {
			p.children = append(p.children, sub.NewProperty())
			p.nameIndex[strings.ToLower(sub.Name)] = i
		}
		return p
	}

	return &simpleProperty{attr: t}
}

func (t *Attribute) asSingleValued() *Attribute {
	if !t.MultiValued {
		return t
	}

	return &Attribute{
		Name:        t.Name,
		Description: t.Description,
		Type:        t.Type,
		Sub:         t.Sub,
		Canonical:   t.Canonical,
		MultiValued: false,
		Required:    t.Required,
		CaseExact:   t.CaseExact,
		Mutability:  t.Mutability,
		Returned:    t.Returned,
		Uniqueness:  t.Uniqueness,
		RefTypes:    t.RefTypes,
		Primary:     t.Primary,
		Identity:    t.Identity,
	}
}
