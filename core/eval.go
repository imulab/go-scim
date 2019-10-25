package core

import (
	"strings"
)

// Interface extension to property so that a property can be evaluated against a filter
type Evaluation interface {
	Property

	// Evaluate this property against the root of the filter tree.
	Evaluate(root *Step) (bool, error)
}

func (c *complexProperty) Evaluate(root *Step) (r bool, err error) {
	if root == nil {
		return false, nil
	}

	switch root.Typ {
	case stepLogicalOperator:
		r, err = c.evaluateLogical(root)
	case stepRelationalOperator:
		r, err = c.evaluateRelational(root)
	default:
		r = false
		err = Errors.InvalidFilter("invalid filter root, cannot be processed")
	}

	return
}

func (c *complexProperty) evaluateLogical(root *Step) (bool, error) {
	switch strings.ToLower(root.Token) {
	case And:
		if l, err := c.Evaluate(root.Left); err != nil {
			return false, err
		} else if r, err := c.Evaluate(root.Right); err != nil {
			return false, err
		} else {
			return l && r, nil
		}

	case Or:
		if l, err := c.Evaluate(root.Left); err != nil {
			return false, err
		} else if l {
			return true, nil
		}

		if r, err := c.Evaluate(root.Right); err != nil {
			return false, err
		} else {
			return r, nil
		}

	case Not:
		if l, err := c.Evaluate(root.Left); err != nil {
			return false, err
		} else {
			return !l, nil
		}

	default:
		panic("not a logical operator")
	}
}

func (c *complexProperty) evaluateRelational(root *Step) (bool, error) {
	var (
		prop Property
		val  interface{}
		err  error
	)
	{
		cursor := root.Left
		prop = c

		for cursor != nil {
			if !cursor.IsPath() {
				return false, Errors.InvalidFilter("path in filter cannot contain other filters")
			}

			prop, err = prop.(*complexProperty).getSubProperty(cursor.Token)
			if err != nil {
				return false, err
			}

			cursor = cursor.Next
		}
	}
	{
		if root.Right != nil {
			val, err = root.Right.compliantValue(prop.Attribute())
			if err != nil {
				return false, err
			}
		}
	}

	switch strings.ToLower(root.Token) {
	case Eq:
		if eqProp, ok := prop.(EqualAware); !ok {
			return false, nil
		} else {
			return eqProp.IsEqualTo(val), nil
		}

	case Ne:
		if eqProp, ok := prop.(EqualAware); !ok {
			return false, nil
		} else {
			return !eqProp.IsEqualTo(val), nil
		}

	case Sw:
		if strProp, ok := prop.(StringOpCapable); !ok {
			return false, nil
		} else {
			return strProp.StartsWith(val), nil
		}

	case Ew:
		if strProp, ok := prop.(StringOpCapable); !ok {
			return false, nil
		} else {
			return strProp.EndsWith(val), nil
		}

	case Co:
		if strProp, ok := prop.(StringOpCapable); !ok {
			return false, nil
		} else {
			return strProp.Contains(val), nil
		}

	case Pr:
		return prop.IsPresent(), nil

	case Lt:
		if orderProp, ok := prop.(OrderAware); !ok {
			return false, nil
		} else {
			return orderProp.IsLessThan(val), nil
		}

	case Le:
		if orderProp, ok := prop.(OrderAware); !ok {
			return false, nil
		} else {
			return orderProp.IsLessThanOrEqualTo(val), nil
		}

	case Gt:
		if orderProp, ok := prop.(OrderAware); !ok {
			return false, nil
		} else {
			return orderProp.IsGreaterThan(val), nil
		}

	case Ge:
		if orderProp, ok := prop.(OrderAware); !ok {
			return false, nil
		} else {
			return orderProp.IsGreaterThanOrEqualTo(val), nil
		}

	default:
		panic("not a relational operator")
	}
}
