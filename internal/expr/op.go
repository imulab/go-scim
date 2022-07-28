package expr

import "strings"

const (
	And = "and"
	Or  = "or"
	Not = "not"
	Eq  = "eq"
	Ne  = "ne"
	Sw  = "sw"
	Ew  = "ew"
	Co  = "co"
	Pr  = "pr"
	Gt  = "gt"
	Ge  = "ge"
	Lt  = "lt"
	Le  = "le"
)

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
