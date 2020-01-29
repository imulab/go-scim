package expr

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"strings"
)

// CompileFilter compiles the given SCIM filter and return the root of the abstract syntax tree, or any error.
//
// For example, for a filter such as:
//	(value eq "foo") and (primary ne true)
// CompileFilter will return an abstract syntax tree in the structure of:
//	               and
//	             /    \
//	          eq        \
//	         /  \         \
//	     value  "foo"     ne
//	                     /  \
//	                primary true
//
func CompileFilter(filter string) (*Expression, error) {
	compiler := &filterCompiler{
		scan:    &filterScanner{},
		data:    append(copyOf(filter), 0, 0),
		off:     0,
		op:      scanFilterSkipSpace,
		opStack: make([]*Expression, 0),
		rsStack: make([]*Expression, 0),
	}
	compiler.scan.init()

	for compiler.hasMore() {
		step, err := compiler.next()
		if err != nil {
			return nil, err
		}

		if step == nil {
			break
		}

		if step.IsLiteral() || step.IsPath() {
			if err := compiler.pushBuildResult(step); err != nil {
				return nil, err
			}
			continue
		}

		switch compiler.pushOperator(step) {
		case pushOpOk:
			break

		case pushOpRightParenthesis:
			for {
				popped := compiler.popOperatorIf(func(top *Expression) bool {
					return !top.IsLeftParenthesis()
				})
				if popped != nil {
					// ignore error. we are sure it won't err
					_ = compiler.pushBuildResult(popped)
				} else {
					break
				}
			}
			if len(compiler.opStack) == 0 {
				return nil, fmt.Errorf("%w: mismatched parenthesis", spec.ErrInvalidFilter)
			} else {
				// discard the left parenthesis
				compiler.opStack = compiler.opStack[:len(compiler.opStack)-1]
			}
			break

		case pushOpInsufficientPriority:
			minPriority := opPriority(step.token)
			for {
				popped := compiler.popOperatorIf(func(top *Expression) bool {
					return opPriority(top.token) >= minPriority
				})
				if popped != nil {
					// ignore error. we are sure it won't err
					_ = compiler.pushBuildResult(popped)
				} else {
					break
				}
			}
			if compiler.pushOperator(step) != pushOpOk {
				panic("flaw in algorithm")
			}
			break
		}
	}

	// pop all remaining operators
	for len(compiler.opStack) > 0 {
		_ = compiler.pushBuildResult(compiler.popOperatorIf(func(top *Expression) bool {
			return true
		}))
	}

	// assertion check
	if len(compiler.rsStack) != 1 || !compiler.rsStack[0].IsOperator() {
		panic("flaw in algorithm")
	}

	// pop off the root so the rest could be GC'ed
	root := compiler.rsStack[0]
	compiler.rsStack = nil

	return root, nil
}

// priority and precedence definitions
var (
	// function to return the relative priority
	opPriority = func(op string) int {
		switch strings.ToLower(op) {
		case And, Or, Not:
			return 50
		case Eq, Ne, Sw, Ew, Co, Pr, Gt, Ge, Lt, Le:
			return 100
		default:
			panic("not an operator")
		}
	}
	// function to return true if left associative, false if right associative
	opLeftAssociative = func(op string) bool {
		switch strings.ToLower(op) {
		case Not:
			return false
		case And, Or, Eq, Ne, Sw, Ew, Co, Pr, Gt, Ge, Lt, Le:
			return true
		default:
			panic("not an operator")
		}
	}
	// function to return operator cardinality
	opCardinality = func(op string) int {
		switch op {
		case Not, Pr:
			return 1
		case And, Or, Eq, Ne, Sw, Ew, Co, Gt, Ge, Lt, Le:
			return 2
		default:
			panic("not an operator")
		}
	}
)

// Compiler that utilizes filterScanner to convert a string based filter query to tree.
type filterCompiler struct {
	scan *filterScanner
	// raw filter in bytes, appended by termination bytes (byte 0)
	data []byte
	// index to the next byte to be read in data
	off int
	// latest op code produced by filter scanner
	op int
	// operator stack used by shunting yard algorithm
	opStack []*Expression
	// result/output stack used by shunting yard algorithm
	rsStack []*Expression
}

