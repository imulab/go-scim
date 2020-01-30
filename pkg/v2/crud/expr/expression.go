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
	// Expression is the basic data structure that composes SCIM filters and SCIM paths. It doubles as a node in a
	// single linked list when acting as a segment in SCIM paths, and a node in a binary tree when acting as a token
	// in SCIM filters.
	Expression struct {
		token string
		typ   exprType
		next  *Expression
		left  *Expression
		right *Expression
	}
)

// Token returns the string representation of this Expression.
func (e *Expression) Token() string {
	return e.token
}

// Next returns the next Expression in the linked list, or nil if this Expression is the tail.
func (e *Expression) Next() *Expression {
	return e.next
}

// Left returns the left child in the tree, or nil if this Expression does not have left child.
func (e *Expression) Left() *Expression {
	return e.left
}

// Right returns the right child in the tree, or nil if this Expression does not have right child.
func (e *Expression) Right() *Expression {
	return e.right
}

// IsPath returns tree if this Expression represents a segment of a SCIM path.
func (e *Expression) IsPath() bool {
	return e.typ == path
}

// IsOperator returns true if this Expression holds a SCIM filter operator (either logical or relational).
func (e *Expression) IsOperator() bool {
	return e.IsLogicalOperator() || e.IsRelationalOperator()
}

// IsLogicalOperator returns true if this Expression holds a SCIM logical filter operator.
func (e *Expression) IsLogicalOperator() bool {
	return e.typ == logicalOp
}

// IsRelationalOperator returns true if this Expression holds a SCIM relational filter operator.
func (e *Expression) IsRelationalOperator() bool {
	return e.typ == relationalOp
}

// IsRootOfFilter returns true if this Expression, while being on the linked list, is also an operator. This indicates
// that this Expression is the root of a filter tree. This method only does a simple test and shall only be called
// while traversing the linked list.
func (e *Expression) IsRootOfFilter() bool {
	return e.IsOperator() && e.left != nil
}

// IsLiteral returns true if this Expression represents a literal.
func (e *Expression) IsLiteral() bool {
	return e.typ == literal
}

// IsParenthesis returns true if this Expression is a parenthesis
func (e *Expression) IsParenthesis() bool {
	return e.typ == parenthesis
}

// IsLeftParenthesis returns true if this Expression is a left parenthesis
func (e *Expression) IsLeftParenthesis() bool {
	return e.typ == parenthesis && e.token == LeftParen
}

// IsRightParenthesis returns true if this Expression is a right parenthesis
func (e *Expression) IsRightParenthesis() bool {
	return e.typ == parenthesis && e.token == RightParen
}

// ContainsFilter returns true if the remaining of the path whose first node is represented
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
