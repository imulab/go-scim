package shared

import (
	"reflect"
	"strings"
)

func newPredicate(filter FilterNode, attrSource AttributeSource) predicate {
	return &predicateImpl{filter, attrSource}
}

type predicate interface {
	evaluate(Complex) bool
}

type predicateFunc func(Complex) bool

type predicateImpl struct {
	filter     FilterNode
	attrSource AttributeSource
}

func (impl *predicateImpl) evaluate(c Complex) bool {
	return impl.getFunc(impl.filter)(c)
}

func (impl *predicateImpl) getFunc(filter FilterNode) predicateFunc {
	switch filter.Data() {
	case And:
		return impl.andFunc(filter)
	case Or:
		return impl.orFunc(filter)
	case Not:
		return impl.notFunc(filter)
	case Eq:
		return impl.eqFunc(filter)
	case Ne:
		return impl.neFunc(filter)
	case Sw:
		return impl.swFunc(filter)
	case Ew:
		return impl.ewFunc(filter)
	case Co:
		return impl.coFunc(filter)
	case Pr:
		return impl.prFunc(filter)
	case Gt:
		return impl.gtFunc(filter)
	case Ge:
		return impl.geFunc(filter)
	case Lt:
		return impl.ltFunc(filter)
	case Le:
		return impl.leFunc(filter)
	}
	return nil
}

func (impl *predicateImpl) andFunc(filter FilterNode) predicateFunc {
	lhs := impl.getFunc(filter.Left())
	rhs := impl.getFunc(filter.Right())
	return func(c Complex) bool {
		return lhs(c) && rhs(c)
	}
}

func (impl *predicateImpl) orFunc(filter FilterNode) predicateFunc {
	lhs := impl.getFunc(filter.Left())
	rhs := impl.getFunc(filter.Right())
	return func(c Complex) bool {
		return lhs(c) || rhs(c)
	}
}

func (impl *predicateImpl) notFunc(filter FilterNode) predicateFunc {
	lhs := impl.getFunc(filter.Left())
	return func(c Complex) bool {
		return !lhs(c)
	}
}

func (impl *predicateImpl) eqFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		return impl.compare(filter.Left(), filter.Right(), c) == equal
	}
}

func (impl *predicateImpl) neFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		return impl.compare(filter.Left(), filter.Right(), c) != equal
	}
}

func (impl *predicateImpl) gtFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		return impl.compare(filter.Left(), filter.Right(), c) == greater
	}
}

func (impl *predicateImpl) geFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		r := impl.compare(filter.Left(), filter.Right(), c)
		return r == greater || r == equal
	}
}

func (impl *predicateImpl) ltFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		return impl.compare(filter.Left(), filter.Right(), c) == less
	}
}

func (impl *predicateImpl) leFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		r := impl.compare(filter.Left(), filter.Right(), c)
		return r == less || r == equal
	}
}

func (impl *predicateImpl) prFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		if filter.Left().Type() != PathOperand {
			return false
		}

		key := filter.Left().Data().(Path)
		lVal := reflect.ValueOf(<-c.Get(key, impl.attrSource))
		if !lVal.IsValid() {
			return false
		}

		if lVal.Kind() == reflect.Interface {
			lVal = lVal.Elem()
		}

		switch lVal.Kind() {
		case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
			return lVal.Len() > 0
		default:
			return true
		}
	}
}

func (impl *predicateImpl) swFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		return impl.stringOp(filter.Left(), filter.Right(), c, func(a, b string) bool {
			return strings.HasPrefix(a, b)
		})
	}
}

func (impl *predicateImpl) ewFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		return impl.stringOp(filter.Left(), filter.Right(), c, func(a, b string) bool {
			return strings.HasSuffix(a, b)
		})
	}
}

func (impl *predicateImpl) coFunc(filter FilterNode) predicateFunc {
	return func(c Complex) bool {
		return impl.stringOp(filter.Left(), filter.Right(), c, func(a, b string) bool {
			return strings.Contains(a, b)
		})
	}
}

