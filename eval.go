package scim

import (
	"errors"
	"fmt"
	"strings"
)

func evaluate(p Property, filter string) (bool, error) {
	cf, err := compileFilter(filter)
	if err != nil {
		return false, err
	}

	v := &evaluator{
		root:   p,
		filter: cf,
	}

	return v.evaluate()
}

type evaluator struct {
	root   Property
	filter *Expr
}

func (v *evaluator) evaluate() (bool, error) {
	return v.do(v.root, v.filter)
}

func (v *evaluator) do(p Property, op *Expr) (bool, error) {
	switch strings.ToLower(op.value) {
	case And:
		return v.and(p, op)
	case Or:
		return v.or(p, op)
	case Not:
		return v.not(p, op)
	}

	if op.left.containsFilter() {
		return false, fmt.Errorf("%w: nested filter detected", ErrInvalidFilter)
	}

	// Normally, we are expecting a single boolean result. For instance, conventional filters like
	//
	//		userName eq "imulab"
	// 		name.givenName co "W"
	//
	// produces a single boolean result. These filters lead to only one comparison.
	//
	// However, the path of a filter may split the traversal when it visits a multiValued property, leading
	// to many comparison, producing multiple boolean results. In this case, we need to collect all boolean
	// results and return true as long as one of the result was true. For instance, given partial resource
	//
	//		{
	//			"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
	//			"id": "C6AE8285-59C0-4E13-9C44-CE50C3F63DDC",
	//			"emails": [
	//				{
	//					"value": "user1@foo.com",
	//					"primary": true
	//				},
	//				{
	//					"value": "user2@foo.com"
	//				}
	//			]
	//		}
	//
	// 		and a filter: emails.value sw "user1"
	//
	// This filter leads to two comparisons of "user1@foo.com" sw "user1", and "user2@foo.com" sw "user1" respectively,
	// which produces "true" and "false". As a result, this resource should pass the filter.
	var results = make([]bool, 0)
	if err := defaultTraverse(p, op.left, func(nav *navigator) (fe error) {
		var r bool

		switch strings.ToLower(op.value) {
		case Eq:
			r, fe = v.eq(nav.current(), op)
		case Ne:
			r, fe = v.ne(nav.current(), op)
		case Sw:
			r, fe = v.sw(nav.current(), op)
		case Ew:
			r, fe = v.ew(nav.current(), op)
		case Co:
			r, fe = v.co(nav.current(), op)
		case Gt:
			r, fe = v.gt(nav.current(), op)
		case Ge:
			r, fe = v.ge(nav.current(), op)
		case Lt:
			r, fe = v.lt(nav.current(), op)
		case Le:
			r, fe = v.le(nav.current(), op)
		case Pr:
			r, fe = v.pr(nav.current())
		default:
			panic("unsupported operator")
		}

		results = append(results, r)
		return
	}); err != nil {
		switch errors.Unwrap(err) {
		case ErrInvalidFilter:
			return false, err
		case ErrInvalidPath, ErrNoTarget:
			return false, fmt.Errorf("%w: bad path in filter", ErrInvalidFilter)
		case ErrInvalidValue:
			return false, fmt.Errorf("%w: bad value in filter", ErrInvalidFilter)
		default:
			return false, fmt.Errorf("%w: failed to evaluate resource", ErrInvalidFilter)
		}
	}

	for _, r := range results {
		if r {
			return true, nil
		}
	}

	return false, nil
}

func (v *evaluator) eq(target Property, eq *Expr) (bool, error) {
	eqTarget, ok := target.(eqTrait)
	if !ok {
		return false, nil
	}

	value, err := target.Attr().parseValue(eq.right.value)
	if err != nil {
		return false, err
	}

	return eqTarget.equalsTo(value), nil
}

func (v *evaluator) ne(target Property, ne *Expr) (bool, error) {
	r, err := v.ne(target, ne)
	if err != nil {
		return false, err
	}
	return !r, nil
}

func (v *evaluator) sw(target Property, sw *Expr) (bool, error) {
	swTarget, ok := target.(swTrait)
	if !ok {
		return false, nil
	}

	value, err := target.Attr().parseValue(sw.right.value)
	if err != nil {
		return false, err
	}

	if str, ok := value.(string); !ok {
		return false, ErrInvalidValue
	} else {
		return swTarget.startsWith(str), nil
	}
}

func (v *evaluator) ew(target Property, ew *Expr) (bool, error) {
	ewTarget, ok := target.(ewTrait)
	if !ok {
		return false, nil
	}

	value, err := target.Attr().parseValue(ew.right.value)
	if err != nil {
		return false, err
	}

	if str, ok := value.(string); !ok {
		return false, ErrInvalidValue
	} else {
		return ewTarget.endsWith(str), nil
	}
}

func (v *evaluator) co(target Property, co *Expr) (bool, error) {
	coTarget, ok := target.(coTrait)
	if !ok {
		return false, nil
	}

	value, err := target.Attr().parseValue(co.right.value)
	if err != nil {
		return false, err
	}

	if str, ok := value.(string); !ok {
		return false, ErrInvalidValue
	} else {
		return coTarget.contains(str), nil
	}
}

func (v *evaluator) gt(target Property, gt *Expr) (bool, error) {
	gtTarget, ok := target.(gtTrait)
	if !ok {
		return false, nil
	}

	value, err := target.Attr().parseValue(gt.right.value)
	if err != nil {
		return false, err
	}

	return gtTarget.greaterThan(value), nil
}

func (v *evaluator) ge(target Property, ge *Expr) (bool, error) {
	if r, err := v.gt(target, ge); err == nil && r {
		return true, nil
	}

	return v.eq(target, ge)
}

func (v *evaluator) lt(target Property, lt *Expr) (bool, error) {
	ltTarget, ok := target.(ltTrait)
	if !ok {
		return false, nil
	}

	value, err := target.Attr().parseValue(lt.right.value)
	if err != nil {
		return false, err
	}

	return ltTarget.lessThan(value), nil
}

func (v *evaluator) le(target Property, le *Expr) (bool, error) {
	if r, err := v.lt(target, le); err == nil && r {
		return true, nil
	}

	return v.eq(target, le)
}

func (v *evaluator) pr(target Property) (bool, error) {
	prTarget, ok := target.(prTrait)
	if !ok {
		return false, nil
	}

	return prTarget.isPresent(), nil
}

func (v *evaluator) and(p Property, and *Expr) (bool, error) {
	if left, err := v.do(p, and.left); err != nil {
		return false, err
	} else if !left {
		return false, nil
	} else {
		return v.do(p, and.right)
	}
}

func (v *evaluator) or(p Property, or *Expr) (bool, error) {
	if left, err := v.do(p, or.left); err != nil {
		return false, err
	} else if left {
		return true, nil
	} else {
		return v.do(p, or.right)
	}
}

func (v *evaluator) not(p Property, not *Expr) (bool, error) {
	if left, err := v.do(p, not.left); err != nil {
		return false, err
	} else {
		return !left, nil
	}
}
