package core

// Interface to hold a piece of data in SCIM. Property is the basic unit of all iterations
// in SCIM.
type Property interface {
	// Return a non-nil attribute about this property
	Attribute() *Attribute
	// Return the property's value in Golang's native type, or nil.
	// Implementations shall document the type returned here.
	Raw() interface{}
	// Return true if this property is unassigned.
	// Unassigned is defined to be nil for singular simple typed properties; empty for multiValued
	// properties; and complex properties are unassigned if and only if all its
	// containing sub properties are unassigned.
	IsUnassigned() bool
	// Returns the number of times this property has been modified. In principal, the mod count increases
	// every time a change is made to the property. Note that this does not equate calling Add, Replace or
	// Delete as these methods may not modify the underlying data if the change proposed has no effect on
	// it.
	ModCount() int
	// Return true if the two properties match each other. Properties match if and only if
	// their attributes match and their values are comparable according to the attribute.
	// For complex properties, two complex properties match if and only if all their sub properties
	// match. For multiValued properties, match only happens when they have the same number of
	// element properties and the element properties all match correspondingly. Two unassigned
	// properties with the same attribute matches each other.
	Matches(another Property) bool
	// Returns the hash of this property's value. This will be helpful in comparing two properties.
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
	// Perform a depth-first-search on the property and invoke the callback function on each visited
	// property from the search. The callback function SHALL NOT block.
	DFS(callback func(property Property))
	// Add a value to the property. If the value already exists, no change will be made. Otherwise, the value will
	// be added to the underlying data structure and mod count increased by one. For simple properties, calling this
	// method equates to calling Replace.
	Add(value interface{}) (bool, error)
	// Replace value of this property. If the value equals to the current value, no change will be made. Otherwise,
	// the underlying value will be replaced and mod count increased by one. Providing a nil value equates to
	// calling Delete.
	Replace(value interface{}) (bool, error)
	// Delete value from this property. If the property is already unassigned, deleting it again has no effect and
	// does not increase mod count. However, as a special case, when mod count is 0, deleting it will increase mod
	// count. This behaviour is designed to distinguish between a system generated unassigned property and user declared
	// unassigned property.
	Delete() (bool, error)
}

// Interface for SCIM properties that serve as a container to other properties. For instance, complex and multiValued
// properties. These properties have additional features to manipulate their contained/children properties.
type Container interface {
	Property
	// Return the number of children properties.
	CountChildren() int
	// Iterate through all children properties and invoke the callback function sequentially.
	// The callback method SHALL NOT block the executing Goroutine.
	ForEachChild(callback func(index int, child Property))
	// Return the child property addressable by the index, or nil.
	ChildAtIndex(index interface{}) Property
	// Add a prototype of the child property to the container,
	// and return the index. Return -1 to indicate no child was created.
	NewChild() int
	// Consolidate and remove any unwanted child properties
	Compact()
}