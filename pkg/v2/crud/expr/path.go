package expr

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"strconv"
)

// CompilePath compiles the given SCIM path expression and returns the head of the path expression linked list, or any error.
// The result may contain a filter root node, depending on the given path expression.
//
// For example, for a path such as:
//	name.familyName
// CompilePath returns a structure like:
//	name -> familyName
//
// For a path such as:
//	emails[value eq "foo@bar.com"].primary
// CompilePath returns a structure like:
//	   emails -> eq -> primary
//	            /  \
//	         value  "foo@bar.com"
//
func CompilePath(path string) (*Expression, error) {
	compiler := &pathCompiler{
		scan: &pathScanner{},
		data: append(copyOf(path), 0, 0),
		off:  0,
		op:   scanPathContinue,
	}
	compiler.scan.init()

	head := &Expression{}
	cursor := head

	for compiler.hasMore() {
		next, err := compiler.next()
		if err != nil {
			return nil, err
		}
		cursor.next = next
		cursor = cursor.next
	}
	cursor = head.next
	head = cursor

	return head, nil
}

// Compiler that utilizes pathScanner to convert a string based path query to a linked list of steps, each representing
// a unit in the path.
type pathCompiler struct {
	scan *pathScanner
	// raw data of the path query
	data []byte
	// index for the next byte to be read
	off int
	// latest op code produced by the pathScanner
	op int
}

// Returns true if there could be more meaningful information to parsed.
func (c *pathCompiler) hasMore() bool {
	return c.op != scanPathEnd && c.op != scanPathError
}

// Produce the next token
func (c *pathCompiler) next() (*Expression, error) {
	if c.op == scanPathError {
		return nil, c.scan.err
	}

	if c.op == scanPathContinue {
		_ = c.skipWhile(scanPathContinue)
	}

	switch c.op {
	case scanPathEnd:
		return nil, nil
	case scanPathBeginStep:
		return c.scanStep()
	case scanPathBeginFilter:
		return c.scanFilter()
	default:
		return nil, c.errCompile()
	}
}

// Scan and make the path step after reading scanPathBeginStep op code.
func (c *pathCompiler) scanStep() (*Expression, error) {
	// start offset is one previous of the internal offset state because this function is only
	// invoked after seeing scanPathBeginStep, hence we are already one pass the actual start
	// of the step
	start := c.off - 1
	end := c.skipWhile(scanPathContinue)

	switch c.op {
	case scanPathEndStep, scanPathEnd:
		c.scanOne() // scan ahead to assist the next
		return &Expression{
			token: string(c.data[start:end]),
			typ:   path,
		}, nil
	case scanPathBeginFilter:
		return &Expression{
			token: string(c.data[start:end]),
			typ:   path,
		}, nil
	default:
		return nil, c.errCompile()
	}
}

// Scan and make the filter step after reading scanPathBeginFilter op code. The work is delegated filterCompiler
func (c *pathCompiler) scanFilter() (*Expression, error) {
	start := c.off
	end := c.skipWhile(scanPathContinue)
	switch c.op {
	case scanPathEndFilter, scanPathEnd:
		root, err := CompileFilter(string(c.data[start:end]))
		if err != nil {
			return nil, err
		}
		c.scanOne()
		return root, nil
	default:
		return nil, c.errCompile()
	}
}

// Scan the next byte of the data
func (c *pathCompiler) scanOne() {
	c.op = c.scan.step(c.scan, c.data[c.off])
	c.off++
}

// Return the offset of the first byte that produced a different op code. Note that given the nature of the scanning,
// the internal offset state will have passed the returned offset, because scanOne method must increment the offset
// after invoking the scanner's step function in order to be in sync with the scanner's state.
func (c *pathCompiler) skipWhile(op int) int {
	for c.off < len(c.data) {
		c.scanOne()
		if c.op != op {
			return c.off - 1
		}
	}

	// If we are having this given op code till the end, return an index greater
	// than the length of the data to signal that.
	return len(c.data) + 1
}

func (c *pathCompiler) errCompile() error {
	return fmt.Errorf("%w: error compiling path", spec.ErrInvalidPath)
}