// Part of the shunting yard algorithm. Push the operator or parenthesis represented by the step argument onto the
// operator stack, maintain priority. Return code to instruct caller for next steps.
func (c *filterCompiler) pushOperator(step *Expression) int {
	if step.IsRightParenthesis() {
		return pushOpRightParenthesis
	}

	if step.IsOperator() && len(c.opStack) > 0 {
		// get top of stack and compare priority
		p := c.opStack[len(c.opStack)-1]
		if p.IsOperator() {
			if opLeftAssociative(step.token) && opPriority(step.token) <= opPriority(p.token) {
				return pushOpInsufficientPriority
			} else if opLeftAssociative(step.token) && opPriority(step.token) < opPriority(p.token) {
				return pushOpInsufficientPriority
			}
		}
	}

	// push onto stack
	c.opStack = append(c.opStack, step)
	return pushOpOk
}

// Part of the shunting yard algorithm. Pop operator off stack if it meets the condition. Else return nil.
func (c *filterCompiler) popOperatorIf(condition func(top *Expression) bool) *Expression {
	if len(c.opStack) == 0 {
		return nil
	}

	top := c.opStack[len(c.opStack)-1]
	if condition(top) {
		// pop stack
		c.opStack = c.opStack[0 : len(c.opStack)-1]
		return top
	}

	return nil
}

// Part of the shunting yard algorithm. Push operators onto the result stack to form a tree.
func (c *filterCompiler) pushBuildResult(step *Expression) error {
	// Literal: push and return
	if step.IsLiteral() {
		c.rsStack = append(c.rsStack, step)
		return nil
	}

	// Path: re-compile and push
	if step.IsPath() {
		head, err := CompilePath(step.token)
		if err != nil {
			return fmt.Errorf("%w: invalid path in filter", spec.ErrInvalidFilter)
		} else if head.ContainsFilter() {
			return fmt.Errorf("%w: illegal nested filter", spec.ErrInvalidFilter)
		}
		c.rsStack = append(c.rsStack, head)
		return nil
	}

	// Assertion check:
	// at this point, step must be operator and stack must not be empty
	if !step.IsOperator() || len(c.rsStack) == 0 {
		panic("algorithm flaw")
	}

	// Pop operators and literals based on operators' cardinality and assemble before
	// push back in.
	switch opCardinality(step.token) {
	case 1:
		{
			first := c.rsStack[len(c.rsStack)-1]
			c.rsStack = c.rsStack[:len(c.rsStack)-1]
			step.left = first
		}
	case 2:
		{
			first, second := c.rsStack[len(c.rsStack)-1], c.rsStack[len(c.rsStack)-2]
			c.rsStack = c.rsStack[:len(c.rsStack)-2]
			step.left = second
			step.right = first
		}
	default:
		panic("unsupported cardinality")
	}
	c.rsStack = append(c.rsStack, step)

	return nil
}

// Returns true if there could be more meaningful information to parsed.
func (c *filterCompiler) hasMore() bool {
	return c.op != scanFilterEnd && c.op != scanFilterError
}

// Produce the next token
func (c *filterCompiler) next() (*Expression, error) {
	if c.op == scanFilterError {
		return nil, c.scan.err
	}

	_ = c.scanWhile(scanFilterSkipSpace)

	switch c.op {
	case scanFilterEnd:
		return nil, nil
	case scanFilterParenthesis:
		return newParenthesis(string(c.data[c.off-1])), nil
	case scanFilterBeginAny:
		return c.scanOperatorOrPath()
	case scanFilterBeginPath:
		return c.scanPath()
	case scanFilterBeginOp:
		return c.scanOperator()
	case scanFilterBeginLiteral:
		return c.scanLiteral()
	default:
		return nil, c.errCompile()
	}
}

// Produce a literal node from the current offset. This should only be called when scanFilterBeginLiteral is returned
// as op code.
func (c *filterCompiler) scanLiteral() (*Expression, error) {
	start := c.off - 1
	end := c.scanWhile(scanFilterContinue)
	switch c.op {
	case scanFilterEndLiteral, scanFilterEnd:
		return newLiteral(string(c.data[start:end])), nil
	default:
		return nil, c.errCompile()
	}
}

