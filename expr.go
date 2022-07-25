package scim

import "strings"

// todo: control complexity and move to internal package, expose compiler and node
func newExpressionCompiler(schemas ...*Schema) *exprCompiler {
	c := &exprCompiler{urnDict: &trie{}}
	for _, sch := range schemas {
		c.urnDict = c.urnDict.insert(c.urnDict, sch.id, 0)
	}
	return c
}

type exprCompiler struct {
	urnDict *trie
}

type exprType int

const (
	exprTypePath exprType = iota
	exprTypeLogical
	exprTypeRelational
	exprTypeLiteral
	exprTypeParenthesis
)

type exprNode struct {
	val   string
	typ   exprType
	left  *exprNode
	right *exprNode
	next  *exprNode
}

func (e *exprNode) isPath() bool {
	return e.typ == exprTypePath
}

func (e *exprNode) isLogicalOperator() bool {
	return e.typ == exprTypeLogical
}

func (e *exprNode) isRelationalOperator() bool {
	return e.typ == exprTypeRelational
}

func (e *exprNode) isOperator() bool {
	return e.isLogicalOperator() || e.isRelationalOperator()
}

func (e *exprNode) isLiteral() bool {
	return e.typ == exprTypeLiteral
}

func (e *exprNode) isParenthesis() bool {
	return e.typ == exprTypeParenthesis
}

func (e *exprNode) isLeftParenthesis() bool {
	return e.isParenthesis() && e.val == leftParen
}

func (e *exprNode) isRightParenthesis() bool {
	return e.isParenthesis() && e.val == rightParen
}

func (e *exprNode) isFilterRoot() bool {
	return (e.isLogicalOperator() || e.isRelationalOperator()) && e.left != nil
}

func (e *exprNode) containsFilter() bool {
	cur := e

	for cur != nil {
		if cur.isFilterRoot() {
			return true
		}
		cur = cur.next
	}

	return false
}

func (e *exprNode) walk(fn func(expr *exprNode), doneMarker *exprNode, doneFn func()) {
	if e == nil {
		return
	}

	fn(e)

	e.left.walk(fn, doneMarker, doneFn)
	e.right.walk(fn, doneMarker, doneFn)
	e.next.walk(fn, doneMarker, doneFn)

	if doneMarker == e {
		doneFn()
	}
}

func newOperator(op string) *exprNode {
	switch strings.ToLower(op) {
	case and, or, not:
		return &exprNode{val: op, typ: exprTypeLogical}
	case eq, ne, sw, ew, co, gt, ge, lt, le, pr:
		return &exprNode{val: op, typ: exprTypeRelational}
	default:
		panic("unexpected token as operator")
	}
}

func newPath(name string) *exprNode {
	return &exprNode{val: name, typ: exprTypePath}
}

func newLiteral(value string) *exprNode {
	return &exprNode{val: value, typ: exprTypeLiteral}
}

func newParenthesis(paren string) *exprNode {
	return &exprNode{val: paren, typ: exprTypeParenthesis}
}
