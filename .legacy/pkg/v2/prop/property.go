package prop

import (
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// Property holds a piece of data and is describe by an Attribute. The data requirement of the property is described by
// an enclosing attribute returned by the Attribute method. The enclosed data may be accessed via Raw and IsUnassigned
// method, and may be modified by Add, Replace and Delete method.
//
// A Property may enclose other properties. Such property is known to be a container property. Default cases of container
// property are the singleValued complex property and the multiValued property, as defined in SCIM. A non-container property
// must return 0 to CountChildren and is generally indifferent to CountChildren, ForEachChild, FindChild and ChildAtIndex
// methods.
type Property interface {
	// Attribute always returns a non-nil attribute to describe this property.
	Attribute() *spec.Attribute
	// Raw return the property's value in Golang's native type. The type correspondence are:
	// 	SCIM string <-> Go string
	//	SCIM integer <-> Go int64
	//	SCIM decimal <-> Go float64
	//	SCIM boolean <-> Go bool
	//	SCIM dateTime <-> Go string
	//	SCIM reference <-> Go string
	//	SCIM binary <-> Go string
	//	SCIM complex <-> Go map[string]interface{}
	//	SCIM multiValued <-> Go []interface{}
	// Property implementations are obliged to return in these types, or return a nil when unassigned. However,
	// implementations are not obliged to represent data in these types internally.
	Raw() interface{}
	// IsUnassigned return true if this property is unassigned. Unassigned is defined to be nil for singular non-complex
	// typed properties; empty for multiValued properties; and complex properties are unassigned if and only
	// if all its containing sub properties are unassigned.
	IsUnassigned() bool
	// Dirty returns true if any of Add/Replace/Delete method was ever called.
	// This method is necessary to distinguish between a naturally unassigned state and a deleted unassigned state.
	// When a property is first constructed, it has no value, and hence lies in an unassigned state. However, this
	// shall not be considered the same with a property returning to an unassigned state after having its value deleted.
	// Such difference is important as the SCIM specification mandates that user may override the server generated
	// default value for a readWrite property by explicitly providing "null" or "[]" (in case of multiValued) in the
	// JSON payload. In such situations, value generators should back off when an unassigned property is also dirty.
	Dirty() bool
	// Hash returns the hash value of this property's value. This will be helpful in comparing two properties.
	// Unassigned values shall return a hash of 0 (zero). Although this will create a potential hash
	// collision, we avoid this problem by checking the unassigned case first before comparing hashes.
	Hash() uint64
	// Matches return true if the two properties match each other. Properties match if and only if
	// their attributes match and their values are comparable according to the attribute.
	// 	For properties carrying singular non-complex attributes, attributes and values are compared.
	// 	For complex properties, two complex properties match if and only if all their identity sub properties match.
	// 	For multiValued properties, match only happens when they have the same number of element properties
	//	and the element properties all match correspondingly.
	// Two unassigned properties with the same attribute matches each other.
	Matches(another Property) bool
	// Clone return an exact clone of the property. The cloned property may share the same instance of attribute and
	// subscribers, but must retain individual copies of their values.
	Clone() Property
	// Add a value to the property and emit an event describing the change.
	// If the value already exists, no change will be made and the emitted event is nil. Otherwise, the value will
	// be added to the underlying data structure and mark the value dirty. For simple properties, calling this
	// method equates to calling Replace.
	Add(value interface{}) (*Event, error)
	// Replace value of this property and emit an event describing the change.
	// If the value equals to the current value, no change will be made and the emitted event is nil. Otherwise,
	// the underlying value will be replaced. Providing a nil value equates to calling Delete.
	Replace(value interface{}) (*Event, error)
	// Delete value from this property and emit an event describing the change.
	// If the property is already unassigned, deleting it again has no effect.
	Delete() (*Event, error)
	// Notify all subscribers of this property of the events.
	Notify(events *Events) error
	// CountChildren returns the number of children properties. Children properties are sub properties for complex
	// properties and element properties for multiValued properties. Other properties have no children.
	CountChildren() int
	// ForEachChild iterates all children properties and invoke callback function.
	ForEachChild(callback func(index int, child Property) error) error
	// FindChild returns the first children property that satisfies the criteria, or nil if none satisfies.
	FindChild(criteria func(child Property) bool) Property
	// ChildAtIndex returns the children property at given index. The type of index vary across implementations.
	ChildAtIndex(index interface{}) (Property, error)
}
