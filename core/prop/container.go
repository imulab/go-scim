package prop

const childNotCreated = -1

// Special SCIM property that contains other property. This interface shall be implemented by complex and multiValued
// properties. A container property usually does not hold value on its own. Instead, it serves as a collection of values
// for its sub properties (i.e. complex property) or element properties (i.e. multiValued property), these contained
// properties are called "children properties" in this context.
type Container interface {
	Property

	// Return the number of children properties.
	CountChildren() int

	// Iterate through all children properties and invoke the callback function sequentially.
	// The callback method SHALL NOT block the executing Goroutine.
	ForEachChild(callback func(index int, child Property) error) error

	// Return the child property addressable by the index, or nil.
	ChildAtIndex(index interface{}) Property

	// Add a prototype of the child property to the container,
	// and return the index. Return childNotCreated (-1) to indicate no child was created.
	NewChild() int

	// Consolidate and remove any unwanted child properties. Delete operations may leave some
	// children properties unassigned. The container may decide to remove them from its collection.
	Compact()

	// Propagate the event for a child property.
	Propagate(e *Event) error
}
