package scim

import "strings"

const (
	And        = "and"
	Or         = "or"
	Not        = "not"
	Eq         = "eq"
	Ne         = "ne"
	Sw         = "sw"
	Ew         = "ew"
	Co         = "co"
	Pr         = "pr"
	Gt         = "gt"
	Ge         = "ge"
	Lt         = "lt"
	Le         = "le"
	leftParen  = "("
	rightParen = ")"
)

type ExprType int

const (
	_ ExprType = iota
	ExprTypePath
	ExprTypeLogical
	ExprTypeRelational
	ExprTypeLiteral
	ExprTypeParenthesis
)

// Expr represents a node in the compiled AST tree for SCIM paths and/or filters. It dubbed as a node in both a
// linked-list and a binary tree. When in a linked-list, it represents a single path segment. When in a binary tree,
// it represents a component of the filter.
type Expr struct {
	value string
	nType ExprType
	next  *Expr
	left  *Expr
	right *Expr
}

func (n *Expr) Value() string {
	return n.value
}

func (n *Expr) IsPath() bool {
	return n.nType == ExprTypePath
}

func (n *Expr) IsLogicalOperator() bool {
	return n.nType == ExprTypeLogical
}

func (n *Expr) IsRelationalOperator() bool {
	return n.nType == ExprTypeRelational
}

func (n *Expr) IsOperator() bool {
	return n.IsLogicalOperator() || n.IsRelationalOperator()
}

func (n *Expr) IsLiteral() bool {
	return n.nType == ExprTypeLiteral
}

func (n *Expr) IsParenthesis() bool {
	return n.nType == ExprTypeParenthesis
}

func (n *Expr) IsLeftParenthesis() bool {
	return n.IsParenthesis() && n.value == leftParen
}

func (n *Expr) IsRightParenthesis() bool {
	return n.IsParenthesis() && n.value == rightParen
}

// IsFilterRoot tests whether this Expr represents a root is a filter AST binary tree. A Expr can be a filter root
// when it hosts an operator (either logical or relational) and its left child is set (at least one operand).
func (n *Expr) IsFilterRoot() bool {
	return n.IsOperator() && n.left != nil
}

// containsFilter tests whether the expression represented by this Expr and its trailing nodes contains a filter.
func (n *Expr) containsFilter() bool {
	cur := n
	for cur != nil {
		if cur.IsFilterRoot() {
			return true
		}
		cur = cur.next
	}
	return false
}

// Walk performs traversal on the hybrid linked-list/binary tree structure of Expr. The callback function fn is
// invoked on each non-nil Expr.
func (n *Expr) Walk(fn func(n *Expr) error) error {
	if n == nil {
		return nil
	}

	if fn != nil {
		return fn(n)
	}

	if err := n.left.Walk(fn); err != nil {
		return err
	}

	if err := n.right.Walk(fn); err != nil {
		return err
	}

	return n.next.Walk(fn)
}

func (n *Expr) equals(m *Expr) bool {
	if n == nil {
		return m == nil
	}

	if n.nType != m.nType || strings.ToLower(n.value) != strings.ToLower(m.value) {
		return false
	}

	if leftEq := n.left.equals(m.left); !leftEq {
		return false
	}

	if rightEq := n.right.equals(m.right); !rightEq {
		return false
	}

	return n.next.equals(m.next)
}

func newOperator(op string) *Expr {
	switch strings.ToLower(op) {
	case And, Or, Not:
		return &Expr{value: op, nType: ExprTypeLogical}
	case Eq, Ne, Sw, Ew, Co, Gt, Ge, Lt, Le, Pr:
		return &Expr{value: op, nType: ExprTypeRelational}
	default:
		panic("unexpected token as operator")
	}
}

func newPath(name string) *Expr {
	return &Expr{value: name, nType: ExprTypePath}
}

func newLiteral(value string) *Expr {
	return &Expr{value: value, nType: ExprTypeLiteral}
}

func newParenthesis(paren string) *Expr {
	return &Expr{value: paren, nType: ExprTypeParenthesis}
}

// opPriority returns a priority rating for the given operator. Relational operators usually have higher
// priority than logical operators.
func opPriority(op string) int {
	switch strings.ToLower(op) {
	case And, Or, Not:
		return 50
	case Eq, Ne, Sw, Ew, Co, Pr, Gt, Ge, Lt, Le:
		return 100
	default:
		panic("not an operator")
	}
}

// isLeftAssociative returns true for the operator if it is left associative. Associativity defines that, when a
// sequence of three or more expressions are combined, whether it would be evaluated left to right (left associative),
// or right to left (right associative). In the context of SCIM operators, the negation operator (not) is right
// associative, and the rest is left associative.
func isLeftAssociative(op string) bool {
	switch strings.ToLower(op) {
	case Not:
		return false
	case And, Or, Eq, Ne, Sw, Ew, Co, Pr, Gt, Ge, Lt, Le:
		return true
	default:
		panic("not an operator")
	}
}

// opCardinality returns the number of expected operands for each operator. The negation operator (not) and the
// present operator (pr) expect one operand, and the rest expects two operands.
func opCardinality(op string) int {
	switch op {
	case Not, Pr:
		return 1
	case And, Or, Eq, Ne, Sw, Ew, Co, Gt, Ge, Lt, Le:
		return 2
	default:
		panic("not an operator")
	}
}
