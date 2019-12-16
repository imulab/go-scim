package spec

import (
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/pkg/core/annotations"
	"sort"
	"strings"
)

const (
	elemSuffix = "$elem"
)

// Enhanced SCIM attribute model. Attribute is the basic unit that describes data requirement in SCIM. It
// includes the data requirement defined in RFC7643. It also includes additional metadata that makes actual
// SCIM processing easier.
type Attribute struct {
	// ==========================
	// Attributes defined in SCIM
	// ==========================
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

	// =====================
	// Additional attributes
	// =====================
	id              string
	index           int
	path            string
	annotationIndex map[string]struct{}
}

// Validate the attribute. This method panics if encounters any error.
func (attr *Attribute) MustValidate() {
	v := &attributeValidator{attr:attr}
	for _, rule := range []func() error {
		v.checkRequired,
		v.checkCaseExact,
		v.checkPrimary,
		v.checkUniqueness,
	} {
		if err := rule(); err != nil {
			panic(err)
		}
	}
}

// Return ID of the attribute that uniquely identifies an attribute globally. ID shall be in
// the format of <schema_urn>:<full_path>. Core attributes has no need to prefix the schema URN.
// For instance, "schemas", "meta.version", "urn:ietf:params:scim:schemas:core:2.0:Group:displayName"
func (attr *Attribute) ID() string {
	return attr.id
}

// Return relative index of this attribute within its parent attribute, or on the top level. This
// index is used to sort the attribute, or the property carrying this attribute in ascending
// order within the scope of its container. The intention is to provide a stable iteration order,
// despite Golang's map does not range in a predictable order. The actual value of the index does
// not matter as long as they can be sorted in ascending order correctly.
func (attr *Attribute) Index() int {
	return attr.index
}

// Return the full path of this attribute. The full path will be the name of the attribute if this
// attribute is on the top level. If the attribute is a sub attribute, it will be attribute names
// delimited by period (".").
func (attr *Attribute) Path() string {
	return attr.path
}

// Return true if this attribute is considered a primary attribute. A primary attribute is annotated with
// annotations.Primary (@Primary). It is only effective when it is annotated on a singular boolean attribute
// as a sub attribute of a multiValued complex attribute. The purpose is that among all the element boolean
// properties carrying this attribute, at most one can assume a true value.
func (attr *Attribute) IsPrimary() bool {
	return attr.HasAnnotation(annotations.Primary)
}

// Return true if this attribute is considered an identity attribute. An identity attribute is annotated with
// annotations.Identity (@Identity). When one or more sub attributes for a complex attribute is annotated with
// @Identity, they will collectively be used to determine the identity of the complex attribute. Other attributes
// that were not annotated with @Identity no longer participates in operations like equality which involves identity.
func (attr *Attribute) IsIdentity() bool {
	return attr.HasAnnotation(annotations.Identity)
}

// Return the name of the attribute.
func (attr *Attribute) Name() string {
	return attr.name
}

// Return human-readable text to describe the attribute.
func (attr *Attribute) Description() string {
	return attr.description
}

// Return the data type of the attribute. The data type cannot be used to definitively
// determine the nature of the attribute unless combined with the method MultiValued.
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
func (attr *Attribute) ForEachCanonicalValue(callback func(canonicalValue string)) {
	for _, each := range attr.canonicalValues {
		callback(each)
	}
}

// Returns true if the a certain canonical value that meets the given criteria is defined among
// the canonical values of this attribute.
func (attr *Attribute) HasCanonicalValue(criteria func(value string) bool) bool {
	for _, each := range attr.canonicalValues {
		if criteria(each) {
			return true
		}
	}
	return false
}

// Return the total number of reference types
func (attr *Attribute) CountReferenceTypes() int {
	return len(attr.referenceTypes)
}

// Iterate through all reference types and invoke the callback function on each value.
func (attr *Attribute) ForEachReferenceType(callback func(referenceType string)) {
	for _, each := range attr.referenceTypes {
		callback(each)
	}
}