func (impl *predicateImpl) stringOp(lhs, rhs FilterNode, c Complex, op func(a, b string) bool) bool {
	if lhs.Type() != PathOperand || rhs.Type() != ConstantOperand {
		return false
	}

	key := lhs.Data().(Path)
	attr := impl.attrSource.GetAttribute(key, false)
	if attr == nil || attr.MultiValued || attr.Type == TypeComplex {
		return false
	}

	lVal := reflect.ValueOf(<-c.Get(key, impl.attrSource))
	if !lVal.IsValid() {
		return false
	} else if lVal.Kind() == reflect.Interface {
		lVal = lVal.Elem()
	}

	rVal := reflect.ValueOf(rhs.Data())
	if !rVal.IsValid() {
		return false
	} else if rVal.Kind() == reflect.Interface {
		lVal = lVal.Elem()
	}

	if !impl.kindOf(lVal, reflect.String) || !impl.kindOf(rVal, reflect.String) {
		return false
	} else {
		if attr.CaseExact {
			return op(lVal.String(), rVal.String())
		} else {
			return op(strings.ToLower(lVal.String()), strings.ToLower(rVal.String()))
		}
	}
}

func (impl *predicateImpl) compare(lhs, rhs FilterNode, c Complex) comparison {
	if lhs.Type() != PathOperand || rhs.Type() != ConstantOperand {
		return invalid
	}

	key := lhs.Data().(Path)
	attr := impl.attrSource.GetAttribute(key, true)
	if attr == nil || attr.MultiValued || attr.Type == TypeComplex {
		return invalid
	}

	lVal := reflect.ValueOf(<-c.Get(key, impl.attrSource))
	if !lVal.IsValid() {
		return invalid
	} else if lVal.Kind() == reflect.Interface {
		lVal = lVal.Elem()
	}

	rVal := reflect.ValueOf(rhs.Data())
	if !rVal.IsValid() {
		return invalid
	} else if rVal.Kind() == reflect.Interface {
		lVal = lVal.Elem()
	}

	switch attr.Type {
	case TypeInteger:
		if !impl.kindOf(lVal, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64) ||
			!impl.kindOf(rVal, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64) {
			return invalid
		} else {
			switch {
			case lVal.Int() == rVal.Int():
				return equal
			case lVal.Int() < rVal.Int():
				return less
			case lVal.Int() > rVal.Int():
				return greater
			}
		}

	case TypeDecimal:
		if !impl.kindOf(lVal, reflect.Float32, reflect.Float64) || !impl.kindOf(rVal, reflect.Float32, reflect.Float64) {
			return invalid
		} else {
			switch {
			case lVal.Float() == rVal.Float():
				return equal
			case lVal.Float() < rVal.Float():
				return less
			case lVal.Float() > rVal.Float():
				return greater
			}
		}

	case TypeBoolean:
		if !impl.kindOf(lVal, reflect.Bool) || !impl.kindOf(rVal, reflect.Bool) {
			return invalid
		} else {
			if lVal.Bool() == rVal.Bool() {
				return equal
			} else {
				return invalid
			}
		}

	case TypeString, TypeBinary, TypeDateTime, TypeReference:
		if !impl.kindOf(lVal, reflect.String) || !impl.kindOf(rVal, reflect.String) {
			return invalid
		} else {
			var a, b string
			if attr.CaseExact {
				a, b = lVal.String(), rVal.String()
			} else {
				a, b = strings.ToLower(lVal.String()), strings.ToLower(rVal.String())
			}
			switch {
			case a == b:
				return equal
			case a < b:
				return less
			case a > b:
				return greater
			}
		}
	}

	return invalid
}

func (impl *predicateImpl) kindOf(v reflect.Value, kinds ...reflect.Kind) bool {
	for _, kind := range kinds {
		if kind == v.Kind() {
			return true
		}
	}
	return false
}

type comparison int

const (
	invalid = comparison(-2)
	less    = comparison(-1)
	equal   = comparison(0)
	greater = comparison(1)
)
