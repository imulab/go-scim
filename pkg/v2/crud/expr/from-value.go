package expr

import (
	"fmt"
)

// FromValue returns a filter that represents attribute equality w.r.t. one of
// the list of raw composite value given to it. Each element of the list must
// be of type map[string]interface{}.
func FromValueList(list []interface{}) (*Expression, error) {
	if len(list) == 0 {
		return nil, nil
	}
	var exprs []*Expression
	for _, v := range list {
		value, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid value type for creating expression: %T", v)
		}
		e, err := FromValue(value)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, e)
	}
	// construct an or-ladder of child expressions
	var head, cur *Expression
	for i, e := range exprs {
		if i == len(exprs)-1 {
			if head == nil { // this is all we had.
				head = e
			} else {
				cur.right = e
			}
		} else {
			// at least one more expression to go after this one, so construct an OR node.
			orNode := newOperator(Or)
			orNode.left = e
			if head == nil {
				head = orNode
			} else {
				// cur must be non-nil as well.
				cur.right = orNode
			}
			cur = orNode
		}
	}
	return head, nil
}

// FromValue returns a filter that represents attribute equality w.r.t. the
// raw composite value given to it. Note that no element of the value can be an
// array.
func FromValue(value map[string]interface{}) (*Expression, error) {
	exprs, err := getLeafExpressions("", value)
	if err != nil {
		return nil, err
	}
	// construct an and-ladder of expressions (e.g. (a eq "x") AND ((b.c eq 1) AND (c.d eq false))
	var head, cur *Expression
	for i, e := range exprs {
		if i == len(exprs)-1 {
			if head == nil { // this is all we had.
				head = e
			} else {
				cur.right = e
			}
		} else {
			// at least one more expression to go after this one, so construct an AND node.
			andNode := newOperator(And)
			andNode.left = e
			if head == nil {
				head = andNode
			} else {
				// cur must be non-nil as well.
				cur.right = andNode
			}
			cur = andNode
		}
	}
	return head, nil
}

func getLeafExpressions(pathPrefix string, value interface{}) ([]*Expression, error) {
	switch v := value.(type) {
	// simple types that Property.Raw() could return.
	case string, int64, float64, bool:
		head := newOperator(Eq)
		head.left = newPath(pathPrefix)
		head.right = newLiteral(fmt.Sprintf("\"%v\"", v))
		return []*Expression{head}, nil
	case map[string]interface{}:
		var exprs []*Expression
		for key, value := range v {
			var prefix string
			if pathPrefix != "" {
				prefix = fmt.Sprintf("%s.%s", pathPrefix, key)
			} else {
				prefix = key
			}
			subexprs, err := getLeafExpressions(prefix, value)
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, subexprs...)
		}
		return exprs, nil
	default:
		// cannot handle arrays, anything else is unexpected.
		return nil, fmt.Errorf("unexpected value %v", v)
	}
}