// events reported by the path scanner, to be consumed by the path compiler.
const (
	scanPathContinue = iota
	scanPathBeginStep
	scanPathEndStep
	scanPathBeginFilter
	scanPathEndFilter
	scanPathError
	scanPathEnd
)

// Return a string based explanation for the path scanner events.
func explainPathEvent(op int) string {
	switch op {
	case scanPathContinue:
		return "continue"
	case scanPathBeginStep:
		return "begin step"
	case scanPathEndStep:
		return "end step"
	case scanPathBeginFilter:
		return "begin filter"
	case scanPathEndFilter:
		return "end filter"
	case scanPathError:
		return "error"
	case scanPathEnd:
		return "end"
	default:
		return "unknown"
	}
}

// A finite state machine that reports the event of the current byte in the path query. The events can be
// consumed by the path compiler to deduce semantic components. This scanner does not report events inside
// any included filters on the path, this should be delegated to filerScanner and filerCompiler.
type pathScanner struct {
	// the step function for the next byte
	step func(*pathScanner, byte) int
	// error incurred during the scanning. Once errored, the state machine shall remain
	// in error state.
	err error
	// number of bytes that has been scanned. This is assisting data that helps formulating
	// error information.
	bytes int64
}

// Initialize value of this scanner.
func (ps *pathScanner) init() {
	ps.step = ps.stateBeginStep
	ps.err = nil
	ps.bytes = 0
}

// Source state in which we are beginning the scanning of a step. Here, we expects a first character
// of a step which is either a namespace or an ordinary step. When this character fits some first character
// of a registered namespace, we shall attempt to match it to a namespace until there's a mismatch; otherwise,
// try viewing it as an ordinary step.
func (ps *pathScanner) stateBeginStep(scan *pathScanner, c byte) int {
	if !isFirstAlphabet(c) {
		return ps.error(c, "invalid character for the first alphabet of SCIM attribute name.")
	}

	match, ok := urnsCache.nextTrie(c)
	if ok {
		ps.step = ps.stateTryNamespaceStep(match)
	} else {
		scan.step = ps.stateInStep
	}

	return scanPathBeginStep
}

// Intermediate state in which we are in a step, but is trying to see if the current step is a reserved namespace.
// If the current character is still a match in the dictionary trie, the state is maintained; otherwise, attempt
// to downgrade the step state to an ordinary step state (stateInStep).
func (ps *pathScanner) stateTryNamespaceStep(root *urns) func(scan *pathScanner, c byte) int {
	return func(scan *pathScanner, c byte) int {
		match, ok := root.nextTrie(c)
		if ok {
			scan.step = ps.stateTryNamespaceStep(match)
			return scanPathContinue
		}

		if isNonFirstAlphabet(c) {
			scan.step = ps.stateInStep
			return scanPathContinue
		}

		if c == ':' && root.isWord() {
			scan.step = ps.stateBeginStep
			return scanPathEndStep
		}

		return ps.error(c, "invalid character after the initial SCIM attribute name character.")
	}
}

// Intermediate state in which we are in a step. A valid non-first SCIM attribute name character maintains current state;
// A path separator(.) starts a new step; A starting filter bracket, depending on the value of allowFilter, starts a filter
// or results in error; The artificial termination byte (byte 0) ends the step. Anything else results in error.
func (ps *pathScanner) stateInStep(scan *pathScanner, c byte) int {
	if isNonFirstAlphabet(c) {
		return scanPathContinue
	}

	switch c {
	case '.':
		scan.step = ps.stateBeginStep
		return scanPathEndStep
	case '[':
		scan.step = ps.stateInFilter
		return scanPathBeginFilter
	case 0:
		scan.step = ps.stateEof
		return scanPathEndStep
	}

	return ps.error(c, "invalid character after the initial SCIM attribute name character.")
}

