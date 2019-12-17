package prop

import (
	"github.com/imulab/go-scim/core/errors"
)

// Create a new navigator to traverse the property.
func NewNavigator(source Property) *Navigator {
	return &Navigator{
		stack: []Property{source},
	}
}

// Navigator is a stack of properties to let caller take control of
// how to traverse through the property structure, as opposed to
// Visitor and DFS traversal where the structure itself is in control.
type Navigator struct {
	stack []Property
}

// Focus on the sub property that goes by the given name (case insensitive), and
// return that sub property, or a noTarget error. The returned property will be
// the newly focused property, as reflected in the Current method. This method is
// intended to be used when the currently focused property is singular complex
// property. Any other property will yield a noTarget error.
func (n *Navigator) FocusName(name string) (Property, error) {
	container, ok := n.Current().(Container)
	if !ok {
		return nil, n.errNoTargetByName(n.Current(), name)
	}

	child := container.ChildAtIndex(name)
	if child == nil {
		return nil, n.errNoTargetByName(container, name)
	}
	n.stack = append(n.stack, child)

	return n.Current(), nil
}

// Focus on the element at index given index, and return the element property, or a
// noTarget error. The returned property will be the newly focused property, as reflected
// in the Current method. This method is intended to be used when the currently focused
// property is a multiValued property. Any other property will yield a noTarget error. If
// index is out of range, a noTarget error is also returned.
func (n *Navigator) FocusIndex(index int) (Property, error) {
	container, ok := n.Current().(Container)
	if !ok {
		return nil, n.errNoTargetByIndex(n.Current(), index)
	}

	child := container.ChildAtIndex(index)
	if child == nil {
		return nil, n.errNoTargetByIndex(container, index)
	}
	n.stack = append(n.stack, child)

	return n.Current(), nil
}

// Focus on the element that meets given criteria, and return the focused property, or a
// noTarget error. The returned property will be the newly focused property, as reflected
// in the Current method.
func (n *Navigator) FocusCriteria(criteria func(child Property) bool) (Property, error) {
	container, ok := n.Current().(Container)
	if !ok {
		return nil, n.errNoTargetByCriteria(n.Current())
	}

	var hit Property = nil
	{
		_ = container.ForEachChild(func(index int, child Property) error {
			if hit == nil && criteria(child) {
				hit = child
			}
			return nil
		})
	}

	if hit == nil {
		return nil, n.errNoTargetByCriteria(n.Current())
	}

	n.stack = append(n.stack, hit)
	return n.Current(), nil
}

// Return the number of properties that was focused, including the currently focused. These
// properties, excluding the current one, can be refocused by calling Retract, one at a time
// in the reversed order that were focused. The minimum depth is one.
func (n *Navigator) Depth() int {
	return len(n.stack)
}

// Return the currently focused property.
func (n *Navigator) Current() Property {
	return n.stack[len(n.stack)-1]
}

// Go back to the last focused property. Note that the source property that this navigator
// is created with in NewNavigator cannot be retracted.
func (n *Navigator) Retract() {
	if n.Depth() > 1 {
		n.stack = n.stack[:len(n.stack)-1]
	}
}

func (n *Navigator) errNoTargetByName(container Property, name string) error {
	return errors.NoTarget("property '%s' has no sub property named '%s'", container.Attribute().Path(), name)
}

func (n *Navigator) errNoTargetByIndex(container Property, index int) error {
	return errors.NoTarget("property '%s' has no element at index %d", container.Attribute().Path(), index)
}

func (n *Navigator) errNoTargetByCriteria(container Property) error {
	return errors.NoTarget("property '%s' has no element meeting the given criteria", container.Attribute().Path())
}

// Create a fluent navigator.
func NewFluentNavigator(source Property) *FluentNavigator {
	return &FluentNavigator{
		nav: NewNavigator(source),
	}
}

// FluentNavigator is wrapper around Navigator. It exposes the same API as Navigator, but
// in a fluent way. In addition, any intermediate error caused by FocusXXX methods are cached
// internally and turns further FocusXXX methods into a no-op.
type FluentNavigator struct {
	nav 	*Navigator
	err 	error
}

func (n *FluentNavigator) Error() error {
	return n.err
}

func (n *FluentNavigator) FocusName(name string) *FluentNavigator {
	if n.err == nil {
		_, n.err = n.nav.FocusName(name)
	}
	return n
}

func (n *FluentNavigator) FocusIndex(index int) *FluentNavigator {
	if n.err == nil {
		_, n.err = n.nav.FocusIndex(index)
	}
	return n
}

func (n *FluentNavigator) FocusCriteria(criteria func(child Property) bool) *FluentNavigator {
	if n.err == nil {
		_, n.err = n.nav.FocusCriteria(criteria)
	}
	return n
}

func (n *FluentNavigator) Depth() int {
	return n.nav.Depth()
}

func (n *FluentNavigator) Current() Property {
	return n.nav.Current()
}

func (n *FluentNavigator) CurrentAsContainer() Container {
	return n.Current().(Container)
}

func (n *FluentNavigator) Retract() *FluentNavigator {
	n.nav.Retract()
	return n
}