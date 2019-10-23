package query

import (
	"fmt"
	"github.com/imulab/go-scim/core"
	"strings"
)

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
	if c == ' ' {
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
func (fs *filterScanner) stateInStringEsc(scan *filterScanner, c byte) int {
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
func (fs *filterScanner) stateInStringEscU(scan *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringEscU1
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the second leading byte of the 4 byte unicode.
func (fs *filterScanner) stateInStringEscU1(scan *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringEscU12
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the third leading byte of the 4 byte unicode.
func (fs *filterScanner) stateInStringEscU12(scan *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringEscU123
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the last byte of the 4 byte unicode.
func (fs *filterScanner) stateInStringEscU123(scan *filterScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		fs.step = fs.stateInStringLiteral
		return scanFilterContinue
	}

	return fs.error(c, "in \\u hexadecimal character escape")
}

// Sink state where the scanner has ended.
func (fs *filterScanner) stateEof(scan *filterScanner, c byte) int {
	if fs.err != nil {
		return scanFilterError
	}

	if fs.parenLevel != 0 {
		fs.step = fs.stateError
		fs.err = core.Errors.InvalidFilter("mismatched parenthesis")
		return scanPathError
	}

	return scanFilterEnd
}

// Sink state where the scanner ends up in error.
func (fs *filterScanner) stateError(scan *filterScanner, c byte) int {
	return scanFilterError
}

// Puts the scanner in error state and formulates the error, while passing on the error event code.
func (fs *filterScanner) error(c byte, context string) int {
	fs.step = fs.stateError
	fs.err = core.Errors.InvalidFilter(strings.TrimSpace(
		fmt.Sprintf("invalid character %s around %d. %s ", quoteChar(c), fs.bytes, context),
	))
	return scanPathError
}

func (fs *filterScanner) errInvalidOperator(c byte) int {
	return fs.error(c, "invalid operator")
}