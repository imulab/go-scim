package core

import "fmt"

type (
	// The attribute model that is used through out SCIM to
	// determine data type and constraint.
	Attribute struct {
		id              string
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
	}
)

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

// Return an exact shallow copy of the attribute, but with the multiValued field set to false, thus
// effectively converting any multiValued attribute to singular attribute.
func (attr *Attribute) AsSingleValued() *Attribute {
	return &Attribute{
		id:              attr.id,
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
	}
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