// Produce an operator node from the current offset. This should only be called when scanFilterBeginOp is returned
// as op code.
func (c *filterCompiler) scanOperator() (*Expression, error) {
	start := c.off - 1
	end := c.scanWhile(scanFilterContinue)
	switch c.op {
	case scanFilterEndOp, scanFilterEnd:
		return newOperator(string(c.data[start:end])), nil
	default:
		return nil, c.errCompile()
	}
}

// Produce an path step from the current offset. Note the returned step will be the unsegmented whole of the path
// in the filter as the filterScanner does not break them down. The caller will need to call CompilePath to replace
// the step with the compiled head.
func (c *filterCompiler) scanPath() (*Expression, error) {
	start := c.off - 1
	end := c.scanWhile(scanFilterContinue)
	switch c.op {
	case scanFilterEndPath, scanFilterEnd:
		return newPath(string(c.data[start:end])), nil
	default:
		return nil, c.errCompile()
	}
}

// Produce a path or operator node from the current offset. This fits the case where we read 'n' at the start of the
// predicate, we are not quite sure whether it is going to be 'not' and just another path with prefix 'n'. This should
// only be called when scanFilterBeginAny is returned as op code.
func (c *filterCompiler) scanOperatorOrPath() (*Expression, error) {
	start := c.off - 1
	end := c.scanWhile(scanFilterContinue)
	switch c.op {
	case scanFilterEndPath:
		return newPath(string(c.data[start:end])), nil
	case scanFilterEndOp:
		return newOperator(string(c.data[start:end])), nil
	default:
		return nil, c.errCompile()
	}
}

// Continue to scan for next while the returned op code is equal to the given op code. This is used mainly to skip
// uninteresting bytes and arrive at the next beginning or ending of something important.
//
// This method also respects the op code 'scanFilterInsertSpace'. This is a virtual "go-back" instruction to indicate
// that we have arrived at something important, but the original op code to return will only implicitly indicate the
// ending of what was started. Although the offset is correct, it would be easier to understand if the scanner can
// explicitly tell us the end of what was started. The 'scanFilterInsertSpace' op code, as a result, instructs us to
// send a space byte (i.e. ' ') to the scanner, in order to receive that explicit ending op code instruction.
func (c *filterCompiler) scanWhile(op int) int {
	for c.off < len(c.data) {
		c.op = c.scan.step(c.scan, c.data[c.off])

		// scanner instructs us to insert space before rescanning the last bit.
		// hence we will not need to increment the offset here as it shall be
		// scanned again
		if c.op == scanFilterInsertSpace {
			c.op = c.scan.step(c.scan, ' ')
			return c.off
		}

		// increment the offset to be in sync with scanner
		c.off++
		if c.op != op {
			return c.off - 1
		}
	}

	return len(c.data) + 1
}

func (c *filterCompiler) errCompile() error {
	return fmt.Errorf("%w: error compiling filter", spec.ErrInvalidFilter)
}

// events reported by the filter scanner, to be consumed by the filter compiler.
const (
	scanFilterContinue = iota
	scanFilterSkipSpace
	scanFilterInsertSpace
	scanFilterBeginAny
	scanFilterBeginPath
	scanFilterEndPath
	scanFilterBeginOp
	scanFilterEndOp
	scanFilterBeginLiteral
	scanFilterEndLiteral
	scanFilterParenthesis
	scanFilterError
	scanFilterEnd
)

// return code for pushOperator method of the filterCompiler
const (
	pushOpOk = iota
	pushOpRightParenthesis
	pushOpInsufficientPriority
)

// Return a string based explanation for the filter scanner events.
func explainFilterEvent(op int) string {
	switch op {
	case scanFilterContinue:
		return "continue"
	case scanFilterSkipSpace:
		return "space"
	case scanFilterInsertSpace:
		return "insert space"
	case scanFilterBeginAny:
		return "any"
	case scanFilterBeginPath:
		return "begin path"
	case scanFilterEndPath:
		return "end path"
	case scanFilterBeginOp:
		return "begin op"
	case scanFilterEndOp:
		return "end op"
	case scanFilterBeginLiteral:
		return "begin literal"
	case scanFilterEndLiteral:
		return "end literal"
	case scanFilterParenthesis:
		return "paren"
	case scanFilterError:
		return "error"
	case scanFilterEnd:
		return "end"
	default:
		return "unknown"
	}
}

