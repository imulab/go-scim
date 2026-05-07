package crud

import (
	"errors"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"strconv"
	"strings"
)

// Evaluate the resource with the given SCIM filter and return the boolean result or an error.
func Evaluate(resource *prop.Resource, filter string) (bool, error) {
	cf, err := expr.CompileFilter(filter)
	if err != nil {
		return false, err
	}
	return evaluator{
		base:   resource.RootProperty(),
		filter: cf,
	}.evaluate()
}

func EvaluateExpressionOnProperty(prop prop.Property, expr *expr.Expression) (bool, error) {
	return evaluator{
		base:   prop,
		filter: expr,
	}.evaluate()
}

type evaluator struct {
	base   prop.Property
	filter *expr.Expression
}

func (v evaluator) evaluate() (bool, error) {
	return v.evalAny(v.base, v.filter)
}

func (v evaluator) evalAny(p prop.Property, op *expr.Expression) (bool, error) {
	switch op.Token() {
	case expr.And:
		return v.evalAnd(p, op)
	case expr.Or:
		return v.evalOr(p, op)
	case expr.Not:
		return v.evalNot(p, op)
	}

	if op.Left().ContainsFilter() {
		return false, fmt.Errorf("%w: nested filter detected", spec.ErrInvalidFilter)
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
	if err := defaultTraverse(p, op.Left(), func(nav prop.Navigator) (fe error) {
		var r bool

		switch op.Token() {
		case expr.Eq:
			r, fe = v.evalEq(nav.Current(), op)
		case expr.Ne:
			r, fe = v.evalNe(nav.Current(), op)
		case expr.Sw:
			r, fe = v.evalSw(nav.Current(), op)
		case expr.Ew:
			r, fe = v.evalEw(nav.Current(), op)
		case expr.Co:
			r, fe = v.evalCo(nav.Current(), op)
		case expr.Gt:
			r, fe = v.evalGt(nav.Current(), op)
		case expr.Ge:
			r, fe = v.evalGe(nav.Current(), op)
		case expr.Lt:
			r, fe = v.evalLt(nav.Current(), op)
		case expr.Le:
			r, fe = v.evalLe(nav.Current(), op)
		case expr.Pr:
			r, fe = v.evalPr(nav.Current())
		default:
			panic("unsupported operator")
		}

		results = append(results, r)
		return
	}); err != nil {
		switch errors.Unwrap(err) {
		case spec.ErrInvalidFilter:
			return false, err
		case spec.ErrInvalidPath, spec.ErrNoTarget:
			return false, fmt.Errorf("%w: bad path in filter", spec.ErrInvalidFilter)
		case spec.ErrInvalidValue:
			return false, fmt.Errorf("%w: bad value in filter", spec.ErrInvalidFilter)
		default:
			return false, fmt.Errorf("%w: failed to evaluate resource", spec.ErrInvalidFilter)
		}
	}

	for _, r := range results {
		if r {
			return true, nil
		}
	}
	return false, nil
}

func (v evaluator) evalEq(target prop.Property, eq *expr.Expression) (bool, error) {
	eqTarget, ok := target.(prop.EqCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), eq.Right().Token())
	if err != nil {
		return false, err
	}

	return eqTarget.EqualsTo(value), nil
}

func (v evaluator) evalNe(target prop.Property, ne *expr.Expression) (bool, error) {
	r, err := v.evalEq(target, ne)
	if err != nil {
		return false, err
	}
	return !r, nil
}

func (v evaluator) evalSw(target prop.Property, sw *expr.Expression) (bool, error) {
	swTarget, ok := target.(prop.SwCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), sw.Right().Token())
	if err != nil {
		return false, err
	}

	if str, ok := value.(string); !ok {
		return false, spec.ErrInvalidValue
	} else {
		return swTarget.StartsWith(str), nil
	}
}

