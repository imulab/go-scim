package spec

import (
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/spec/internal"
	"reflect"
	"sort"
	"strings"
)

// Attribute models a superset of defined SCIM attributes. It serves as the basic unit that describes data requirement
// this project.
//
// In addition to attributes defined in SCIM, four additional attributes are defined: id, index, path and annotations.
// Id defines a unique identifier for this attribute, which by convention, should be the full URN name. Index defines
// a relative ascending sort index for the attribute among other attributes on the same level, which can be used to ensure
// proper ordering and nice presentation. Path is joined attribute names from the root, which is used to determine the
// full attribute name without further calculation. Annotations are of type map[string]map[string]interface{}, which is
// used as the major extension point to define additional behaviour. It is widely used by other processing mechanisms
// in this module. These four additional attributes have JSON names of "id", "_index", "_path" and "_annotations"
// respectively.
//
// As an example, a typical attribute can be defined in JSON as:
//	{
//		"id": "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
//		"name": "familyName",
//		"type": "string",
//		"multiValued": false,
//		"required": false,
//		"caseExact": false,
//		"mutability": "readWrite",
// 		"returned": "default",
// 		"uniqueness": "none",
//		"_index": 1,
//		"_path": "name.familyName",
//		"_annotations": {
//			"@Identity": {}
//		}
//	}
// The above example defines a typical "name.familyName" attribute in the User resource type. It can be universally
// identified by "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName"; can be placed at the second position (index 1)
// among sub attributes of "urn:ietf:params:scim:schemas:core:2.0:User:name"; has a path of "name.familyName" and has a
// single parameter-less annotation of "@Identity".
//
// Because Attribute is at the core of the SCIM protocol, it is designed to be read only as much as possible. All data
// access, including iteration, search operations, need to be done via accessor methods. For simple fields, look for the
// methods corresponding to the field name. For example, field name can be accessed via method Name. For array fields,
// look for methods prefixed with ForEach, Exists and Count to perform iteration, search and get-length operations.
//
// Attribute also introduces the notion of an element attribute. An element attribute is a derived attribute from a
// multiValued attribute in order to represent the data requirements of its elements. Because elements of multiValued
// properties can be generated and deleted on the fly, these attributes also need to be so. DeriveElementAttribute can
// be used to generate a new element attribute. IsElementAttributeOf can be used to check if the current attribute is a
// derived element attribute.
//
// As of now, Attribute is parsed to and from JSON using special adapter structures that exposes and hides certain fields.
// This design is subject to change when we move to treat Schema as just another resource.
// See also:
//	issue https://github.com/imulab/go-scim/issues/40
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

// ID returns the id of the attribute that globally identifies the attribute.
// Attribute ids are in the format of <schema_urn>:<full_path>. Core attributes has no need to prefix the schema URN.
// For instance, "schemas", "meta.version", "urn:ietf:params:scim:schemas:core:2.0:Group:displayName"
func (attr *Attribute) ID() string {
	return attr.id
}

// Path returns the full path of this attribute.
// The full path are the name of all attributes from the root to this attribute, delimited by period (".").
// For instance, "id", "meta.version", "emails.value".
func (attr *Attribute) Path() string {
	return attr.path
}

// Name returns the name of the attribute.
func (attr *Attribute) Name() string {
	return attr.name
}

// Description returns human-readable text to describe the attribute.
func (attr *Attribute) Description() string {
	return attr.description
}

// Type return the data type of the attribute.
func (attr *Attribute) Type() Type {
	return attr.typ
}

// MultiValued return whether the attribute allows several instance of properties to be defined.
func (attr *Attribute) MultiValued() bool {
	return attr.multiValued
}

// Required return true when the attribute is required.
func (attr *Attribute) Required() bool {
	return attr.required
}

// CaseExact returns true if the attribute's value is case sensitive. Case sensitivity only applies to string attributes.
func (attr *Attribute) CaseExact() bool {
	return attr.caseExact
}

// Mutability return the mutability definition of the attribute.
func (attr *Attribute) Mutability() Mutability {
	return attr.mutability
}

// Returned returns the returned definition of the attribute.
func (attr *Attribute) Returned() Returned {
	return attr.returned
}

// Uniqueness return the uniqueness definition of the attribute.
func (attr *Attribute) Uniqueness() Uniqueness {
	return attr.uniqueness
}

// ForEachCanonicalValues invokes callback function on each defined canonical values
func (attr *Attribute) ForEachCanonicalValues(callback func(canonicalValue string)) {
	for _, cv := range attr.canonicalValues {
		callback(cv)
	}
}

// ExistsCanonicalValue returns true if the canonical value that meets the criteria exists; false otherwise.
func (attr *Attribute) ExistsCanonicalValue(criteria func(canonicalValue string) bool) bool {
	for _, cv := range attr.canonicalValues {
		if criteria(cv) {
			return true
		}
	}
	return false
}

// CountCanonicalValues returns the total number of canonical values defined.
func (attr *Attribute) CountCanonicalValues() int {
	return len(attr.canonicalValues)
}

// ForEachReferenceTypes invokes callback function on each defined reference types
func (attr *Attribute) ForEachReferenceTypes(callback func(referenceType string)) {
	for _, rt := range attr.referenceTypes {
		callback(rt)
	}
}