// A finite state machine that reports the event of the current byte in the filter query. The events can be
// consumed by the filter compiler to deduce semantic components. This scanner does not report step components
// of a attribute path, but merely the start and end of the path, further parsing should be delegated to pathScanner
// and pathCompiler.
type filterScanner struct {
	// step function for the next byte
	step func(*filterScanner, byte) int
	// parenthesis level, should be 0 at the end
	parenLevel int
	// error incurred during the scanning. Once errored, the state machine shall remain
	// in error state.
	err error
	// number of bytes that has been scanned. This is assisting data that helps formulating
	// error information.
	bytes int64
}

// Initialize the scanner for use
func (fs *filterScanner) init() {
	fs.step = fs.stateBeginPredicate
	fs.parenLevel = 0
	fs.err = nil
	fs.bytes = 0
}

// Source state of filter scanner. We expect a predicate here. A predicate can start with an attribute path name, or
// a left parenthesis (for grouping), or the first character of the 'not' logical operator.
func (fs *filterScanner) stateBeginPredicate(scan *filterScanner, c byte) int {
	if c == ' ' {
		return scanFilterSkipSpace
	}

	switch c {
	case 'n', 'N':
		// we are not sure whether the token would be 'not', or just a path starting with 'n'.
		scan.step = fs.stateN
		return scanFilterBeginAny
	case '(':
		fs.parenLevel++
		scan.step = fs.stateBeginPredicate
		return scanFilterParenthesis
	}

	// A simple alphabet that does not start with 'n' or 'N', hence could not be 'not': this
	// should be a path
	if isFirstAlphabet(c) {
		scan.step = fs.stateInPath
		return scanFilterBeginPath
	}

	return fs.error(c, "invalid character at the start of the predicate")
}

// Intermediate state where the predicate ends. A right parenthesis could signal the end of grouping; 'a' and 'o' could
// signal logical and/or operator; the termination byte could signal end of the filter.
func (fs *filterScanner) stateEndPredicate(scan *filterScanner, c byte) int {
	if c == ' ' {
		return scanFilterSkipSpace
	}

	switch c {
	case ')':
		scan.parenLevel--
		return scanFilterParenthesis
	case 'a', 'A':
		// logical and
		scan.step = fs.stateOpA
		return scanFilterBeginOp
	case 'o', 'O':
		// logical or
		scan.step = fs.stateOpO
		return scanFilterBeginOp
	case 0:
		scan.step = fs.stateEof
		return scanFilterEnd
	}

	return fs.error(c, "invalid character at the end of the predicate")
}

// Intermediate state where the last character was 'n' (case insensitive) at the start of the predicate. This could lead
// to a logical not operator if the current character is 'o' (case insensitive), or lead to a path name instead.
func (fs *filterScanner) stateN(scan *filterScanner, c byte) int {
	switch c {
	case 'o', 'O':
		scan.step = fs.stateNo
		return scanFilterContinue
	}

	// just a path
	if c == '.' || c == ':' || isNonFirstAlphabet(c) {
		scan.step = fs.stateInPath
		return scanFilterContinue
	}

	return fs.error(c, "invalid character in path")
}

// Intermediate state where the last two characters were 'n' and 'o' (case insensitive) at the start of the predicate.
// This could lead to a logical not operator if the current character is 't' (case insensitive), or lead to a path
// name instead.
func (fs *filterScanner) stateNo(scan *filterScanner, c byte) int {
	switch c {
	case 't', 'T':
		scan.step = fs.stateNot
		return scanFilterContinue
	}

	// just a path
	if c == '.' || c == ':' || isNonFirstAlphabet(c) {
		scan.step = fs.stateInPath
		return scanFilterContinue
	}

	return fs.error(c, "invalid character in path")
}

