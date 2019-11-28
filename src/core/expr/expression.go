package expr

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