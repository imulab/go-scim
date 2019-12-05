package crud

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/expr"
	"github.com/imulab/go-scim/src/core/prop"
)

// Walk down the currently focused property in the navigator, following the current node in the query expression,
// and eventually invoke callback on the property corresponding to the end of the query.
func traverse(nav *prop.Navigator, query *expr.Expression, callback func(target core.Property) error) error {
	if query == nil {
		return callback(nav.Current())
	}

	if query.IsRootOfFilter() {
		if nav.Current().Attribute().SingleValued() {
			return errors.InvalidFilter("filter cannot be applied to singular properties")
		}

		return nav.Current().(core.Container).ForEachChild(func(_ int, child core.Property) error {
			if r, e := evaluate(child, query); e != nil {
				return e
			} else if r {
				return traverse(prop.NewNavigator(child), query.Next(), callback)
			} else {
				return nil
			}
		})
	}

	if nav.Current().Attribute().MultiValued() {
		return nav.Current().(core.Container).ForEachChild(func(_ int, child core.Property) error {
			childNav := prop.NewNavigator(child)
			if _, err := childNav.FocusName(query.Token()); err != nil {
				return err
			} else {
				return traverse(childNav, query.Next(), callback)
			}
		})
	} else {
		if _, err := nav.FocusName(query.Token()); err != nil {
			return err
		} else {
			return traverse(nav, query.Next(), callback)
		}
	}
}
