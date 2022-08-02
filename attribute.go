package scim

import "encoding/json"

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

func (t *Attribute) singleValued() *Attribute {
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

type attributeDsl Attribute

// StringAttribute starts an Attribute builder DSL for string typed attributes.
func StringAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeString}
}

// IntegerAttribute starts an Attribute builder DSL for integer typed attributes.
func IntegerAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeInteger}
}

// DecimalAttribute starts an Attribute builder DSL for string typed attributes.
func DecimalAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeDecimal}
}

// BooleanAttribute starts an Attribute builder DSL for boolean typed attributes.
func BooleanAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeBoolean}
}

// DateTimeAttribute starts an Attribute builder DSL for dateTime typed attributes.
func DateTimeAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeDateTime}
}

// BinaryAttribute starts an Attribute builder DSL for binary typed attributes.
func BinaryAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeBinary}
}

// ReferenceAttribute starts an Attribute builder DSL for reference typed attributes.
func ReferenceAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeReference}
}

// ComplexAttribute starts an Attribute builder DSL for complex typed attributes.
func ComplexAttribute(name string) *attributeDsl {
	return &attributeDsl{name: name, typ: TypeComplex}
}

func (d *attributeDsl) Describe(text string) *attributeDsl {
	d.description = text
	return d
}

func (d *attributeDsl) CanonicalValues(values ...string) *attributeDsl {
	d.canonical = append(d.canonical, values...)
	return d
}

func (d *attributeDsl) ReferenceTypes(values ...string) *attributeDsl {
	d.refTypes = append(d.refTypes, values...)
	return d
}

func (d *attributeDsl) MultiValued() *attributeDsl {
	d.multiValued = true
	return d
}

func (d *attributeDsl) Required() *attributeDsl {
	d.required = true
	return d
}

func (d *attributeDsl) CaseExact() *attributeDsl {
	d.caseExact = true
	return d
}

func (d *attributeDsl) ReadOnly() *attributeDsl {
	d.mutability = MutabilityReadOnly
	return d
}

func (d *attributeDsl) WriteOnly() *attributeDsl {
	d.mutability = MutabilityWriteOnly
	return d
}

func (d *attributeDsl) Immutable() *attributeDsl {
	d.mutability = MutabilityImmutable
	return d
}

func (d *attributeDsl) AlwaysReturn() *attributeDsl {
	d.returned = ReturnedAlways
	return d
}

func (d *attributeDsl) NeverReturn() *attributeDsl {
	d.returned = ReturnedNever
	return d
}

func (d *attributeDsl) ReturnOnRequest() *attributeDsl {
	d.returned = ReturnedRequest
	return d
}

func (d *attributeDsl) UniqueOnServer() *attributeDsl {
	d.uniqueness = UniquenessServer
	return d
}

func (d *attributeDsl) UniqueGlobally() *attributeDsl {
	d.uniqueness = UniquenessGlobal
	return d
}

func (d *attributeDsl) Primary() *attributeDsl {
	if d.typ != TypeBoolean {
		panic("only boolean attribute can be marked as primary")
	}
	d.primary = true
	return d
}

func (d *attributeDsl) Identity() *attributeDsl {
	d.identity = true
	return d
}

func (d *attributeDsl) build() *Attribute {
	return (*Attribute)(d)
}