// Returns true if the a certain reference type that meets the given criteria is defined among
// the reference types of this attribute.
func (attr *Attribute) HasReferenceType(criteria func(value string) bool) bool {
	for _, each := range attr.referenceTypes {
		if criteria(each) {
			return true
		}
	}
	return false
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

// Return the sub attribute that goes by the name, or nil
func (attr *Attribute) SubAttributeForName(name string) *Attribute {
	for _, eachSubAttribute := range attr.subAttributes {
		if eachSubAttribute.GoesBy(name) {
			return eachSubAttribute
		}
	}
	return nil
}

// Return true if one or more of this attribute's sub attributes is marked as identity
func (attr *Attribute) HasIdentitySubAttributes() bool {
	for _, subAttr := range attr.subAttributes {
		if subAttr.IsIdentity() {
			return true
		}
	}
	return false
}

// Return true if one of this attribute's sub attribute is marked as primary.
func (attr *Attribute) HasPrimarySubAttribute() bool {
	for _, subAttr := range attr.subAttributes {
		if subAttr.IsPrimary() {
			return true
		}
	}
	return false
}

// Return the number of total annotations on this attribute.
func (attr *Attribute) CountAnnotations() int {
	return len(attr.annotationIndex)
}

// Returns true if this attribute is annotated with the given value
func (attr *Attribute) HasAnnotation(annotation string) bool {
	_, ok := attr.annotationIndex[annotation]
	return ok
}

// Iterate through all annotations on this attribute and invoke callback function on each.
func (attr *Attribute) ForEachAnnotation(callback func(annotation string)) {
	for annotation := range attr.annotationIndex {
		callback(annotation)
	}
}

// Return true if this attribute can be addressed by the given name. The method performs a case insensitive
// comparision of the provided name against the attribute's id, path, and name. If any matches, the attribute
// is considered addressable by that name.
func (attr *Attribute) GoesBy(name string) bool {
	switch strings.ToLower(name) {
	case strings.ToLower(attr.id), strings.ToLower(attr.path), strings.ToLower(attr.name):
		return true
	default:
		return false
	}
}

// Returns true if this attribute is the derived element attribute of the other attribute.
func (attr *Attribute) IsElementAttributeOf(other *Attribute) bool {
	return fmt.Sprintf("%s%s", other.ID(), elemSuffix) == attr.ID()
}

// Create the element attribute from this attribute. The element attribute is a derived attribute design for
// elements of a multiValued attribute. The ID of the attribute is suffixed with elemSuffix ("$elem"), and
// the multiValued attribute is set to false. All other attributes are carried over.
func (attr *Attribute) NewElementAttribute(annotations ...string) *Attribute {
	elemAttr := &Attribute{
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
		id:              fmt.Sprintf("%s%s", attr.ID(), elemSuffix),
		index:           attr.index,
		path:            attr.path,
		annotationIndex: map[string]struct{}{},
	}
	for _, annotation := range annotations {
		elemAttr.annotationIndex[annotation] = struct{}{}
	}
	return elemAttr
}

// Sort the sub attributes recursively based on the index.
func (attr *Attribute) Sort() {
	attr.ForEachSubAttribute(func(subAttribute *Attribute) {
		subAttribute.Sort()
	})
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

// Returns true if the two attributes are equal. This method checks pointer equality first.
// If does not match, goes on to check whether id matches.
func (attr *Attribute) Equals(other *Attribute) bool {
	return (attr == other) || attr.id == other.id
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
	attr.Sort()
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
		Primary         bool                    `json:"_primary"`		// deprecated
		Identity        bool                    `json:"_identity"`		// deprecated
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
	attr.annotationIndex = map[string]struct{}{}
	for _, annotation := range u.Annotations {
		attr.annotationIndex[annotation] = struct{}{}
	}
	if len(u.SubAttributes) > 0 {
		attr.subAttributes = make([]*Attribute, len(u.SubAttributes), len(u.SubAttributes))
		for i, sub := range u.SubAttributes {
			subAttribute := new(Attribute)
			sub.fill(subAttribute)
			attr.subAttributes[i] = subAttribute
		}
	}
}

type attributeValidator struct {
	attr *Attribute
}

func (v *attributeValidator) checkRequired() error {
	for _, required := range []struct {
		v    string
		name string
	}{
		{v: v.attr.id, name: "id"},
		{v: v.attr.name, name: "name"},
		{v: v.attr.path, name: "path"},
	} {
		if len(required.v) == 0 {
			return fmt.Errorf("'%s' is required in attribute", required.name)
		}
	}
	return nil
}

func (v *attributeValidator) checkCaseExact() error {
	nonCaseExactAllowed := map[Type]bool{
		TypeString:    true,
		TypeInteger:   true,
		TypeDecimal:   true,
		TypeBoolean:   true,
		TypeReference: false,
		TypeDateTime:  false,
		TypeBinary:    false,
		TypeComplex:   false,
	}[v.attr.Type()]
	if !v.attr.CaseExact() && !nonCaseExactAllowed {
		return fmt.Errorf("%s attribute must be case exact", v.attr.Type().String())
	}
	return nil
}

func (v *attributeValidator) checkUniqueness() error {
	uniquenessAllowed := map[Type]bool{
		TypeString:    true,
		TypeInteger:   true,
		TypeDecimal:   true,
		TypeBoolean:   true,
		TypeReference: true,
		TypeDateTime:  true,
		TypeBinary:    false,
		TypeComplex:   false,
	}[v.attr.Type()]
	if v.attr.Uniqueness() != UniquenessNone && !uniquenessAllowed {
		return fmt.Errorf("%s attribute must not be unique", v.attr.Type().String())
	}
	return nil
}

func (v *attributeValidator) checkPrimary() error {
	count := 0
	for _, subAttr := range v.attr.subAttributes {
		if subAttr.IsPrimary() {
			if subAttr.Type() != TypeBoolean || subAttr.MultiValued() {
				return fmt.Errorf("only singular boolean attribute can be annotated %s (violation %s)",
					annotations.Primary, v.attr.ID())
			}
			count++
		}
	}
	switch count {
	case 0:
	case 1:
		if v.attr.SingleValued() {
			return fmt.Errorf(
				"only multiValued complex property can annotate %s on its boolean sub attribute (violation: %s)",
				annotations.Primary,
				v.attr.ID())
		}
	default:
		return fmt.Errorf("only one boolean sub attribute can be annotated %s (violation: %s)",
			annotations.Primary, v.attr.ID())
	}
	return nil
}

var (
	_ json.Marshaler   = (*Attribute)(nil)
	_ json.Unmarshaler = (*Attribute)(nil)
)
