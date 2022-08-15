package scim

import "fmt"

func defaultTraverse(property Property, query *Expr, callback func(nav *navigator) error) error {
	t := &traverser{
		nav:             newNavigator(property),
		fn:              callback,
		elementStrategy: selectAllChild,
	}
	return t.traverse(query)
}

func primaryOrFirstTraverse(property Property, query *Expr, callback func(nav *navigator) error) error {
	t := &traverser{
		nav:             newNavigator(property),
		fn:              callback,
		elementStrategy: primaryOrFirstChild,
	}
	return t.traverse(query)
}

type traverser struct {
	nav             *navigator
	fn              func(nav *navigator) error
	elementStrategy elementStrategy
}

func (t *traverser) traverse(query *Expr) error {
	if query == nil {
		return t.fn(t.nav)
	}

	if query.IsFilterRoot() {
		if !t.nav.current().Attr().multiValued {
			return fmt.Errorf("%w: filter applied to singular attribute", ErrInvalidFilter)
		}
		return t.traverseQualifiedElements(query)
	}

	if t.nav.current().Attr().multiValued {
		return t.traverseSelectedElements(query)
	}

	return t.traverseNext(query)
}

func (t *traverser) traverseNext(path *Expr) error {
	t.nav.dot(path.value)
	if t.nav.hasError() {
		return t.nav.err
	}

	defer t.nav.retract()

	return t.traverse(path.next)
}

func (t *traverser) traverseSelectedElements(query *Expr) error {
	strategy := t.elementStrategy(t.nav.current())

	return t.nav.current().ForEach(func(index int, child Property) error {
		if !strategy(index, child) {
			return nil
		}

		t.nav.at(index)
		if t.nav.hasError() {
			return t.nav.err
		}

		defer t.nav.retract()

		return t.traverse(query)
	})
}

func (t *traverser) traverseQualifiedElements(filter *Expr) error {
	panic("TODO")
}

// elementStrategy is a high-level function that returns a function to determine whether a child property should be traversed
type elementStrategy func(container Property) func(index int, child Property) bool

// selectAllChild implements elementStrategy to visit all children.
func selectAllChild(_ Property) func(index int, child Property) bool {
	return func(index int, child Property) bool {
		return true
	}
}

// primaryOrFirstChild implements elementStrategy to visit the primary child property that is true or the first child
// property in case the former does not exist.
func primaryOrFirstChild(container Property) func(index int, child Property) bool {
	truePrimaryChild := container.Find(func(child Property) bool {
		return child.Attr().primary && child.Value() == true
	})

	if truePrimaryChild != nil {
		return func(index int, child Property) bool {
			return child == truePrimaryChild
		}
	}

	return func(index int, child Property) bool {
		return index == 0
	}
}
