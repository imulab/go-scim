package expr

import "strings"

type nodeType int

const (
	_ nodeType = iota
	nodeTypePath
	nodeTypeLogical
	nodeTypeRelational
	nodeTypeLiteral
	nodeTypeParenthesis
)

// Node represents a node in the compiled AST tree for SCIM paths and/or filters. It dubbed as a node in both a
// linked-list and a binary tree. When in a linked-list, it represents a single path segment. When in a binary tree,
// it represents a component of the filter.
type Node struct {
	value string
	nType nodeType
	next  *Node
	left  *Node
	right *Node
}

func (n *Node) IsPath() bool {
	return n.nType == nodeTypePath
}

func (n *Node) IsLogicalOperator() bool {
	return n.nType == nodeTypeLogical
}

func (n *Node) IsRelationalOperator() bool {
	return n.nType == nodeTypeRelational
}

func (n *Node) IsOperator() bool {
	return n.IsLogicalOperator() || n.IsRelationalOperator()
}

func (n *Node) IsLiteral() bool {
	return n.nType == nodeTypeLiteral
}

func (n *Node) IsParenthesis() bool {
	return n.nType == nodeTypeParenthesis
}

func (n *Node) IsLeftParenthesis() bool {
	return n.IsParenthesis() && n.value == leftParen
}

func (n *Node) IsRightParenthesis() bool {
	return n.IsParenthesis() && n.value == rightParen
}

// IsFilterRoot tests whether this Node represents a root is a filter AST binary tree. A Node can be a filter root
// when it hosts an operator (either logical or relational) and its left child is set (at least one operand).
func (n *Node) IsFilterRoot() bool {
	return n.IsOperator() && n.left != nil
}

// containsFilter tests whether the expression represented by this Node and its trailing nodes contains a filter.
func (n *Node) containsFilter() bool {
	cur := n
	for cur != nil {
		if cur.IsFilterRoot() {
			return true
		}
		cur = cur.next
	}
	return false
}

// walk performs traversal on the hybrid linked-list/binary tree structure of Node. The callback function fn is
// invoked on each non-nil Node.
func (n *Node) walk(fn func(n *Node)) {
	if n == nil {
		return
	}

	if fn != nil {
		fn(n)
	}

	n.left.walk(fn)
	n.right.walk(fn)
	n.next.walk(fn)
}

func newOperator(op string) *Node {
	switch strings.ToLower(op) {
	case and, or, not:
		return &Node{value: op, nType: nodeTypeLogical}
	case eq, ne, sw, ew, co, gt, ge, lt, le, pr:
		return &Node{value: op, nType: nodeTypeRelational}
	default:
		panic("unexpected token as operator")
	}
}

func newPath(name string) *Node {
	return &Node{value: name, nType: nodeTypePath}
}

func newLiteral(value string) *Node {
	return &Node{value: value, nType: nodeTypeLiteral}
}

func newParenthesis(paren string) *Node {
	return &Node{value: paren, nType: nodeTypeParenthesis}
}
