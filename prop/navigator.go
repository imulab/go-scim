package prop

import (
	"fmt"
	"github.com/elvsn/scim.go/spec"
)

// Navigate returns a navigator that allows caller to freely navigate the property structure and maintains the navigation
// history to enable retraction at any time. The navigator also exposes delegate methods to modify the property, and
// propagate modification events to upstream properties.
func Navigate(property Property) *Navigator {
	return &Navigator{stack: []Property{property}}
}

type Navigator struct {
	stack []Property
	err   error
}

// Error returns any error occurred during navigation. An error will prevent further navigation taking place.
func (n *Navigator) Error() error {
	return n.err
}

// Depth return the number of properties that was focused, including the currently focused. These
// properties, excluding the current one, can be refocused by calling Retract, one at a time
// in the reversed order that were focused. The minimum depth is one.
func (n *Navigator) Depth() int {
	return len(n.stack)
}

// Current returns the currently focused property.
func (n *Navigator) Current() Property {
	return n.stack[len(n.stack)-1]
}

// Retract goes back to the last focused property. The source property that this navigator was created with cannot be retracted.
func (n *Navigator) Retract() {
	if n.Depth() > 1 {
		n.stack = n.stack[:len(n.stack)-1]
	}
}

// Dot focuses on the sub property that goes by the given name (case insensitive)
func (n *Navigator) Dot(name string) *Navigator {
	if n.err != nil {
		return n
	}

	child, err := n.Current().childAtIndex(name)
	if err != nil {
		n.err = fmt.Errorf("%w: no target named '%s' from '%s'", spec.ErrNoTarget, name, n.Current().Attribute().Path())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

// At focuses on the element property at given index.
func (n *Navigator) At(index int) *Navigator {
	if n.err != nil {
		return n
	}

	child, err := n.Current().childAtIndex(index)
	if err != nil {
		n.err = fmt.Errorf("%w: no target at index '%d' from '%s'", spec.ErrNoTarget, index, n.Current().Attribute().Path())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

// Where focuses on the first child property meeting given criteria.
func (n *Navigator) Where(criteria func(child Property) bool) *Navigator {
	if n.err != nil {
		return n
	}

	child := n.Current().findChild(criteria)
	if child == nil {
		n.err = fmt.Errorf("%w: no target meeting criteria from '%s'", spec.ErrNoTarget, n.Current().Attribute().Path())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

// Add delegates for Add of the Current property and propagates events to upstream properties.
func (n *Navigator) Add(value interface{}) error {
	return n.delegateMod(func() (event *Event, err error) {
		return n.Current().Add(value)
	})
}

// Replace delegates for Replace of the Current property and propagates events to upstream properties.
func (n *Navigator) Replace(value interface{}) error {
	return n.delegateMod(func() (event *Event, err error) {
		return n.Current().Replace(value)
	})
}

// Delete delegates for Delete of the Current property and propagates events to upstream properties.
func (n *Navigator) Delete() error {
	return n.delegateMod(func() (event *Event, err error) {
		return n.Current().Delete()
	})
}

func (n *Navigator) delegateMod(mod func() (*Event, error)) error {
	if n.err != nil {
		return n.err
	}

	ev, err := mod()
	if err != nil {
		return err
	}

	if ev != nil {
		events := ev.ToEvents()
		for i := len(n.stack) - 1; i >= 0; i-- {
			if err := n.stack[i].Notify(events); err != nil {
				return err
			}
		}
	}

	return nil
}
