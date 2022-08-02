package core

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrPath   = errors.New("path error")
	ErrTarget = errors.New("target error")
)

// Navigator is a controlled mechanism to traverse the Property tree. It should be used in cases where the caller has
// knowledge of what to access. For example, when de-serializing JSON into a Property, caller has knowledge of the JSON
// structure, therefore knows what to access in the Property structure.
//
// Navigator does not return error on every error-able operation, implementation will mostly be stateful. Callers
// should call Error or HasError to check if the Navigator is currently in the error state, after performing one or
// possibly several chained operations fluently.
//
// The navigation stack can be advanced by calling Dot, At and Where methods. Which methods to use depends on the context.
// The stack can be retracted by calling Retract. The top and the bottom item on the stack can be queried by calling
// Current and Source. The stack depth is available via Depth.
type Navigator struct {
	stack []Property
	err   error
}

func (n *Navigator) Error() error {
	return n.err
}

func (n *Navigator) HasError() bool {
	return n.err != nil
}

func (n *Navigator) ClearError() {
	n.err = nil
}

func (n *Navigator) Depth() int {
	return len(n.stack)
}

func (n *Navigator) Source() Property {
	return n.stack[0]
}

func (n *Navigator) Current() Property {
	return n.stack[len(n.stack)-1]
}

func (n *Navigator) Retract() *Navigator {
	if n.Depth() > 1 {
		n.stack = n.stack[:len(n.stack)-1]
	}
	return n
}

func (n *Navigator) Dot(name string) *Navigator {
	if n.err != nil {
		return n
	}

	child := n.Current().ByIndex(name)
	if child == nil {
		n.err = fmt.Errorf("%w: no attribute named '%s' from '%s'", ErrPath, name, n.buildAttributePath())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

func (n *Navigator) At(index int) *Navigator {
	if n.err != nil {
		return n
	}

	child := n.Current().ByIndex(index)
	if child == nil {
		n.err = fmt.Errorf("%w: no target at index '%d' from '%s'", ErrTarget, index, n.buildAttributePath())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

func (n *Navigator) Where(criteria func(child Property) bool) *Navigator {
	if n.err != nil {
		return n
	}

	child := n.Current().Find(criteria)
	if child == nil {
		n.err = fmt.Errorf("%w: no target meeting criteria from '%s'", ErrTarget, n.buildAttributePath())
		return n
	}

	n.stack = append(n.stack, child)
	return n
}

func (n *Navigator) ForEachChild(callback func(index int, child Property) error) error {
	if n.err != nil {
		return n.err
	}
	return n.Current().ForEach(callback)
}

// Add delegates for Property.Add
func (n *Navigator) Add(value any) *Navigator {
	if n.err != nil {
		return n
	}
	n.err = n.Current().Add(value)
	return n
}

// Replace delegates for Property.Set
func (n *Navigator) Replace(value interface{}) *Navigator {
	if n.err != nil {
		return n
	}
	n.err = n.Current().Set(value)
	return n
}

// Delete delegates for Property.Delete
func (n *Navigator) Delete() *Navigator {
	if n.err != nil {
		return n
	}
	n.Current().Delete()
	return n
}

func (n *Navigator) buildAttributePath() string {
	var sb strings.Builder
	for i, p := range n.stack {
		if i > 0 {
			sb.WriteRune('.')
		}
		sb.WriteString(p.Attr().name)
	}
	return sb.String()
}
