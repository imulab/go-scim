package core

import (
	"fmt"
)

// A navigator lets the caller assume control of the navigation in the resource's property structure. The caller
// may focus and de-focus (release) on a named property at any time. The pace is usually dictated by external data
// structures, which is common when performing works like deserialization. The navigator also exposes Crud methods
// to alter value on the currently focused property.
type Navigator interface {
	// Returns the property which matches the selector, or returned an error. The selector can be a string based
	// name or a int based index. A string based name will attempt to select a sub property, assuming the current
	// property is a single valued complex property. An int based index will attempt to select an element property,
	// assuming the current property is a multiValued property.
	Focus(selector interface{}) (Property, error)
	// Returns the property that was most recently focused. If no property was focused, return
	// the top property.
	Current() Property
	// Stop focusing on the current property and go back to the last focused property.
	Release()
	// When the current property is a multiValued property, insert a prototype element property into
	// it and return its index. The caller can then use Focus with the index to focus on the prototype
	// element and continue operating. When the current property is not a multiValued property, return -1.
	NewPrototype() int
	// Add value to the current property
	Add(value interface{}) error
	// Replace a value to the current property.
	Replace(value interface{}) error
	// Delete value from current property.
	Delete() error
}

// Create a new navigator on the resource.
func NewNavigator(resource *Resource) Navigator {
	return &defaultNavigator{
		stack: []Property{resource.base},
	}
}

// default implementation of navigator
type defaultNavigator struct {
	stack []Property
}

func (n *defaultNavigator) Focus(selector interface{}) (Property, error) {
	switch selector.(type) {
	case string:
		return n.focusSub(selector.(string))
	case int:
		return n.focusElement(selector.(int))
	default:
		panic("invalid argument: expects string or int")
	}
}

func (n *defaultNavigator) focusSub(selector string) (Property, error) {
	current := n.Current()

	if current.Attribute().MultiValued || current.Attribute().Type != TypeComplex {
		return nil, Errors.noTarget(
			fmt.Sprintf("sub path '%s' does not exist under '%s'", selector, current.Attribute().DisplayName()),
		)
	}

	subProp, err := current.(*complexProperty).getSubProperty(selector)
	if err != nil {
		return nil, err
	}
	n.stack = append(n.stack, subProp)

	return subProp, nil
}

func (n *defaultNavigator) focusElement(selector int) (Property, error) {
	current := n.Current()

	if !current.Attribute().MultiValued || len(current.(*multiValuedProperty).props) <= selector {
		return nil, Errors.noTarget(
			fmt.Sprintf("item of index '%d' does not exist under '%s'", selector, current.Attribute().DisplayName()),
		)
	}

	elemProp := current.(*multiValuedProperty).props[selector]
	n.stack = append(n.stack, elemProp)

	return elemProp, nil
}

func (n *defaultNavigator) Current() Property {
	if len(n.stack) == 0 {
		panic("failed to get current property: zero length stack")
	}
	return n.stack[len(n.stack)-1]
}

func (n *defaultNavigator) Release() {
	if len(n.stack) == 0 {
		panic("failed to release property: zero length stack")
	}
	n.stack = n.stack[:len(n.stack)-1]
}

func (n *defaultNavigator) NewPrototype() int {
	if !n.Current().Attribute().MultiValued {
		return -1
	}

	current := n.Current().(*multiValuedProperty)
	current.props = append(current.props, Properties.New(current.Attribute().ToSingleValued()))
	return len(current.props) - 1
}

func (n *defaultNavigator) Add(value interface{}) error {
	return n.Current().(Crud).Add(nil, value)
}

func (n *defaultNavigator) Replace(value interface{}) error {
	return n.Current().(Crud).Replace(nil, value)
}

func (n *defaultNavigator) Delete() error {
	return n.Current().(Crud).Delete(nil)
}
