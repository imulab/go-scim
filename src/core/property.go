package core

// Interface to hold a piece of data in SCIM. Property is the basic unit of all iterations
// in SCIM.
type Property interface {
	// Return a non-nil attribute about this property
	Attribute() *Attribute
	// Return the property's value in Golang's native type, or nil.
	// Implementations shall document the type returned here.
	Raw() interface{}
	// Return true if this property is unassigned, and true if the unassigned state
	// is considered dirty.
	// Unassigned is defined to be nil for singular simple typed properties; empty for multiValued
	// properties; and complex properties are unassigned if and only if all its
	// containing sub properties are unassigned.
	// Dirtiness is defined to have been explicitly nullified by the user, as opposed to naturally
	// unassigned when the property is constructed from its attribute. When the property is freshly
	// constructed, it is not dirty. If the property remain untouched throughout its life time, it will
	// remain unassigned and not dirty. If the property is explicitly unassigned through Add, Replace and
	// Delete operations, it shall be considered dirty. Dirty unassigned behaves differently when it comes
	// to default value assignments.
	IsUnassigned() (unassigned bool, dirty bool)
	// Return the number of children properties. Simple typed properties always return 0.
	// Complex and multiValued properties return the number of sub properties and the number
	// of element properties respectively.
	CountChildren() int
	// Iterate through all children properties and invoke the callback function sequentially.
	// The callback method SHALL NOT block the executing Goroutine.
	ForEachChild(callback func(child Property))
	// Return true if the two properties match each other. Properties match if and only if
	// their attributes match and their values are comparable according to the attribute.
	// For complex properties, two complex properties match if and only if all their sub properties
	// match. For multiValued properties, match only happens when they have the same number of
	// element properties and the element properties all match correspondingly. Two unassigned
	// properties with the same attribute matches each other.
	Matches(another Property) bool
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
	// Add a value to the property. Operation shall render the property dirty. If property's value does
	// not change after the operation, false will be returned. If value is compatible, err will be returned.
	Add(value interface{}) (bool, error)
	// Replace value of this property. Operation shall render the property dirty. If property's value does
	//	// not change after the operation, false will be returned. If value is compatible, err will be returned.
	Replace(value interface{}) (bool, error)
	// Delete value from this property. Operation shall render the property dirty. If property's value does
	//	// not change after the operation, false will be returned.
	Delete() (bool, error)
	// Consolidate and remove any unwanted child properties
	Compact()
}
