package expr

import "strings"

const (
	exprPath exprType = iota
	exprLogicalOp
	exprRelationalOp
	exprLiteral
	exprParenthesis
)

type (
	// type of expression
	exprType int
	// An expression is a basic unit in filters and paths. It maintains
	// relation with other expressions as a node in linked list and tree
	// to form the entire filter and/or path.
	Expression struct {
		token string
		typ   exprType
		next  *Expression
		left  *Expression
		right *Expression
	}
)

func (e *Expression) Token() string {
	return e.token
}

func (e *Expression) IsPath() bool {
	return e.typ == exprPath
}

func (e *Expression) IsOperator() bool {
	return e.IsLogicalOperator() || e.IsRelationalOperator()
}

func (e *Expression) IsLogicalOperator() bool {
	return e.typ == exprLogicalOp
}

func (e *Expression) IsRelationalOperator() bool {
	return e.typ == exprRelationalOp
}

func (e *Expression) IsRootOfFilter() bool {
	return e.IsOperator() && e.left != nil
}

func (e *Expression) IsLiteral() bool {
	return e.typ == exprLiteral
}

func (e *Expression) IsParenthesis() bool {
	return e.typ == exprParenthesis
}

func (e *Expression) IsLeftParenthesis() bool {
	return e.typ == exprParenthesis && e.token == LeftParen
}

func (e *Expression) IsRightParenthesis() bool {
	return e.typ == exprParenthesis && e.token == RightParen
}

// Returns true if the remaining of the path whose first node is represented
// by this expression contains a filter.
func (e *Expression) ContainsFilter() bool {
	c := e
	for c != nil {
		if c.IsRootOfFilter() {
			return true
		}
		c = c.next
	}
	return false
}

// Traverse the hybrid linked list / tree structure connected to the current step. cb is the callback function invoked
// for each step; marker and done comprises the termination mechanism. When the current step finishes its traversal, it
// compares itself against marker. If they are equal, invoke the done function to let the caller know we have returned
// to the node that the Walk function is initially invoked on, hence the traversal has ended.
func (e *Expression) Walk(cb func(expression *Expression), marker *Expression, done func()) {
	if e == nil {
		return
	}

	cb(e)
	e.left.Walk(cb, marker, done)
	e.right.Walk(cb, marker, done)
	e.next.Walk(cb, marker, done)

	if marker == e {
		done()
	}
}

func newOperator(op string) *Expression {
	switch strings.ToLower(op) {
	case And, Or, Not:
		return &Expression{
			token: op,
			typ:   exprLogicalOp,
		}
	case Eq, Ne, Sw, Ew, Co, Gt, Ge, Lt, Le, Pr:
		return &Expression{
			token: op,
			typ:   exprRelationalOp,
		}
	default:
		panic("not an operator")
	}
}

func newPath(path string) *Expression {
	return &Expression{
		token: path,
		typ:   exprPath,
	}
}

func newLiteral(value string) *Expression {
	return &Expression{
		token: value,
		typ:   exprLiteral,
	}
}

func newParenthesis(paren string) *Expression {
	return &Expression{
		token: paren,
		typ:   exprParenthesis,
	}
}
