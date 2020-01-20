package prop

import "github.com/imulab/go-scim/core/spec"

// Property holds a piece of data and is describe by an Attribute. A property can exist by itself, or it can exist
// as a sub property of another property, in which case, the containing property is known as a container property.
type Property interface {
	// Return a non-nil attribute to describe this property.
	Attribute() *spec.Attribute

	// Return the parent container of this property. If this attribute has no
	// parent container, return nil.
	Parent() Container

	// Return the property's value in Golang's native type. The chosen types corresponding to
	// SCIM attribute types are:
	// string - string
	// integer - int64
	// decimal - float64
	// boolean - bool
	// dateTime - string
	// reference - string
	// binary - string
	// complex - map[string]interface{}
	// multiValued - []interface{}
	// Property implementations are obliged to return in these types, or return a nil in case of
	// unassigned. However, implementations are not obliged to represent data in these types internally.
	Raw() interface{}

	// Return true if this property is unassigned. Unassigned is defined to be nil for singular non-complex
	// typed properties; empty for multiValued properties; and complex properties are unassigned if and only
	// if all its containing sub properties are unassigned.
	IsUnassigned() bool

	// Return true if the two properties match each other. Properties match if and only if
	// their attributes match and their values are comparable according to the attribute.
	// For properties carrying singular non-complex attributes, attributes and values are compared.
	// For complex properties, two complex properties match if and only if all their identity sub
	// properties match.
	// For multiValued properties, match only happens when they have the same number of element properties
	// and the element properties all match correspondingly.
	// Two unassigned properties with the same attribute matches each other.
	Matches(another Property) bool

	// Returns the hash value of this property's value. This will be helpful in comparing two properties.
	// Unassigned values shall return a hash of 0 (zero). Although this will create a potential hash
	// collision, we avoid this problem by checking the unassigned case first before comparing hashes.
	Hash() uint64

	// Return true if the property's value is equal to the given value. If the given value
	// is nil, always return false. This method corresponds to the 'eq' filter operator.
	// If implementation cannot apply the 'eq' operator, return an error.
	EqualsTo(value interface{}) (bool, error)

	// Return true if the property's value starts with the given value. If the given value
	// is nil, always return false. This method corresponds to the 'sw' filter operator.
	// If implementation cannot apply the 'sw' operator, return an error.
	StartsWith(value string) (bool, error)

	// Return true if the property's value ends with the given value. If the given value
	// is nil, always return false. This method corresponds to the 'ew' filter operator.
	// If implementation cannot apply the 'ew' operator, return an error.
	EndsWith(value string) (bool, error)

	// Return true if the property's value contains the given value. If the given value
	// is nil, always return false. This method corresponds to the 'co' filter operator.
	// If implementation cannot apply the 'co' operator, return an error.
	Contains(value string) (bool, error)

	// Return true if the property's value is greater than the given value. If the given value
	// is nil, always return false. This method corresponds to the 'gt' filter operator.
	// If implementation cannot apply the 'gt' operator, return an error.
	GreaterThan(value interface{}) (bool, error)

	// Return true if the property's value is greater than the given value. If the given value
	// is nil, always return false. This method corresponds to the 'lt' filter operator.
	// If implementation cannot apply the 'lt' operator, return an error.
	LessThan(value interface{}) (bool, error)

	// Return true if the property's value is present. Presence is defined to be non-nil and non-empty.
	// This method corresponds to the 'pr' operator and shall be implemented by all implementations.
	Present() bool

	// Add a value to the property. If the value already exists, no change will be made. Otherwise, the value will
	// be added to the underlying data structure and mod count increased by one. For simple properties, calling this
	// method equates to calling Replace.
	Add(value interface{}) error

	// Replace value of this property. If the value equals to the current value, no change will be made. Otherwise,
	// the underlying value will be replaced. Providing a nil value equates to calling Delete.
	Replace(value interface{}) error

	// Delete value from this property. If the property is already unassigned, deleting it again has no effect.
	Delete() error

	// Returns true if any of Add/Replace/Delete method was ever called. This method is necessary to distinguish
	// between a naturally unassigned state and a deleted unassigned state. When a property is first constructed,
	// it has no value, and hence lies in an unassigned state. However, this shall not be considered the same with
	// a property returning to an unassigned state after having its value deleted. Such difference is important as
	// the SCIM specification mandates that user may override the server generated default value for a readWrite
	// property by explicitly providing "null" or "[]" (in case of multiValued) in the JSON payload. In such situations,
	// value generators should back off when an unassigned property is also dirty.
	Dirty() bool

	// Add the subscriber to the properties emitted events
	Subscribe(subscriber Subscriber)

	// Return an exact clone of the property. The cloned property may share the same instance of attribute and
	// subscribers, but must retain individual copies of their values. The cloned property should be attached to
	// the new parent container, which usually is the result of previous Clone call.
	Clone(parent Container) Property
}