// Intermediate state where the last three characters were 'n', 'o' and 't' (case insensitive) at the start of the predicate.
// This could lead to a logical not operator if the current character is can indicate the end of an operator. Otherwise,
// this could only lead to a path name instead.
func (fs *filterScanner) stateNot(scan *filterScanner, c byte) int {
	switch c {
	case ' ':
		scan.step = fs.stateBeginPredicate
		return scanFilterEndOp
	case '(':
		// ask caller to replay with a space so we can enter the condition above
		// the parenthesis count will be incremented after the replay so we don't
		// deal with it here
		return scanFilterInsertSpace
	}

	// seem like just a path that starts with 'not' (i.e. notes.title)
	if c == '.' || c == ':' || isNonFirstAlphabet(c) {
		scan.step = fs.stateInPath
		return scanFilterContinue
	}

	return fs.error(c, "invalid character in path")
}

// Intermediate state where we are inside an attribute path name. A space character would end the path name and start
// an operator.
func (fs *filterScanner) stateInPath(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginOp
		return scanFilterEndPath
	}

	if c == '.' || c == ':' || isNonFirstAlphabet(c) {
		return scanFilterContinue
	}

	return fs.error(c, "invalid character in path")
}

// Intermediate state at the beginning of an operator defined by SCIM query protocol.
func (fs *filterScanner) stateBeginOp(scan *filterScanner, c byte) int {
	if c == ' ' {
		return scanFilterSkipSpace
	}

	switch c {
	case 'a', 'A':
		// and
		scan.step = fs.stateOpA
		return scanFilterBeginOp
	case 'c', 'C':
		// co
		scan.step = fs.stateOpC
		return scanFilterBeginOp
	case 'e', 'E':
		// eq, ew
		scan.step = fs.stateOpE
		return scanFilterBeginOp
	case 'g', 'G':
		// gt, ge
		scan.step = fs.stateOpG
		return scanFilterBeginOp
	case 'l', 'L':
		// lt, le
		scan.step = fs.stateOpL
		return scanFilterBeginOp
	case 'n', 'N':
		// not, ne
		scan.step = fs.stateOpN
		return scanFilterBeginOp
	case 'o', 'O':
		// or
		scan.step = fs.stateOpO
		return scanFilterBeginOp
	case 'p', 'P':
		// pr
		scan.step = fs.stateOpP
		return scanFilterBeginOp
	case 's', 'S':
		// sw
		scan.step = fs.stateOpS
		return scanFilterBeginOp
	}

	return fs.error(c, "invalid character in operator")
}