func (v evaluator) evalEw(target prop.Property, ew *expr.Expression) (bool, error) {
	ewTarget, ok := target.(prop.EwCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), ew.Right().Token())
	if err != nil {
		return false, err
	}

	if str, ok := value.(string); !ok {
		return false, spec.ErrInvalidValue
	} else {
		return ewTarget.EndsWith(str), nil
	}
}

func (v evaluator) evalCo(target prop.Property, co *expr.Expression) (bool, error) {
	coTarget, ok := target.(prop.CoCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), co.Right().Token())
	if err != nil {
		return false, err
	}

	if str, ok := value.(string); !ok {
		return false, spec.ErrInvalidValue
	} else {
		return coTarget.Contains(str), nil
	}
}

func (v evaluator) evalGt(target prop.Property, gt *expr.Expression) (bool, error) {
	gtTarget, ok := target.(prop.GtCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), gt.Right().Token())
	if err != nil {
		return false, err
	}

	return gtTarget.GreaterThan(value), nil
}

func (v evaluator) evalGe(target prop.Property, ge *expr.Expression) (bool, error) {
	geTarget, ok := target.(prop.GeCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), ge.Right().Token())
	if err != nil {
		return false, err
	}

	return geTarget.GreaterThanOrEqualTo(value), nil
}

func (v evaluator) evalLt(target prop.Property, lt *expr.Expression) (bool, error) {
	ltTarget, ok := target.(prop.LtCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), lt.Right().Token())
	if err != nil {
		return false, err
	}

	return ltTarget.LessThan(value), nil
}

func (v evaluator) evalLe(target prop.Property, le *expr.Expression) (bool, error) {
	leTarget, ok := target.(prop.LeCapable)
	if !ok {
		return false, nil
	}

	value, err := v.normalize(target.Attribute(), le.Right().Token())
	if err != nil {
		return false, err
	}

	return leTarget.LessThanOrEqualTo(value), nil
}

func (v evaluator) evalPr(target prop.Property) (bool, error) {
	prTarget, ok := target.(prop.PrCapable)
	if !ok {
		return false, nil
	}

	return prTarget.Present(), nil
}

func (v evaluator) evalAnd(p prop.Property, and *expr.Expression) (bool, error) {
	if left, err := v.evalAny(p, and.Left()); err != nil {
		return false, err
	} else if !left {
		return false, nil
	} else {
		return v.evalAny(p, and.Right())
	}
}

func (v evaluator) evalOr(p prop.Property, or *expr.Expression) (bool, error) {
	if left, err := v.evalAny(p, or.Left()); err != nil {
		return false, err
	} else if left {
		return true, nil
	} else {
		return v.evalAny(p, or.Right())
	}
}

func (v evaluator) evalNot(p prop.Property, not *expr.Expression) (bool, error) {
	if left, err := v.evalAny(p, not.Left()); err != nil {
		return false, err
	} else {
		return !left, nil
	}
}

// Take the raw string presentation of a value and normalize it to corresponding types according to the attribute.
func (v evaluator) normalize(attr *spec.Attribute, token string) (interface{}, error) {
	switch attr.Type() {
	case spec.TypeString, spec.TypeDateTime, spec.TypeBinary, spec.TypeReference:
		if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
			token = strings.TrimPrefix(token, "\"")
			token = strings.TrimSuffix(token, "\"")
			return token, nil
		} else {
			return nil, spec.ErrInvalidValue
		}
	case spec.TypeInteger:
		if i64, err := strconv.ParseInt(token, 10, 64); err != nil {
			return nil, spec.ErrInvalidValue
		} else {
			return i64, nil
		}
	case spec.TypeDecimal:
		if f64, err := strconv.ParseFloat(token, 64); err != nil {
			return nil, spec.ErrInvalidValue
		} else {
			return f64, nil
		}
	case spec.TypeBoolean:
		if b, err := strconv.ParseBool(token); err != nil {
			return nil, spec.ErrInvalidValue
		} else {
			return b, nil
		}
	default:
		return nil, spec.ErrInvalidValue
	}
}
