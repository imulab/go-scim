package scim

import "encoding/json"

// Attribute describes data requirement of a SCIM property.
type Attribute struct {
	name            string
	description     string
	typ             Type
	subAttrs        []*Attribute
	canonicalValues []string
	multiValued     bool
	required        bool
	caseExact       bool
	mutability      Mutability
	returned        Returned
	uniqueness      Uniqueness
	refTypes        []string

	primary  bool
	identity bool
}

func (a *Attribute) toSingleValued() *Attribute {
	return &Attribute{
		name:            a.name,
		description:     a.description,
		typ:             a.typ,
		subAttrs:        a.subAttrs,
		canonicalValues: a.canonicalValues,
		multiValued:     false,
		required:        a.required,
		caseExact:       a.caseExact,
		mutability:      a.mutability,
		returned:        a.returned,
		uniqueness:      a.uniqueness,
		refTypes:        a.refTypes,
		primary:         a.primary,
		identity:        a.identity,
	}
}

func (a *Attribute) MarshalJSON() ([]byte, error) {
	type attributeJSON struct {
		Name            string       `json:"name"`
		Description     string       `json:"description,omitempty"`
		Type            Type         `json:"type"`
		SubAttrs        []*Attribute `json:"subAttributes,omitempty"`
		CanonicalValues []string     `json:"canonicalValues,omitempty"`
		MultiValued     bool         `json:"multiValued,omitempty"`
		Required        bool         `json:"required,omitempty"`
		CaseExact       bool         `json:"caseExact,omitempty"`
		Mutability      Mutability   `json:"mutability,omitempty"`
		Returned        Returned     `json:"returned,omitempty"`
		Uniqueness      Uniqueness   `json:"uniqueness,omitempty"`
		RefTypes        []string     `json:"referenceTypes,omitempty"`
	}

	return json.Marshal(attributeJSON{
		Name:            a.name,
		Description:     a.description,
		Type:            a.typ,
		SubAttrs:        a.subAttrs,
		CanonicalValues: a.canonicalValues,
		MultiValued:     a.multiValued,
		Required:        a.required,
		CaseExact:       a.caseExact,
		Mutability:      a.mutability,
		Returned:        a.returned,
		Uniqueness:      a.uniqueness,
		RefTypes:        a.refTypes,
	})
}

func StringAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeString)
}

func IntegerAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeInteger)
}

func DecimalAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeDecimal)
}

func DateTimeAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeDateTime)
}

func BooleanAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeBoolean)
}

func BinaryAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeBinary)
}

func ReferenceAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeReference)
}

func ComplexAttribute(name string) *attributeDsl {
	return newAttributeDsl(name, TypeComplex)
}

func newAttributeDsl(name string, typ Type) *attributeDsl {
	if len(name) == 0 {
		panic("name is required")
	}

	return &attributeDsl{
		attr: &Attribute{name: name, typ: typ},
	}
}

type attributeDsl struct {
	attr *Attribute
}

func (d *attributeDsl) Description(description string) *attributeDsl {
	d.attr.description = description
	return d
}

func (d *attributeDsl) MultiValued() *attributeDsl {
	d.attr.multiValued = true
	return d
}

func (d *attributeDsl) Required() *attributeDsl {
	d.attr.required = true
	return d
}

func (d *attributeDsl) CanonicalValues(values ...string) *attributeDsl {
	if d.attr.typ != TypeString {
		panic("canonicalValues only applies to string attribute")
	}
	d.attr.canonicalValues = append(d.attr.canonicalValues, values...)
	return d
}

func (d *attributeDsl) CaseExact() *attributeDsl {
	if d.attr.typ != TypeString && d.attr.typ != TypeReference {
		panic("caseExact only applies to string or reference attribute")
	}
	d.attr.caseExact = true
	return d
}

func (d *attributeDsl) ReadOnly() *attributeDsl {
	d.attr.mutability = MutabilityReadOnly
	return d
}

func (d *attributeDsl) WriteOnly() *attributeDsl {
	d.attr.mutability = MutabilityWriteOnly
	return d
}

func (d *attributeDsl) Immutable() *attributeDsl {
	d.attr.mutability = MutabilityImmutable
	return d
}

func (d *attributeDsl) AlwaysReturn() *attributeDsl {
	d.attr.returned = ReturnedAlways
	return d
}

func (d *attributeDsl) NeverReturn() *attributeDsl {
	d.attr.returned = ReturnedNever
	return d
}

func (d *attributeDsl) ReturnOnRequest() *attributeDsl {
	d.attr.returned = ReturnedRequest
	return d
}

func (d *attributeDsl) UniqueOnServer() *attributeDsl {
	d.attr.uniqueness = UniquenessServer
	return d
}

func (d *attributeDsl) UniqueGlobally() *attributeDsl {
	d.attr.uniqueness = UniquenessGlobal
	return d
}

func (d *attributeDsl) ReferenceTypes(types ...string) *attributeDsl {
	if d.attr.typ != TypeReference {
		panic("referenceType only applies to reference attributes")
	}
	d.attr.refTypes = append(d.attr.refTypes, types...)
	return d
}

// MarkAsPrimary marks a singular boolean attribute as the "primary" attribute. It is useful when such attribute is
// defined as a sub attribute of a multiValued complex attribute. Within a multiValued complex property, only one element
// may have a sub property that is both primary and true. This means that, when another element's primary sub property
// is assigned to true, the original true primary is automatically assigned false.
//
// As an example, when "emails.primary" is marked as primary:
//
//	// The original structure:
//	{
//	  "emails": [
//	    {"value": "foo@bar.com", "type": "work", "primary": true}
//	  ]
//	}
//
//	// When added with:
//	{"value": "baz@bar.com", "type": "home", "primary": true}
//
//	// Becomes:
//	{
//	  "emails": [
//	    {"value": "foo@bar.com", "type": "work", "primary": false},
//	    {"value": "baz@bar.com", "type": "home", "primary": true}
//	  ]
//	}
//
// This method may only be used on singular boolean attributes. Calling on any other type panics.
func (d *attributeDsl) MarkAsPrimary() *attributeDsl {
	if d.attr.typ != TypeBoolean || d.attr.multiValued {
		panic("only singular boolean attributes can be marked as primary")
	}
	d.attr.primary = true
	return d
}

func (d *attributeDsl) MarkAsIdentity() *attributeDsl {
	if d.attr.typ == TypeComplex || d.attr.multiValued {
		panic("only singular non-complex attributes can be marked as identity")
	}
	d.attr.identity = true
	return d
}

func (d *attributeDsl) WithSubAttributes(sub ...*Attribute) *attributeDsl {
	if d.attr.typ != TypeComplex {
		panic("only complex attributes can have sub attributes")
	}
	d.attr.subAttrs = append(d.attr.subAttrs, sub...)
	return d
}

func (d *attributeDsl) Build() *Attribute {
	return d.attr
}
