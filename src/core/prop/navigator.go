package prop

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
)

func NewNavigator(source core.Property) *Navigator {
	return &Navigator{stack: []core.Property{source}}
}

// Navigator is a stack of properties to let caller take control of
// how to traverse through the property structure, as opposed to
// Visitor and DFS traversal where the structure itself is in control.
type Navigator struct {
	stack	[]core.Property
}

// Focus on the sub property that goes by the given name (case insensitive), and
// return that sub property, or a noTarget error. The returned property will be
// the newly focused property, as reflected in the Current method. This method is
// intended to be used when the currently focused property is singular complex
// property. Any other property will yield a noTarget error.
func (n *Navigator) FocusName(name string) (core.Property, error) {
	container, ok := n.Current().(core.Container)
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
func (n *Navigator) FocusIndex(index int) (core.Property, error) {
	container, ok := n.Current().(core.Container)
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

// Return the number of properties that was focused, including the currently focused. These
// properties, excluding the current one, can be refocused by calling Retract, one at a time
// in the reversed order that were focused. The minimum depth is one.
func (n *Navigator) Depth() int {
	return len(n.stack)
}

// Return the currently focused property.
func (n *Navigator) Current() core.Property {
	return n.stack[len(n.stack)-1]
}

// Go back to the last focused property. Note that the source property that this navigator
// is created with in NewNavigator cannot be retracted.
func (n *Navigator) Retract() {
	if n.Depth() > 1 {
		n.stack = n.stack[:len(n.stack)-1]
	}
}

func (n *Navigator) errNoTargetByName(container core.Property, name string) error {
	return errors.NoTarget("property '%s' has no sub property named '%s'", container.Attribute().Path(), name)
}

func (n *Navigator) errNoTargetByIndex(container core.Property, index int) error {
	return errors.NoTarget("property '%s' has no element at index %d", container.Attribute().Path(), index)
}
