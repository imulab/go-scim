package scim

import (
	"encoding/json"
	"fmt"
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

	// id is an internally maintained identifier of the attribute. Its value is in the format of <schema urn>:<path>.
	id string

	// path is an internally maintained full name of the attribute. The path starts from the attribute name at the root
	// of the resource and ends with the current attribute name, delimited by period. For example, nickName, emails.value.
	path string
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
		id:          fmt.Sprintf("%s$elem", t.id),
		path:        t.path,
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
	MultiValued bool         `json:"multiValued,omitempty"`
	Required    bool         `json:"required,omitempty"`
	CaseExact   bool         `json:"caseExact,omitempty"`
	Mutability  Mutability   `json:"mutability,omitempty"`
	Returned    Returned     `json:"returned,omitempty"`
	Uniqueness  Uniqueness   `json:"uniqueness,omitempty"`
	RefTypes    []string     `json:"referenceTypes,omitempty"`
	Primary     bool         `json:"-"`
	Identity    bool         `json:"-"`
}

// createProperty constructs a new Property instance with this Attribute. The returned Property is always unassigned.
func (t *Attribute) createProperty() Property {
	if t.multiValued {
		return &multiProperty{attr: t, elem: []Property{}}
	}

	if t.typ == TypeComplex {
		p := &complexProperty{attr: t, nameIndex: map[string]int{}}
		for i, sub := range p.attr.sub {
			p.children = append(p.children, sub.createProperty())
			p.nameIndex[strings.ToLower(sub.name)] = i
		}
		return p
	}

	return &simpleProperty{attr: t}
}

// toSingleValued returns an Attribute instance with multiValued set to false. It the multiValued is already false, the
// same instance is returned.
func (t *Attribute) toSingleValued() *Attribute {
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

type attributeDsl struct {
	Attribute
	namespace string
	prefix    string
}

func (d *attributeDsl) Name(name string) *attributeDsl {
	d.name = name
	return d
}

func (d *attributeDsl) Describe(text string) *attributeDsl {
	d.description = text
	return d
}

func (d *attributeDsl) String() *attributeDsl {
	d.typ = TypeString
	return d
}

func (d *attributeDsl) Integer() *attributeDsl {
	d.typ = TypeInteger
	return d
}

func (d *attributeDsl) Decimal() *attributeDsl {
	d.typ = TypeDecimal
	return d
}

func (d *attributeDsl) Boolean() *attributeDsl {
	d.typ = TypeBoolean
	return d
}

func (d *attributeDsl) DateTime() *attributeDsl {
	d.typ = TypeDateTime
	return d
}

func (d *attributeDsl) Binary() *attributeDsl {
	d.typ = TypeBinary
	return d
}

func (d *attributeDsl) Reference() *attributeDsl {
	d.typ = TypeReference
	return d
}

func (d *attributeDsl) Complex() *attributeDsl {
	d.typ = TypeComplex
	return d
}

func (d *attributeDsl) CanonicalValues(values ...string) *attributeDsl {
	d.canonical = append(d.canonical, values...)
	return d
}

func (d *attributeDsl) ReferenceTypes(values ...string) *attributeDsl {
	if d.typ != TypeReference {
		panic("referenceTypes can only be applied to reference typed attributes")
	}
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

func (d *attributeDsl) SubAttributes(fn func(sd *attributeListDsl)) *attributeDsl {
	if d.typ != TypeComplex {
		panic("sub attributes can only be added complex typed attributes")
	}

	if len(d.name) == 0 {
		panic("set name first")
	}

	sd := &attributeListDsl{
		namespace: d.namespace,
		prefix: func() string {
			if len(d.prefix) == 0 {
				return d.name
			} else {
				return fmt.Sprintf("%s.%s", d.prefix, d.name)
			}
		}(),
	}
	fn(sd)

	for _, it := range sd.list {
		d.sub = append(d.sub, it.build())
	}

	return d
}

func (d *attributeDsl) build() *Attribute {
	attr := &d.Attribute

	if len(d.prefix) == 0 {
		attr.path = attr.name
	} else {
		attr.path = fmt.Sprintf("%s.%s", d.prefix, attr.name)
	}

	attr.id = fmt.Sprintf("%s:%s", d.namespace, attr.path)

	return attr
}

type attributeListDsl struct {
	list      []*attributeDsl
	namespace string
	prefix    string
}

func (d *attributeListDsl) Add(fn func(d *attributeDsl)) *attributeListDsl {
	d0 := &attributeDsl{namespace: d.namespace, prefix: d.prefix}
	fn(d0)
	d.list = append(d.list, d0)
	return d
}

func (d *attributeListDsl) build() []*Attribute {
	var attrs []*Attribute
	for _, it := range d.list {
		attrs = append(attrs, it.build())
	}
	return attrs
}

var (
	coreAttributes = new(attributeListDsl).Add(func(d *attributeDsl) {
		d.Name("schemas").Reference().MultiValued().Required().CaseExact().AlwaysReturn()
	}).Add(func(d *attributeDsl) {
		d.Name("id").String().CaseExact().AlwaysReturn().ReadOnly().UniqueGlobally()
	}).Add(func(d *attributeDsl) {
		d.Name("externalId").String()
	}).Add(func(d *attributeDsl) {
		d.Name("meta").Complex().ReadOnly().SubAttributes(func(sd *attributeListDsl) {
			sd.Add(func(d *attributeDsl) {
				d.Name("resourceType").String().CaseExact().ReadOnly()
			}).Add(func(d *attributeDsl) {
				d.Name("created").DateTime().ReadOnly()
			}).Add(func(d *attributeDsl) {
				d.Name("lastModified").DateTime().ReadOnly()
			}).Add(func(d *attributeDsl) {
				d.Name("location").Reference().ReadOnly()
			}).Add(func(d *attributeDsl) {
				d.Name("version").String().ReadOnly()
			})
		})
	}).build()
)