// Intermediate state in operator where the last character was 'a' (case insensitive). The current character must be
// 'n' (case insensitive) to lead to the logical and operator.
func (fs *filterScanner) stateOpA(scan *filterScanner, c byte) int {
	if c == 'n' || c == 'N' {
		scan.step = fs.stateOpAn
		return scanFilterContinue
	}
	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where the last two characters were 'a' and 'n' (case insensitive). The current
// character must be 'd' (case insensitive) to lead to the logical and operator.
func (fs *filterScanner) stateOpAn(scan *filterScanner, c byte) int {
	if c == 'd' || c == 'D' {
		scan.step = fs.stateOpAnd
		return scanFilterContinue
	}
	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where the last three characters were 'a', 'n' and 'd' (case insensitive). The current
// character must end the operator.
func (fs *filterScanner) stateOpAnd(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginPredicate
		return scanFilterEndOp
	}

	if c == '(' {
		return scanFilterInsertSpace
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where the last character was 'c' (case insensitive). The current character must be
// 'o' (case insensitive) to lead to a relational co operator.
func (fs *filterScanner) stateOpC(scan *filterScanner, c byte) int {
	if c == 'o' || c == 'O' {
		scan.step = fs.stateOpCo
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where the last two characters were 'c' and 'o' (case insensitive). The current
// character must be space to end the operator.
func (fs *filterScanner) stateOpCo(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last character was 'e' (case insensitive). The current character should be
// 'q' or 'w' (case insensitive) to lead to eq/ew relational operator.
func (fs *filterScanner) stateOpE(scan *filterScanner, c byte) int {
	switch c {
	case 'q', 'Q':
		scan.step = fs.stateOpEq
		return scanFilterContinue
	case 'w', 'W':
		scan.step = fs.stateOpEw
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'e' and 'q' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpEq(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'e' and 'w' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpEw(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last character was 'g' (case insensitive). The current character should be
// 't' or 'e' (case insensitive) to lead to gt/ge relational operator.
func (fs *filterScanner) stateOpG(scan *filterScanner, c byte) int {
	switch c {
	case 't', 'T':
		scan.step = fs.stateOpGt
		return scanFilterContinue
	case 'e', 'E':
		scan.step = fs.stateOpGe
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'g' and 't' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpGt(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'g' and 'e' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpGe(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last character was 'l' (case insensitive). The current character should be
// 't' or 'e' (case insensitive) to lead to gt/ge relational operator.
func (fs *filterScanner) stateOpL(scan *filterScanner, c byte) int {
	switch c {
	case 't', 'T':
		scan.step = fs.stateOpLt
		return scanFilterContinue
	case 'e', 'E':
		scan.step = fs.stateOpLe
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'l' and 't' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpLt(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'l' and 'e' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpLe(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last character was 'n' (case insensitive). The current character should be
// 'o' or 'e' (case insensitive) to lead to 'not' logical operator or 'ne' relational operator.
func (fs *filterScanner) stateOpN(scan *filterScanner, c byte) int {
	switch c {
	case 'o', 'O':
		scan.step = fs.stateOpNo
		return scanFilterContinue
	case 'e', 'E':
		scan.step = fs.stateOpNe
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'n' and 'e' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpNe(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'n' and 'o' (case insensitive). The current character
// must be 't' (case insensitive) to lead to 'not' logical operator.
func (fs *filterScanner) stateOpNo(scan *filterScanner, c byte) int {
	if c == 't' || c == 'T' {
		scan.step = fs.stateOpNot
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last three characters were 'n', 'o' and 't' (case insensitive). The current
// character must end the operator.
func (fs *filterScanner) stateOpNot(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginPredicate
		return scanFilterEndOp
	}

	if c == '(' {
		return scanFilterInsertSpace
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last character was 'o' (case insensitive). The current character should be
// 'r' (case insensitive) to lead to 'or' logical operator.
func (fs *filterScanner) stateOpO(scan *filterScanner, c byte) int {
	if c == 'r' || c == 'R' {
		scan.step = fs.stateOpOr
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'o' and 'r' (case insensitive). The current character
// must end the operator.
func (fs *filterScanner) stateOpOr(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginPredicate
		return scanFilterEndOp
	}

	if c == '(' {
		return scanFilterInsertSpace
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last character was 'p' (case insensitive). The current character should be
// 'r' (case insensitive) to lead to 'or' logical operator.
func (fs *filterScanner) stateOpP(scan *filterScanner, c byte) int {
	if c == 'r' || c == 'R' {
		scan.step = fs.stateOpPr
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 'p' and 'r' (case insensitive). The current character
// must end the predicate.
func (fs *filterScanner) stateOpPr(scan *filterScanner, c byte) int {
	if c == ' ' || c == 0 {
		scan.step = fs.stateEndPredicate
		return scanFilterEndOp
	}

	if c == ')' {
		return scanFilterInsertSpace
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last character was 'e' (case insensitive). The current character should be
// 'w' (case insensitive) to lead to sw relational operator.
func (fs *filterScanner) stateOpS(scan *filterScanner, c byte) int {
	if c == 'w' || c == 'W' {
		scan.step = fs.stateOpSw
		return scanFilterContinue
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state in operator where last two characters were 's' and 'w' (case insensitive). The current character
// must end the operator with space.
func (fs *filterScanner) stateOpSw(scan *filterScanner, c byte) int {
	if c == ' ' {
		scan.step = fs.stateBeginLiteral
		return scanFilterEndOp
	}

	return fs.errInvalidOperator(c)
}

// Intermediate state at the start of a literal. We distinguish between string and non-string literal.
func (fs *filterScanner) stateBeginLiteral(scan *filterScanner, c byte) int {
	switch c {
	case '"':
		scan.step = fs.stateInStringLiteral
		return scanFilterBeginLiteral
	case 't', 'T', 'f', 'F', '-', '+', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		scan.step = fs.stateInNonStringLiteral
		return scanFilterBeginLiteral
	}

	return fs.error(c, "invalid literal")
}

// Intermediate state at the end of a literal.
func (fs *filterScanner) stateEndLiteral(scan *filterScanner, c byte) int {
	if c == ' ' {
		return scanFilterSkipSpace
	}

	if c == ')' {
		fs.parenLevel--
		scan.step = fs.stateEndPredicate
		return scanFilterParenthesis
	}

	if c == 0 {
		scan.step = fs.stateEof
		return scanFilterEnd
	}

	// ending a literal can also mean ending a predicate implicitly
	return fs.stateEndPredicate(scan, c)
}

// Intermediate state in a string literal.
func (fs *filterScanner) stateInStringLiteral(scan *filterScanner, c byte) int {
	if c == '\\' {
		scan.step = fs.stateInStringEsc
	}

	if c == '"' {
		scan.step = fs.stateEndStringLiteral
	}

	return scanFilterContinue
}

// Intermediate state after ending a string literal with double quote. This state is necessary so that the reported
// events can be easily interpreted against the index to produce a string literal with starting and ending double quotes.
func (fs *filterScanner) stateEndStringLiteral(scan *filterScanner, c byte) int {
	switch c {
	case ' ':
		scan.step = fs.stateEndLiteral
		return scanFilterEndLiteral
	case ')':
		return scanFilterInsertSpace
	case 0:
		scan.step = fs.stateEof
		return scanFilterEndLiteral
	}

	return fs.error(c, "invalid character trailing string literal")
}

// Intermediate state in a non-string literal. Here, we only care about termination of the literal.
func (fs *filterScanner) stateInNonStringLiteral(scan *filterScanner, c byte) int {
	switch c {
	case ' ':
		scan.step = fs.stateEndLiteral
		return scanFilterEndLiteral
	case ')':
		return scanFilterInsertSpace
	case 0:
		scan.step = fs.stateEof
		return scanFilterEndLiteral
	default:
		return scanFilterContinue
	}
}

// Intermediate state where we are inside an escaped string. Regular escape character return the state to stateInString.
// A unicode escape character (i.e \u0000) enter the state into escaped unicode string.
func (fs *filterScanner) stateInStringEsc(_ *filterScanner, c byte) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		fs.step = fs.stateInStringLiteral
		return scanFilterContinue
	case 'u':
		fs.step = fs.stateInStringEscU
		return scanFilterContinue
	}
	return fs.error(c, "invalid character in string literal")
}

// Intermediate state where we are at the leading byte of the 4 byte unicode.
func (fs *filterScanner) stateInStringEscU(_ *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringEscU1
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the second leading byte of the 4 byte unicode.
func (fs *filterScanner) stateInStringEscU1(_ *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringEscU12
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the third leading byte of the 4 byte unicode.
func (fs *filterScanner) stateInStringEscU12(_ *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringEscU123
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the last byte of the 4 byte unicode.
func (fs *filterScanner) stateInStringEscU123(_ *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringLiteral
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Sink state where the scanner has ended.
func (fs *filterScanner) stateEof(_ *filterScanner, _ byte) int {
	if fs.err != nil {
		return scanFilterError
	}

	if fs.parenLevel != 0 {
		fs.step = fs.stateError
		fs.err = fmt.Errorf("%w: mismatched parenthesis", spec.ErrInvalidFilter)
		return scanPathError
	}

	return scanFilterEnd
}

// Sink state where the scanner ends up in error.
func (fs *filterScanner) stateError(_ *filterScanner, _ byte) int {
	return scanFilterError
}

// Puts the scanner in error state and formulates the error, while passing on the error event code.
func (fs *filterScanner) error(c byte, hint string) int {
	fs.step = fs.stateError
	if len(hint) == 0 {
		hint = "n/a"
	}
	fs.err = fmt.Errorf("%w: invalid character %s around position %d (hint:%s)",
		spec.ErrInvalidFilter, quoteChar(c), fs.bytes, hint)
	return scanFilterError
}

func (fs *filterScanner) errInvalidOperator(c byte) int {
	return fs.error(c, "invalid operator")
}
