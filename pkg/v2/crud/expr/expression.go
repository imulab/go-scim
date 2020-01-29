package expr

import "strings"

const (
	path exprType = iota
	logicalOp
	relationalOp
	literal
	parenthesis
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

func (e *Expression) Next() *Expression {
	return e.next
}

func (e *Expression) Left() *Expression {
	return e.left
}

func (e *Expression) Right() *Expression {
	return e.right
}

func (e *Expression) IsPath() bool {
	return e.typ == path
}

func (e *Expression) IsOperator() bool {
	return e.IsLogicalOperator() || e.IsRelationalOperator()
}

func (e *Expression) IsLogicalOperator() bool {
	return e.typ == logicalOp
}

func (e *Expression) IsRelationalOperator() bool {
	return e.typ == relationalOp
}

func (e *Expression) IsRootOfFilter() bool {
	return e.IsOperator() && e.left != nil
}

func (e *Expression) IsLiteral() bool {
	return e.typ == literal
}

func (e *Expression) IsParenthesis() bool {
	return e.typ == parenthesis
}

func (e *Expression) IsLeftParenthesis() bool {
	return e.typ == parenthesis && e.token == LeftParen
}

func (e *Expression) IsRightParenthesis() bool {
	return e.typ == parenthesis && e.token == RightParen
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

// Walk traverses the hybrid linked list / tree structure connected to the current step. cb is the callback function invoked
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
			typ:   logicalOp,
		}
	case Eq, Ne, Sw, Ew, Co, Gt, Ge, Lt, Le, Pr:
		return &Expression{
			token: op,
			typ:   relationalOp,
		}
	default:
		panic("not an operator")
	}
}

func newPath(pathName string) *Expression {
	return &Expression{
		token: pathName,
		typ:   path,
	}
}

func newLiteral(value string) *Expression {
	return &Expression{
		token: value,
		typ:   literal,
	}
}

func newParenthesis(paren string) *Expression {
	return &Expression{
		token: paren,
		typ:   parenthesis,
	}
}
