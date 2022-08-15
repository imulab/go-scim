package scim

import (
	"fmt"
	"strings"
)

func newNavigator(prop Property) *navigator {
	return &navigator{stack: []Property{prop}}
}

// navigator is a controlled mechanism to traverse the Property tree. It should be used in cases where the caller has
// knowledge of what to access. For example, when de-serializing JSON into a Property, caller has knowledge of the JSON
// structure, therefore knows what to access in the Property structure.
//
// navigator does not return error on every error-able operation, implementation will mostly be stateful. Callers
// should call Error or hasError to check if the navigator is currently in the error state, after performing one or
// possibly several chained operations fluently.
//
// The navigation stack can be advanced by calling dot, at and where methods. Which methods to use depends on the context.
// The stack can be retracted by calling retract. The top and the bottom item on the stack can be queried by calling
// current and source. The stack depth is available via depth.
type navigator struct {
	stack []Property
	err   error
}

func (n *navigator) hasError() bool {
	return n.err != nil
}

func (n *navigator) clearError() {
	n.err = nil
}

func (n *navigator) depth() int {
	return len(n.stack)
}

func (n *navigator) source() Property {
	return n.stack[0]
}

func (n *navigator) current() Property {
	return n.stack[len(n.stack)-1]
}

func (n *navigator) retract() *navigator {
	if n.depth() > 1 {
		n.stack = n.stack[:len(n.stack)-1]
	}
	return n
}

func (n *navigator) dot(name string) *navigator {
	if n.err != nil {
		return n
	}

	child := n.current().ByIndex(name)
	if child == nil {
		n.err = fmt.Errorf("%w: no attribute named '%s' from '%s'", ErrInvalidPath, name, n.buildAttributePath())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

func (n *navigator) at(index int) *navigator {
	if n.err != nil {
		return n
	}

	child := n.current().ByIndex(index)
	if child == nil {
		n.err = fmt.Errorf("%w: no target at index '%d' from '%s'", ErrNoTarget, index, n.buildAttributePath())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

func (n *navigator) where(criteria func(child Property) bool) *navigator {
	if n.err != nil {
		return n
	}

	child := n.current().Find(criteria)
	if child == nil {
		n.err = fmt.Errorf("%w: no target meeting criteria from '%s'", ErrNoTarget, n.buildAttributePath())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

func (n *navigator) forEachChild(callback func(index int, child Property) error) error {
	if n.err != nil {
		return n.err
	}
	return n.current().ForEach(callback)
}

// add delegates for Property.Add
func (n *navigator) add(value any) *navigator {
	if n.err != nil {
		return n
	}
	n.err = n.current().Add(value)
	return n
}

// replace delegates for Property.Set
func (n *navigator) replace(value interface{}) *navigator {
	if n.err != nil {
		return n
	}
	n.err = n.current().Set(value)
	return n
}

// delete delegates for Property.Delete
func (n *navigator) delete() *navigator {
	if n.err != nil {
		return n
	}
	n.current().Delete()
	return n
}

func (n *navigator) buildAttributePath() string {
	var sb strings.Builder
	for i, p := range n.stack {
		if i > 0 {
			sb.WriteRune('.')
		}
		sb.WriteString(p.Attr().name)
	}
	return sb.String()
}