// Intermediate state in which we are in the filter. In general, we skip through everything to look for the end of
// the filter (']'), because this scanner does not deal with filters. However, we need to carefully deal with double
// quote as it can entail a literal ending bracket, which does not count as the end of the filter.
func (ps *pathScanner) stateInFilter(scan *pathScanner, c byte) int {
	if c == '"' {
		scan.step = ps.stateInString
		return scanPathContinue
	}

	if c == ']' {
		scan.step = ps.stateEndFilter
		return scanPathEndFilter
	}

	return scanPathContinue
}

// Intermediate state in which we are concluded the filter. Here, we only expects a path separator, after which a new
// step shall begin; or the artificial terminating byte (byte 0) to signal to end of the path.
func (ps *pathScanner) stateEndFilter(scan *pathScanner, c byte) int {
	if c == '.' {
		scan.step = ps.stateBeginStep
		return scanPathContinue
	}

	if c == 0 {
		scan.step = ps.stateEof
		return scanPathEnd
	}

	return ps.error(c, "invalid character after the end of filter.")
}

// Intermediate state where we are inside a string. Another double quote shall end the string and return us to the filter
// state. However, we need the carefully treat the escape character as it can entail an escaped string ('\"foo\"') which
// should not be counted the double quote that ends the string literal.
func (ps *pathScanner) stateInString(scan *pathScanner, c byte) int {
	if c == '\\' {
		scan.step = ps.stateInStringEsc
	} else if c == '"' {
		scan.step = ps.stateInFilter
	}
	return scanPathContinue
}

// Intermediate state where we are inside an escaped string. Regular escape character return the state to stateInString.
// A unicode escape character (i.e \u0000) enter the state into escaped unicode string.
func (ps *pathScanner) stateInStringEsc(_ *pathScanner, c byte) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		ps.step = ps.stateInString
		return scanPathContinue
	case 'u':
		ps.step = ps.stateInStringEscU
		return scanPathContinue
	}
	return ps.error(c, "in string escape code")
}

// Intermediate state where we are at the leading byte of the 4 byte unicode.
func (ps *pathScanner) stateInStringEscU(_ *pathScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		ps.step = ps.stateInStringEscU1
		return scanPathContinue
	}
	// numbers
	return ps.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the second leading byte of the 4 byte unicode.
func (ps *pathScanner) stateInStringEscU1(_ *pathScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		ps.step = ps.stateInStringEscU12
		return scanPathContinue
	}
	// numbers
	return ps.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the third leading byte of the 4 byte unicode.
func (ps *pathScanner) stateInStringEscU12(_ *pathScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		ps.step = ps.stateInStringEscU123
		return scanPathContinue
	}
	// numbers
	return ps.error(c, "in \\u hexadecimal character escape")
}

// Intermediate state where we are at the last byte of the 4 byte unicode.
func (ps *pathScanner) stateInStringEscU123(_ *pathScanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		ps.step = ps.stateInString
		return scanPathContinue
	}
	// numbers
	return ps.error(c, "in \\u hexadecimal character escape")
}

// Sink state where the scanner has ended.
func (ps *pathScanner) stateEof(_ *pathScanner, _ byte) int {
	if ps.err != nil {
		return scanPathError
	}
	return scanPathEnd
}

// Sink state where the scanner ends up in error.
func (ps *pathScanner) stateError(_ *pathScanner, _ byte) int {
	return scanPathError
}

// Puts the scanner in error state and formulates the error, while passing on the error event code.
func (ps *pathScanner) error(c byte, hint string) int {
	ps.step = ps.stateError
	if len(hint) == 0 {
		hint = "n/a"
	}
	ps.err = fmt.Errorf("%w: invalid character %s around position %d (hint:%s)",
		spec.ErrInvalidPath, quoteChar(c), ps.bytes, hint)
	return scanPathError
}

// Returns true if the byte can be the first alphabet of a SCIM attribute name.
func isFirstAlphabet(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '$'
}

// Returns true if the byte can be the non-first alphabet of a SCIM attribute name.
func isNonFirstAlphabet(c byte) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9') ||
		c == '-' ||
		c == '_'
}

func quoteChar(c byte) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}

func copyOf(raw string) []byte {
	data := make([]byte, len(raw))
	copy(data, raw)
	return data
}
