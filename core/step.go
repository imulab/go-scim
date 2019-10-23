package core

import (
	"fmt"
	"strconv"
	"strings"
)

// step types
const (
	stepPath = iota
	stepLogicalOperator
	stepRelationalOperator
	stepLiteral
	stepParenthesis
)

// A hybrid node representing a segment of path or a node in the filter tree.
type Step struct {
	Token	string
	Typ		int
	Next	*Step
	Left	*Step
	Right	*Step
}

// Returns true if this step represents a path (segment).
func (s Step) IsPath() bool {
	return s.Typ == stepPath
}

// Returns true if this step represents a logical or relational operator.
func (s Step) IsOperator() bool {
	return s.Typ == stepLogicalOperator || s.Typ == stepRelationalOperator
}

// Returns true if this step represents a literal value.
func (s Step) IsLiteral() bool {
	return s.Typ == stepLiteral
}

// Returns true if this step represents a parenthesis.
func (s Step) IsParenthesis() bool {
	return s.Typ == stepParenthesis
}

// Strip quotes around the value. This method is supposed to be called when
// the caller knows or assumes this step is a value step and contains a string
// value.
func (s Step) stripQuotes() string {
	v := s.Token
	v = strings.TrimPrefix(v, "\"")
	v = strings.TrimSuffix(v, "\"")
	return v
}

// Parse the value of this step and return the value in its compliant Go type.
func (s Step) compliantValue(attr *Attribute) (val interface{}, err error) {
	switch attr.Type {
	case TypeString, TypeReference, TypeDateTime, TypeBinary:
		val = s.stripQuotes()
		err = nil
	case TypeInteger:
		val, err = strconv.ParseInt(s.Token, 10, 64)
	case TypeDecimal:
		val, err = strconv.ParseFloat(s.Token, 64)
	case TypeBoolean:
		val, err = strconv.ParseBool(s.Token)
	default:
		panic("not a value")
	}

	if err != nil {
		err = Errors.invalidValue(fmt.Sprintf("'%s' is not a valid value for %s.", s.Token, attr.Type))
	}

	return
}

var (
	// Entry point for creating steps.
	Steps = stepFactory{}
	// Singleton for left and right parenthesis step.
	leftParenStep = &Step{
		Token: LeftParen,
		Typ:   stepParenthesis,
	}
	rightParenStep = &Step{
		Token: RightParen,
		Typ: stepParenthesis,
	}
)

// Factory methods for creating a new step
type stepFactory struct{}

// Create a new path step
func (f stepFactory) NewPath(path string) *Step {
	return &Step{
		Token: path,
		Typ:   stepPath,
	}
}

// Create a new linked list of path steps and return its head.
func (f stepFactory) NewPathChain(paths ...string) *Step {
	if len(paths) < 0 {
		return nil
	}

	var (
		head = &Step{} // dummy head
		cursor = head
	)
	for _, path := range paths {
		cursor.Next = f.NewPath(path)
		cursor = cursor.Next
	}

	return head.Next
}

// Create a new logical operator step
func (f stepFactory) NewLogicalOperator(op string) *Step {
	return &Step{
		Token: op,
		Typ:   stepLogicalOperator,
	}
}

// Create a new relational operator step
func (f stepFactory) NewRelationalOperator(op string) *Step {
	return &Step{
		Token: op,
		Typ:   stepRelationalOperator,
	}
}

// Create a new literal step.
func (f stepFactory) NewLiteral(val string) *Step {
	return &Step{
		Token: val,
		Typ:   stepLiteral,
	}
}

// Create (return) a new parenthesis step.
func (f stepFactory) NewParenthesis(paren string) *Step {
	switch paren {
	case LeftParen:
		return leftParenStep
	case RightParen:
		return rightParenStep
	default:
		panic("not a parenthesis")
	}
}