// ExistsReferenceType returns true if the reference type that meets the criteria exists; false otherwise.
func (attr *Attribute) ExistsReferenceType(criteria func(referenceType string) bool) bool {
	for _, rt := range attr.referenceTypes {
		if criteria(rt) {
			return true
		}
	}
	return false
}

// CountReferenceTypes returns the total number of reference types.
func (attr *Attribute) CountReferenceTypes() int {
	return len(attr.referenceTypes)
}

// ForEachSubAttribute invokes callback function on each sub attribute.
func (attr *Attribute) ForEachSubAttribute(callback func(subAttribute *Attribute) error) error {
	for _, eachSubAttribute := range attr.subAttributes {
		if err := callback(eachSubAttribute); err != nil {
			return err
		}
	}
	return nil
}

// FindSubAttribute returns the sub attribute that matches the criteria, or returns nil if no sub attribute meets criteria.
func (attr *Attribute) FindSubAttribute(criteria func(subAttr *Attribute) bool) *Attribute {
	for _, subAttr := range attr.subAttributes {
		if criteria(subAttr) {
			return subAttr
		}
	}
	return nil
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

// Return the total number of sub attributes.
func (attr *Attribute) CountSubAttributes() int {
	return len(attr.subAttributes)
}

// GoesBy returns true if this attribute can be addressed by the given name.
func (attr *Attribute) GoesBy(name string) bool {
	switch strings.ToLower(name) {
	case strings.ToLower(attr.id), strings.ToLower(attr.path), strings.ToLower(attr.name):
		return true
	default:
		return false
	}
}

// DFS perform a depth-first-traversal on the given attribute and invokes callback
func (attr *Attribute) DFS(callback func(attr *Attribute)) {
	callback(attr)
	for _, each := range attr.subAttributes {
		each.DFS(callback)
	}
}

// Annotation returns the annotation parameters by the given name (case sensitive) and a boolean indicating whether
// this annotations exists.
func (attr *Attribute) Annotation(name string) (params map[string]interface{}, ok bool) {
	params, ok = attr.annotations[name]
	return
}

// ForEachAnnotation iterates through annotations and invoke callback.
func (attr *Attribute) ForEachAnnotation(callback func(annotation string, params map[string]interface{})) {
	for k, v := range attr.annotations {
		callback(k, v)
	}
}

// IsElementAttributeOf returns true if this attribute is the derived element attribute of the other attribute.
func (attr *Attribute) IsElementAttributeOf(other *Attribute) bool {
	return fmt.Sprintf("%s%s", other.ID(), elemSuffix) == attr.ID()
}

// DeriveElementAttribute create an element attribute of this attribute. This method is only meaningful when invoked
// on a multiValued attribute.
//
// The derived element attribute will inherit most properties from this attribute except for id, multiValued and annotations.
// Id will be suffixed with "$elem". MultiValued will be set to false. Annotations will be derived from the parameters of
// "@ElementAnnotations" from this attribute.
//
// For example, a multiValued attribute (certain fields omitted for brevity) like
//	{
//		"id": "urn:ietf:params:scim:schemas:core:2.0:User:emails",
//		"name": "emails",
//		"type": "complex",
//		"multiValued": true,
//		"_annotations": {
//			"@ElementAnnotations": {
//				"@StateSummary": {}
//			}
//		},
//		...
//	}
// will derive an element attribute of:
//	{
//		"id": "urn:ietf:params:scim:schemas:core:2.0:User:emails$elem",
//		"name": "emails",
//		"type": "complex",
//		"multiValued": false,
//		"_annotations": {
//			"@StateSummary": {}
//		},
//		...
//	}
//
// See also:
//	annotation.ElementAnnotations
func (attr *Attribute) DeriveElementAttribute() *Attribute {
	elemAttr := Attribute{
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
		annotations:     map[string]map[string]interface{}{},
	}

	if param, ok := attr.Annotation(annotation.ElementAnnotations); ok {
		for k, v := range param {
			if rv := reflect.ValueOf(v); rv.Kind() != reflect.Map {
				continue
			} else if rv.Len() == 0 {
				elemAttr.annotations[k] = map[string]interface{}{}
			} else {
				elemAttr.annotations[k] = v.(map[string]interface{})
			}
		}
	}

	return &elemAttr
}

// PublicValues returns a representation of this attribute's public values (non-extension) in a data structure
// that conforms to the definition of prop.Property#Raw. However, this method does not implement it.
func (attr *Attribute) PublicValues() interface{} {
	data := map[string]interface{}{
		"name":            attr.name,
		"description":     attr.description,
		"type":            attr.typ.String(),
		"multiValued":     attr.multiValued,
		"required":        attr.required,
		"caseExact":       attr.caseExact,
		"mutability":      attr.mutability.String(),
		"returned":        attr.returned.String(),
		"uniqueness":      attr.uniqueness.String(),
		"canonicalValues": attr.canonicalValues,
		"referenceTypes":  attr.referenceTypes,
	}
	if len(attr.subAttributes) > 0 {
		var subPV []interface{}
		for _, subAttr := range attr.subAttributes {
			subPV = append(subPV, subAttr.PublicValues())
		}
		data["subAttributes"] = subPV
	}
	return data
}

// Equals returns true if the two attributes are considered equal.
func (attr *Attribute) Equals(other *Attribute) bool {
	return (attr == other) || attr.id == other.id
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

const (
	elemSuffix = "$elem"
)
