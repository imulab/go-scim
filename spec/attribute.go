package spec

import (
	"encoding/json"
	"github.com/elvsn/scim.go/spec/internal"
	"sort"
)

// Enhanced SCIM attribute model.
// Attribute is the basic unit that describes data requirement in SCIM. It includes the data requirement defined
// in RFC7643. It also includes additional metadata that makes actual SCIM processing easier.
type Attribute struct {
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
	id              string                            // unique id of the attribute
	index           int                               // relative index in ascending order
	path            string                            // SCIM path name from the root attribute
	annotations     map[string]map[string]interface{} // annotations that provide additional processing hint
}

func (attr *Attribute) MarshalJSON() ([]byte, error) {
	m := internal.AttributeMarshaler{
		SubAttributes: []*internal.AttributeMarshaler{},
	}
	attr.convertToMarshaler(&m)
	return json.Marshal(m)
}

func (attr *Attribute) convertToMarshaler(m *internal.AttributeMarshaler) {
	m.Name = attr.name
	m.Description = attr.description
	m.Type = attr.typ.String()
	m.CanonicalValues = attr.canonicalValues
	m.MultiValued = attr.multiValued
	m.Required = attr.required
	m.CaseExact = attr.caseExact
	m.Mutability = attr.mutability.String()
	m.Returned = attr.returned.String()
	m.Uniqueness = attr.uniqueness.String()
	m.ReferenceTypes = attr.referenceTypes

	for _, subAttr := range attr.subAttributes {
		sm := internal.AttributeMarshaler{
			SubAttributes: []*internal.AttributeMarshaler{},
		}
		subAttr.convertToMarshaler(&sm)
		m.SubAttributes = append(m.SubAttributes, &sm)
	}
}

func (attr *Attribute) UnmarshalJSON(raw []byte) error {
	var um internal.AttributeUnmarshaler
	if err := json.Unmarshal(raw, &um); err != nil {
		return err
	}

	attr.convertFromUnmarshaler(&um)
	attr.sort() // sort the sub attributes recursively to maintain strong order

	return nil
}

func (attr *Attribute) convertFromUnmarshaler(um *internal.AttributeUnmarshaler) {
	attr.id = um.ID
	attr.name = um.Name
	attr.description = um.Description
	attr.typ = mustParseType(um.Type)
	attr.canonicalValues = um.CanonicalValues
	attr.multiValued = um.MultiValued
	attr.required = um.Required
	attr.caseExact = um.CaseExact
	attr.mutability = mustParseMutability(um.Mutability)
	attr.returned = mustParseReturned(um.Returned)
	attr.uniqueness = mustParseUniqueness(um.Uniqueness)
	attr.referenceTypes = um.ReferenceTypes
	attr.index = um.Index
	attr.path = um.Path
	attr.annotations = um.Annotations
	attr.subAttributes = []*Attribute{}

	for _, subum := range um.SubAttributes {
		subAttr := new(Attribute)
		subAttr.convertFromUnmarshaler(subum)
		attr.subAttributes = append(attr.subAttributes, subAttr)
	}
}

func (attr *Attribute) sort() {
	for _, subAttr := range attr.subAttributes {
		subAttr.sort()
	}
	sort.Sort(attr)
}

func (attr *Attribute) Len() int {
	return len(attr.subAttributes)
}

func (attr *Attribute) Less(i, j int) bool {
	return attr.subAttributes[i].index < attr.subAttributes[j].index
}

func (attr *Attribute) Swap(i, j int) {
	attr.subAttributes[i], attr.subAttributes[j] = attr.subAttributes[j], attr.subAttributes[i]
}
