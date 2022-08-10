package scim

import (
	"fmt"
	"github.com/imulab/go-scim/internal/jsonutil"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// deserialize is the entrypoint to deserialize JSON
func deserialize(json []byte, navigator *navigator) error {
	if err := jsonutil.CheckValid(json, &jsonutil.Scanner{}); err != nil {
		return err
	}

	state := &deserializeState{
		data:      json,
		off:       0,
		opCode:    jsonutil.ScanContinue,
		scan:      jsonutil.Scanner{},
		navigator: navigator,
	}
	state.scan.Reset()

	// skip the first few spaces
	state.scanWhile(jsonutil.ScanSkipSpace)
	return state.parseComplexProperty(false)
}

// State of the deserialization process. In essence, a scanner is used to infer contextual information about what the
// current byte means. It is used in conjunction with off (offset) and opCode. In addition, a navigator is used to record
// the tracks of traversal inside the complete structure of the resource. The location of properties can be in sync with
// the current context by reacting to some signals emitted by the scanner, such as scanStartObject or scanEndArray.
//
// As a side note, all parseXXX methods of this object shall maintain one courtesy: after done parsing the part of the
// data of interest to the method, consume as much empty spaces or separators (i.e. scanObjectValue, scanArrayValue) as
// possible so that the next parseXXX method invoked will not have to skip spaces as its first task.
type deserializeState struct {
	data      []byte
	off       int // next read offset in data
	opCode    int // last read result
	scan      jsonutil.Scanner
	navigator *navigator
}

func (d *deserializeState) errInvalidSyntax(msg string, args ...interface{}) error {
	return fmt.Errorf("%w: %s (pos:%d)", ErrInvalidSyntax, fmt.Sprintf(msg, args...), d.off)
}

// Parses the attribute/field name in a JSON object. This method expects a quoted string and skips through
// as much empty spaces and colon (appears as scanObjectKey) after it as possible.
func (d *deserializeState) parseFieldName() (string, error) {
	if d.opCode != jsonutil.ScanBeginLiteral {
		return "", d.errInvalidSyntax("expects attribute name")
	}

	start := d.off - 1 // position of the first double quote
	d.scanWhile(jsonutil.ScanContinue)
	end := d.off - 1 // position of the character after the second double quote

Skip:
	for {
		switch d.opCode {
		case jsonutil.ScanObjectKey, jsonutil.ScanSkipSpace:
			d.scanNext()
		default:
			break Skip
		}
	}

	return string(d.data[start+1 : end-1]), nil
}

// Parses a top level or embedded JSON object. When parsing a top level object, allowNull shall be false as top level
// object does not correspond to any field name and hence cannot be null; when parsing an embedded object, allowNull may
// be true. This method expects '{' (appears as scanBeginObject) to be the current byte
func (d *deserializeState) parseComplexProperty(allowNull bool) error {
	// expects '{', and depending on allowNull, allowing for the null literal.
	if d.opCode != jsonutil.ScanBeginObject {
		if allowNull && d.opCode == jsonutil.ScanBeginLiteral {
			return d.parseNull()
		}
		return d.errInvalidSyntax("expects a json object")
	}

	// skip any potential spaces between '{' and '"'
	d.scanWhile(jsonutil.ScanSkipSpace)

kvs:
	for d.opCode != jsonutil.ScanEndObject {
		// Focus on the property that corresponds to the field name
		var (
			p   Property
			err error
		)
		{
			attrName, err := d.parseFieldName()
			if err != nil {
				return err
			}
			p = d.navigator.dot(attrName).current()
			if d.navigator.hasError() {
				return d.navigator.err
			}
		}

		// Parse field value
		if p.Attr().multiValued {
			err = d.parseMultiValuedProperty()
		} else {
			err = d.parseSingleValuedProperty()
		}
		if err != nil {
			return err
		}

		// Exit focus on the field value property
		d.navigator.retract()

		// Fast forward to the next field name/value pair, or exit the loop.
	fastForward:
		for {
			switch d.opCode {
			case jsonutil.ScanEndObject:
				d.scanNext()
				break kvs
			case jsonutil.ScanEnd:
				break kvs
			case jsonutil.ScanSkipSpace, jsonutil.ScanObjectValue, jsonutil.ScanEndArray:
				d.scanNext()
			default:
				break fastForward
			}
		}
	}

	// Courtesy: skip any spaces between '}' and the next tokens
	if d.opCode == jsonutil.ScanSkipSpace {
		d.scanWhile(jsonutil.ScanSkipSpace)
	}

	return nil
}

// Delegate method to parse single valued field values. The caller must ensure that the currently focused property
// is indeed single valued.
func (d *deserializeState) parseSingleValuedProperty() error {
	switch d.navigator.current().Attr().typ {
	case TypeString, TypeDateTime, TypeBinary, TypeReference:
		return d.parseStringProperty()
	case TypeInteger:
		return d.parseIntegerProperty()
	case TypeDecimal:
		return d.parseDecimalProperty()
	case TypeBoolean:
		return d.parseBooleanProperty()
	case TypeComplex:
		return d.parseComplexProperty(true)
	default:
		panic("invalid attribute type")
	}
}

// Parses a JSON array. This method expects '[' (appears as scanBeginArray) to be the current byte, or the literal
// null.
func (d *deserializeState) parseMultiValuedProperty() error {
	// Expect '[' or null.
	if d.opCode != jsonutil.ScanBeginArray {
		if d.opCode == jsonutil.ScanBeginLiteral {
			return d.parseNull()
		}
		return d.errInvalidSyntax("expects JSON array")
	}

	// Skip any spaces between '[' and the potential first element
	d.scanNext()
	if d.opCode == jsonutil.ScanSkipSpace {
		d.scanWhile(jsonutil.ScanSkipSpace)
	}

elements:
	for d.opCode != jsonutil.ScanEndArray {
		mv, ok := d.navigator.current().(appendElement)
		if !ok {
			return d.errInvalidSyntax("non-multiValued property at json array")
		}

		i, post := mv.appendElement()
		if d.navigator.at(i); d.navigator.hasError() {
			return d.navigator.err
		}

		// Parse the focused element property
		err := d.parseSingleValuedProperty()
		if err != nil {
			return err
		}

		// Invoke the cleanup function on the multiValued property: multi-primary, unassigned, duplicate items will be
		// cleaned up automatically. This is a new behavior in v3 where we no longer strongly distinguish unassigned-ness
		// and explicit nulls (as a result, a simple property impl), as the end-game is now user-defined models. If a
		// mapping cannot locate a property in the resource, it is deemed null anyway.
		post()

		// Exit the focus
		d.navigator.retract()

		// Fast forward to the next element, or exit the loop.
	fastForward:
		for {
			switch d.opCode {
			case jsonutil.ScanEndArray:
				d.scanNext()
				break elements
			case jsonutil.ScanSkipSpace, jsonutil.ScanArrayValue:
				d.scanNext()
			default:
				break fastForward
			}
		}
	}

	// Courtesy: skip any spaces between ']' and the next tokens
	if d.opCode == jsonutil.ScanSkipSpace {
		d.scanWhile(jsonutil.ScanSkipSpace)
	}

	return nil
}

// Parses a JSON string. This method expects a double quoted literal and the null literal.
func (d *deserializeState) parseStringProperty() error {
	p := d.navigator.current()

	// check property type
	if p.Attr().multiValued || (p.Attr().typ != TypeString &&
		p.Attr().typ != TypeDateTime &&
		p.Attr().typ != TypeReference &&
		p.Attr().typ != TypeBinary) {
		return d.errInvalidSyntax("expects string based property for '%s'", p.Attr().path)
	}

	// should start with literal
	if d.opCode != jsonutil.ScanBeginLiteral {
		return d.errInvalidSyntax("expects json literal")
	}

	start := d.off - 1 // position of the first double quote
	d.scanWhile(jsonutil.ScanContinue)
	end := d.off - 1 // position of the character after the second double quote

	if d.isNull(start, end) {
		d.navigator.current().Delete()
		return nil
	}

	if d.data[start] != '"' || d.data[end-1] != '"' {
		return d.errInvalidSyntax("expects string literal value for '%s'", p.Attr().path)
	}

	v, ok := unquote(d.data[start:end])
	if !ok {
		return d.errInvalidSyntax("failed to unquote json string for '%s'", p.Attr().path)
	}

	if err := d.navigator.current().Set(v); err != nil {
		return err
	}

	return nil
}

// Parses a JSON integer. This method expects an integer literal and the null literal.
func (d *deserializeState) parseIntegerProperty() error {
	p := d.navigator.current()

	// check property type
	if p.Attr().multiValued || p.Attr().typ != TypeInteger {
		return d.errInvalidSyntax("expects integer property for '%s'", p.Attr().path)
	}

	// should start with literal
	if d.opCode != jsonutil.ScanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(jsonutil.ScanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if d.isNull(start, end) {
		d.navigator.current().Delete()
		return nil
	}

	val, err := strconv.ParseInt(string(d.data[start:end]), 10, 64)
	if err != nil {
		return d.errInvalidSyntax("expects integer value")
	}

	if err := d.navigator.current().Set(val); err != nil {
		return err
	}

	return nil
}

// Parses a JSON boolean. This method expects the true, false, or null literal.
func (d *deserializeState) parseBooleanProperty() error {
	p := d.navigator.current()

	// check property type
	if p.Attr().multiValued || p.Attr().typ != TypeBoolean {
		return d.errInvalidSyntax("expects decimal property for '%s'", p.Attr().path)
	}

	// should start with literal
	if d.opCode != jsonutil.ScanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(jsonutil.ScanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if d.isNull(start, end) {
		d.navigator.current().Delete()
		return nil
	}

	if d.isTrue(start, end) {
		if err := d.navigator.current().Set(true); err != nil {
			return err
		}
	} else if d.isFalse(start, end) {
		if err := d.navigator.current().Set(false); err != nil {
			return err
		}
	} else {
		if isHackingForMicrosoft, err := d.tryHackForMicrosoftADBooleanIssue(p, start, end); isHackingForMicrosoft {
			return err
		}
		return d.errInvalidSyntax("expects boolean value")
	}

	return nil
}

// Parses a JSON decimal. This method expects a decimal literal and the null literal.
func (d *deserializeState) parseDecimalProperty() error {
	p := d.navigator.current()

	// check property type
	if p.Attr().multiValued || p.Attr().typ != TypeDecimal {
		return d.errInvalidSyntax("expects decimal property for '%s'", p.Attr().path)
	}

	// should start with literal
	if d.opCode != jsonutil.ScanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(jsonutil.ScanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if d.isNull(start, end) {
		d.navigator.current().Delete()
		return nil
	}

	val, err := strconv.ParseFloat(string(d.data[start:end]), 64)
	if err != nil {
		return d.errInvalidSyntax("expects decimal value")
	}

	if err := d.navigator.current().Set(val); err != nil {
		return err
	}

	return nil
}

// Parses the JSON null literal.
func (d *deserializeState) parseNull() error {
	// should start with literal
	if d.opCode != jsonutil.ScanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(jsonutil.ScanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if !d.isNull(start, end) {
		return d.errInvalidSyntax("expects null")
	}

	d.navigator.current().Delete()

	return nil
}

func (d *deserializeState) isNull(start, end int) bool {
	return end-start == 4 &&
		d.data[start] == 'n' &&
		d.data[start+1] == 'u' &&
		d.data[start+2] == 'l' &&
		d.data[start+3] == 'l'
}

func (d *deserializeState) isTrue(start, end int) bool {
	return end-start == 4 &&
		d.data[start] == 't' &&
		d.data[start+1] == 'r' &&
		d.data[start+2] == 'u' &&
		d.data[start+3] == 'e'
}

func (d *deserializeState) isFalse(start, end int) bool {
	return end-start == 5 &&
		d.data[start] == 'f' &&
		d.data[start+1] == 'a' &&
		d.data[start+2] == 'l' &&
		d.data[start+3] == 's' &&
		d.data[start+4] == 'e'
}

func (d *deserializeState) isTrueInMicrosoftFormat(start, end int) bool {
	return end-start == 6 &&
		d.data[start] == '"' &&
		(d.data[start+1] == 't' || d.data[start+1] == 'T') &&
		d.data[start+2] == 'r' &&
		d.data[start+3] == 'u' &&
		d.data[start+4] == 'e' &&
		d.data[start+5] == '"'
}

func (d *deserializeState) isFalseInMicrosoftFormat(start, end int) bool {
	return end-start == 7 &&
		d.data[start] == '"' &&
		(d.data[start+1] == 'F' || d.data[start+1] == 'f') &&
		d.data[start+2] == 'a' &&
		d.data[start+3] == 'l' &&
		d.data[start+4] == 's' &&
		d.data[start+5] == 'e' &&
		d.data[start+6] == '"'
}

// Microsoft Azure Directory passes boolean values in as "True" and "False". In order to support this popular
// use case, we include a hack here temporarily for this issue only. Thanks to @plamenGo.
// See https://github.com/imulab/go-scim/pull/67
func (d *deserializeState) tryHackForMicrosoftADBooleanIssue(p Property, start, end int) (bool, error) {
	if strings.ToLower(p.Attr().path) != "active" {
		// We are only hacking for the "active" property for now.
		return false, nil
	}

	if d.isTrueInMicrosoftFormat(start, end) {
		return true, d.navigator.current().Set(true)
	} else if d.isFalseInMicrosoftFormat(start, end) {
		return true, d.navigator.current().Set(false)
	}

	panic("Microsoft may have fixed the boolean issue")
}

// scanWhile processes bytes in d.data[d.off:] until it
// receives a scan code not equal to op.
func (d *deserializeState) scanWhile(op int) {
	s, data, i := &d.scan, d.data, d.off
	for i < len(d.data) {
		newOp := s.Step()(s, data[i])
		i++
		if newOp != op {
			d.opCode = newOp
			d.off = i
			return
		}
	}

	d.off = len(d.data) + 1 // mark processed EOF with len+1
	d.opCode = d.scan.Eof()
}

// scanNext processed the next byte (as in d.data[d.off])
func (d *deserializeState) scanNext() {
	s, data, i := &d.scan, d.data, d.off
	if i < len(data) {
		d.opCode = s.Step()(s, data[i])
		d.off = i + 1
	} else {
		d.opCode = s.Eof()
		d.off = len(data) + 1 // mark processed EOF with len+1
	}
}

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func unquote(s []byte) (t string, ok bool) {
	s, ok = unquoteBytes(s)
	t = string(s)
	return
}

func unquoteBytes(s []byte) (t []byte, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	s = s[1 : len(s)-1]

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room? Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}
