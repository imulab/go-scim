package core

import (
	"encoding/json"
	"strings"
)

// Attribute describes the data type of Property.
type Attribute struct {
	name        string
	description string
	typ         Type
	sub         []*Attribute
	canonical   []string
	multiValued bool
	required    bool
	caseExact   bool
	mutability  Mutability
	returned    Returned
	uniqueness  Uniqueness
	refTypes    []string

	// primary is a custom setting of the Attribute, applicable to a single-valued boolean attribute who is the sub
	// attribute of a multivalued complex attribute. When set to true, only one of the properties carrying this Attribute
	// may have a true value. When a new carrier of true value emerges, the original true value carrier is set to false.
	//
	// For example, the Attribute emails.primary can have its primary set to true. Then the following structure:
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
	primary bool

	// identity is a custom setting of the Attribute, applicable to any sub attributes of a multivalued complex attribute.
	// When set true, it indicates that the property holding this Attribute wish to participate in the identity comparison.
	// These properties will contribute to the hash of their container complex property. When two complex properties
	// have the same Attribute and the same hash, they are deemed as duplicate of each other. If these duplicated
	// properties are elements of a multivalued property, they may be subject to deduplication.
	identity bool
}

// NewProperty constructs a new Property instance with this Attribute. The returned Property is always unassigned.
func (t *Attribute) NewProperty() Property {
	if t.multiValued {
		return &multiProperty{attr: t, elem: []Property{}}
	}

	if t.typ == TypeComplex {
		p := &complexProperty{attr: t, nameIndex: map[string]int{}}
		for i, sub := range p.attr.sub {
			p.children = append(p.children, sub.NewProperty())
			p.nameIndex[strings.ToLower(sub.name)] = i
		}
		return p
	}

	return &simpleProperty{attr: t}
}

func (t *Attribute) asSingleValued() *Attribute {
	if !t.multiValued {
		return t
	}

	return &Attribute{
		name:        t.name,
		description: t.description,
		typ:         t.typ,
		sub:         t.sub,
		canonical:   t.canonical,
		multiValued: false,
		required:    t.required,
		caseExact:   t.caseExact,
		mutability:  t.mutability,
		returned:    t.returned,
		uniqueness:  t.uniqueness,
		refTypes:    t.refTypes,
		primary:     t.primary,
		identity:    t.identity,
	}
}

func (t *Attribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(attributeJSON{
		Name:        t.name,
		Description: t.description,
		Type:        t.typ,
		Sub:         t.sub,
		Canonical:   t.canonical,
		MultiValued: t.multiValued,
		Required:    t.required,
		CaseExact:   t.caseExact,
		Mutability:  t.mutability,
		Returned:    t.returned,
		Uniqueness:  t.uniqueness,
		RefTypes:    t.refTypes,
		Primary:     t.primary,
		Identity:    t.identity,
	})
}

type attributeJSON struct {
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
	Primary     bool         `json:"primary,omitempty"`
	Identity    bool         `json:"identity,omitempty"`
}

type AttributeDsl struct {
	Name        string
	Description string
	Type        Type
	Sub         []*Attribute
	Canonical   []string
	MultiValued bool
	Required    bool
	CaseExact   bool
	Mutability  Mutability
	Returned    Returned
	Uniqueness  Uniqueness
	RefTypes    []string
	Primary     bool
	Identity    bool
}

func (d *AttributeDsl) Describe(text string) *AttributeDsl {
	d.Description = text
	return d
}
