package crud

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

func defaultTraverse(property prop.Property, query *expr.Expression, callback func(nav prop.Navigator) error) error {
	return traverser{
		nav:             prop.Navigate(property),
		callback:        callback,
		elementStrategy: selectAllStrategy,
	}.traverse(query)
}

func primaryOrFirstTraverse(property prop.Property, query *expr.Expression, callback func(nav prop.Navigator) error) error {
	return traverser{
		nav:             prop.Navigate(property),
		callback:        callback,
		elementStrategy: primaryOrFirstStrategy,
	}.traverse(query)
}

type traverser struct {
	nav             prop.Navigator                 // stateful navigator for the resource being traversed
	callback        func(nav prop.Navigator) error // callback function to be invoked when target is reached
	elementStrategy elementStrategy                // strategy to select element properties to traverse for multiValued properties
}

func (t traverser) traverse(query *expr.Expression) error {
	if query == nil {
		return t.callback(t.nav)
	}

	if query.IsRootOfFilter() {
		if !t.nav.Current().Attribute().MultiValued() {
			return fmt.Errorf("%w: filter applied to singular attribute", spec.ErrInvalidFilter)
		}
		return t.traverseQualifiedElements(query)
	}

	if t.nav.Current().Attribute().MultiValued() {
		return t.traverseSelectedElements(query)
	}

	return t.traverseNext(query)
}

func (t traverser) traverseNext(query *expr.Expression) error {
	t.nav.Dot(query.Token())
	if err := t.nav.Error(); err != nil {
		return err
	}
	defer t.nav.Retract()

	return t.traverse(query.Next())
}

func (t traverser) traverseSelectedElements(query *expr.Expression) error {
	selector := t.elementStrategy(t.nav.Current())

	return t.nav.Current().ForEachChild(func(index int, child prop.Property) error {
		if !selector(index, child) { // skip elements not satisfied by strategy
			return nil
		}

		t.nav.At(index)
		if err := t.nav.Error(); err != nil {
			return err
		}
		defer t.nav.Retract()

		return t.traverse(query)
	})
}

func (t traverser) traverseQualifiedElements(filter *expr.Expression) error {
	return t.nav.ForEachChild(func(index int, child prop.Property) error {
		t.nav.At(index)
		if err := t.nav.Error(); err != nil {
			return err
		}
		defer t.nav.Retract()

		r, err := evaluator{base: t.nav.Current(), filter: filter}.evaluate()
		if err != nil {
			return err
		} else if !r {
			return nil
		}

		return t.traverse(filter.Next())
	})
}

type elementStrategy func(multiValuedComplex prop.Property) func(index int, child prop.Property) bool

var (
	// strategy to traverse all children elements
	selectAllStrategy elementStrategy = func(multiValuedComplex prop.Property) func(index int, child prop.Property) bool {
		return func(index int, child prop.Property) bool {
			return true
		}
	}
	// strategy to traverse the element whose primary attribute is true, or the first element when no primary attribute is true
	primaryOrFirstStrategy elementStrategy = func(multiValuedComplex prop.Property) func(index int, child prop.Property) bool {
		primaryAttr := multiValuedComplex.Attribute().FindSubAttribute(func(subAttr *spec.Attribute) bool {
			_, ok := subAttr.Annotation(annotation.Primary)
			return ok && subAttr.Type() == spec.TypeBoolean
		})

		if primaryAttr != nil {
			truePrimary := multiValuedComplex.FindChild(func(child prop.Property) bool {
				p, err := child.ChildAtIndex(primaryAttr.Name())
				return err == nil && p != nil && p.Raw() == true
			})
			if truePrimary != nil {
				return func(index int, child prop.Property) bool {
					return child == truePrimary
				}
			}
		}

		return func(index int, child prop.Property) bool {
			return index == 0
		}
	}
)
