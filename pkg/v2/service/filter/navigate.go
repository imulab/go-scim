package filter

import (
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// Marker property that indicates flexNavigator is out of sync
var outOfSync = outOfSyncProperty{}

// IsOutOfSync returns true when the given property is the marker out-of-sync property.
func IsOutOfSync(property prop.Property) bool {
	return property == outOfSync
}

// flexNavigator is a Navigator implementation that can be used in two ways.
//
// First, as a follow-along navigator that synchronizes with another Navigator. When the other navigator focuses on
// some child property, caller will try to synchronously focus on the same child property with this navigator by calling
// Dot, At or Where. If synchronization cannot be achieved (i.e. because a child property does not exist), an outOfSync
// marker property is pushed onto the trace stack. Caller needs to check whether Current() == outOfSync to determine if
// the navigator is still in sync. However, since something is always pushed onto the stack, the stack Depth will be
// in sync, hence at some point, the navigator will become in sync again when all the outOfSync property is retracted.
//
// Second, as an active navigator that synchronizes itself with a Visitor. In this use mode, caller is no longer interested
// in Dot, At and Where, because the DFS order is maintained by the Visitor. The caller only needs to Push the visited
// property and Retract when Visit ends.
//
// This navigator does not need an initializing Property to start with and can fully Retract to an empty stack.
type flexNavigator struct {
	stack []prop.Property
	err   error
}

func (n *flexNavigator) Error() error {
	return n.err
}

func (n *flexNavigator) HasError() bool {
	return n.err != nil
}

func (n *flexNavigator) Depth() int {
	return len(n.stack)
}

func (n *flexNavigator) Source() prop.Property {
	return n.stack[0]
}

func (n *flexNavigator) Current() prop.Property {
	return n.stack[len(n.stack)-1]
}

func (n *flexNavigator) Last() prop.Property {
	if len(n.stack) < 2 {
		return nil
	}
	return n.stack[len(n.stack)-2]
}

func (n *flexNavigator) Retract() prop.Navigator {
	if len(n.stack) > 0 {
		n.stack = n.stack[:len(n.stack)-1]
	}
	return n
}

func (n *flexNavigator) Dot(name string) prop.Navigator {
	if IsOutOfSync(n.Current()) {
		n.Push(outOfSync)
		return n
	}

	if n.err != nil {
		return n
	}

	child, err := n.Current().ChildAtIndex(name)
	if err != nil {
		n.Push(outOfSync)
		return n
	}

	n.Push(child)
	return n
}

func (n *flexNavigator) At(index int) prop.Navigator {
	if IsOutOfSync(n.Current()) {
		n.Push(outOfSync)
		return n
	}

	if n.err != nil {
		return n
	}

	child, err := n.Current().ChildAtIndex(index)
	if err != nil {
		n.Push(outOfSync)
		return n
	}

	n.Push(child)
	return n
}

func (n *flexNavigator) Where(criteria func(child prop.Property) bool) prop.Navigator {
	if IsOutOfSync(n.Current()) {
		n.Push(outOfSync)
		return n
	}

	if n.err != nil {
		return n
	}

	child := n.Current().FindChild(criteria)
	if child == nil {
		n.Push(outOfSync)
		return n
	}

	n.Push(child)
	return n
}

func (n *flexNavigator) Add(value interface{}) prop.Navigator {
	n.err = n.delegateMod(func() (event *prop.Event, err error) {
		return n.Current().Add(value)
	})
	return n
}

func (n *flexNavigator) Replace(value interface{}) prop.Navigator {
	n.err = n.delegateMod(func() (event *prop.Event, err error) {
		return n.Current().Replace(value)
	})
	return n
}

func (n *flexNavigator) Delete() prop.Navigator {
	n.err = n.delegateMod(func() (event *prop.Event, err error) {
		return n.Current().Delete()
	})
	return n
}

func (n *flexNavigator) delegateMod(mod func() (*prop.Event, error)) error {
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

func (n *flexNavigator) ForEachChild(callback func(index int, child prop.Property) error) error {
	if n.err != nil {
		return n.err
	}
	return n.Current().ForEachChild(callback)
}

func (n *flexNavigator) Push(p prop.Property) {
	n.stack = append(n.stack, p)
}

// Empty implementation of Property to be pushed onto the stack of flexNavigator to indicate that the
// navigator is currently out of sync. Property methods of this implementation should not be called, as they
// are meaningless.
type outOfSyncProperty struct{}

func (p outOfSyncProperty) Attribute() *spec.Attribute {
	return nil
}

func (p outOfSyncProperty) Raw() interface{} {
	return nil
}

func (p outOfSyncProperty) IsUnassigned() bool {
	return false
}

func (p outOfSyncProperty) Dirty() bool {
	return false
}

func (p outOfSyncProperty) Hash() uint64 {
	return 0
}

func (p outOfSyncProperty) Matches(_ prop.Property) bool {
	return false
}

func (p outOfSyncProperty) Clone() prop.Property {
	return nil
}

func (p outOfSyncProperty) Add(_ interface{}) (*prop.Event, error) {
	return nil, nil
}

func (p outOfSyncProperty) Replace(_ interface{}) (*prop.Event, error) {
	return nil, nil
}

func (p outOfSyncProperty) Delete() (*prop.Event, error) {
	return nil, nil
}

func (p outOfSyncProperty) Notify(_ *prop.Events) error {
	return nil
}

func (p outOfSyncProperty) CountChildren() int {
	return 0
}

func (p outOfSyncProperty) ForEachChild(_ func(index int, child prop.Property) error) error {
	return nil
}

func (p outOfSyncProperty) FindChild(_ func(child prop.Property) bool) prop.Property {
	return nil
}

func (p outOfSyncProperty) ChildAtIndex(_ interface{}) (prop.Property, error) {
	return nil, nil
}
