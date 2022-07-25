package scim

import (
	"strings"
)

const (
	leftParen  = "("
	rightParen = ")"
	and        = "and"
	or         = "or"
	not        = "not"
	eq         = "eq"
	ne         = "ne"
	sw         = "sw"
	ew         = "ew"
	co         = "co"
	pr         = "pr"
	gt         = "gt"
	ge         = "ge"
	lt         = "lt"
	le         = "le"
)

func getOpPriority(op string) int {
	switch strings.ToLower(op) {
	case and, or, not:
		return 50
	case eq, ne, sw, ew, co, pr, gt, ge, lt, le:
		return 100
	default:
		panic("unexpected op as operator")
	}
}

func getOpLeftAssociativity(op string) bool {
	switch strings.ToLower(op) {
	case not:
		return false
	case and, or, eq, ne, sw, ew, co, pr, gt, ge, lt, le:
		return true
	default:
		panic("unexpected op as operator")
	}
}

func getOpCardinality(op string) int {
	switch strings.ToLower(op) {
	case not, pr:
		return 1
	case and, or, eq, ne, sw, ew, co, gt, ge, lt, le:
		return 2
	default:
		panic("unexpected op as operator")
	}
}
