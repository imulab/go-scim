package crud

import (
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/expr"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"strconv"
	"strings"
)

// Evaluate if the property meets the compiled SCIM filter.
func evaluate(property prop.Property, filter *expr.Expression) (bool, error) {
	if filter == nil {
		return false, errors.InvalidFilter("filter is invalid")
	}

	switch filter.Token() {
	case expr.And:
		if left, err := evaluate(property, filter.Left()); err != nil {
			return false, err
		} else if !left {
			return false, nil
		} else {
			return evaluate(property, filter.Right())
		}
	case expr.Or:
		if left, err := evaluate(property, filter.Left()); err != nil {
			return false, err
		} else if left {
			return true, nil
		} else {
			return evaluate(property, filter.Right())
		}
	case expr.Not:
		if left, err := evaluate(property, filter.Left()); err != nil {
			return false, err
		} else {
			return !left, nil
		}
	}

	if filter.Left().ContainsFilter() {
		return false, errors.InvalidFilter("nested filter is not allowed")
	}

	var (
		result bool
	)
	if err := traverse(prop.NewNavigator(property), filter.Left(), func(target prop.Property) (fe error) {
		var value interface{}
		if filter.Token() != expr.Pr {
			value, fe = normalize(target.Attribute(), filter.Right().Token())
			if fe != nil {
				return
			}
		}

		switch filter.Token() {
		case expr.Eq:
			result, fe = target.EqualsTo(value)
		case expr.Ne:
			if eq, e := target.EqualsTo(value); e != nil {
				fe = e
			} else {
				result = !eq
			}
		case expr.Sw:
			if str, ok := value.(string); !ok {
				fe = errors.InvalidFilter("sw operator cannot be applied to non-string value")
			} else {
				result, fe = target.StartsWith(str)
			}
		case expr.Ew:
			if str, ok := value.(string); !ok {
				fe = errors.InvalidFilter("ew operator cannot be applied to non-string value")
			} else {
				result, fe = target.EndsWith(str)
			}
		case expr.Co:
			if str, ok := value.(string); !ok {
				fe = errors.InvalidFilter("co operator cannot be applied to non-string value")
			} else {
				result, fe = target.Contains(str)
			}
		case expr.Gt:
			result, fe = target.GreaterThan(value)
		case expr.Ge:
			if gt, e := target.GreaterThan(value); e != nil {
				fe = e
			} else if gt {
				result = true
			} else {
				result, fe = target.EqualsTo(value)
			}
		case expr.Lt:
			result, fe = target.LessThan(value)
		case expr.Le:
			if lt, e := target.LessThan(value); e != nil {
				fe = e
			} else if lt {
				result = true
			} else {
				result, fe = target.EqualsTo(value)
			}
		case expr.Pr:
			result = target.Present()
		default:
			panic("invalid operator")
		}
		return
	}); err != nil {
		return false, errors.InvalidFilter(err.Error())
	}

	return result, nil
}

// Take the raw string presentation of a value and normalize it to corresponding types according to the attribute.
func normalize(attr *spec.Attribute, token string) (interface{}, error) {
	switch attr.Type() {
	case spec.TypeString, spec.TypeDateTime, spec.TypeBinary, spec.TypeReference:
		if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
			token = strings.TrimPrefix(token, "\"")
			token = strings.TrimSuffix(token, "\"")
			return token, nil
		} else {
			return nil, errors.InvalidFilter("'%s' expects string value, but value was unquoted", attr.Path())
		}
	case spec.TypeInteger:
		if i64, err := strconv.ParseInt(token, 10, 64); err != nil {
			return nil, errors.InvalidFilter("'%s' expects integer value", attr.Path())
		} else {
			return i64, nil
		}
	case spec.TypeDecimal:
		if f64, err := strconv.ParseFloat(token, 64); err != nil {
			return nil, errors.InvalidFilter("'%s' expects decimal value", attr.Path())
		} else {
			return f64, nil
		}
	case spec.TypeBoolean:
		if b, err := strconv.ParseBool(token); err != nil {
			return nil, errors.InvalidFilter("'%s' expects boolean value", attr.Path())
		} else {
			return b, nil
		}
	default:
		return nil, errors.InvalidFilter("'%s' cannot be directly compared to any value", attr.Path())
	}
}
