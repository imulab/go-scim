package json

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// Deserialize is the entry point of JSON deserialization. Unmarshal the JSON input bytes into a pre-prepared unassigned
// structure of Resource.
func Deserialize(json []byte, resource *prop.Resource) error {
	if err := checkValid(json, &scanner{}); err != nil {
		return err
	}

	state := &deserializeState{
		data:      json,
		off:       0,
		opCode:    scanContinue,
		scan:      scanner{},
		navigator: resource.Navigator(),
	}
	state.scan.reset()

	// skip the first few spaces
	state.scanWhile(scanSkipSpace)
	return state.parseComplexProperty(false)
}

// Entry point to deserialize a piece of JSON data into the given property. The JSON data is expected to be the content
// of a json.RawMessage parsed from the built-in encoding/json mechanism, hence, it should not contain any preceding
// spaces, and should a fragment of valid JSON.
//
// The allowElementForArray option is provided to allow JSON array element values be provided for a multiValued property
// so that it will be de-serialized as its element. The result will be a multiValued property containing a single element.
func DeserializeProperty(json []byte, property prop.Property, allowElementForArray bool) error {
	state := &deserializeState{
		data:      json,
		off:       0,
		opCode:    scanContinue,
		scan:      scanner{},
		navigator: prop.Navigate(property),
	}
	state.scan.reset()

	// Since this function is intended for bytes from json.RawMessage, it is not possible for it to precede with
	// spaces. Hence, simply use scanNext to read in the first byte, then use stateBeginValue to forcibly set the
	// state and op code. This is necessary since we are dealing with potentially just a fragment of valid JSON.
	state.scanNext()
	state.opCode = stateBeginValue(&state.scan, state.data[0])

	if !property.Attribute().MultiValued() {
		return state.parseSingleValuedProperty()
	} else {
		// Check the value is indeed a JSON array
		if state.data[0] == '[' {
			return state.parseMultiValuedProperty()
		}

		// We may choose to allow callers to provide value that corresponds to multiValue element
		// to be provided as a value for the multiValue property itself. If this feature is enabled,
		// we will parse the value as the multiValued element and add it to the multiValued container.
		if !allowElementForArray {
			return state.errInvalidSyntax("expects JSON array")
		}

		if mv, ok := state.navigator.Current().(interface {
			AppendElement() int
		}); !ok {
			return state.errInvalidSyntax("non-multiValued property at json array")
		} else {
			i := mv.AppendElement()
			if i < 0 {
				return fmt.Errorf("%w: failed to create property to host json array element", spec.ErrInternal)
			}
			state.navigator.At(i)
			defer state.navigator.Retract()
			if state.navigator.Error() != nil {
				return state.navigator.Error()
			}
			return state.parseSingleValuedProperty()
		}
	}
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
	scan      scanner
	navigator prop.Navigator
}

func (d *deserializeState) errInvalidSyntax(msg string, args ...interface{}) error {
	return fmt.Errorf("%w: %s (pos:%d)", spec.ErrInvalidSyntax, fmt.Sprintf(msg, args...), d.off)
}

// Parses the attribute/field name in a JSON object. This method expects a quoted string and skips through
// as much empty spaces and colon (appears as scanObjectKey) after it as possible.
func (d *deserializeState) parseFieldName() (string, error) {
	if d.opCode != scanBeginLiteral {
		return "", d.errInvalidSyntax("expects attribute name")
	}

	start := d.off - 1 // position of the first double quote
	d.scanWhile(scanContinue)
	end := d.off - 1 // position of the character after the second double quote

Skip:
	for {
		switch d.opCode {
		case scanObjectKey, scanSkipSpace:
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
	if d.opCode != scanBeginObject {
		if allowNull && d.opCode == scanBeginLiteral {
			return d.parseNull()
		}
		return d.errInvalidSyntax("expects a json object")
	}

	// skip any potential spaces between '{' and '"'
	d.scanWhile(scanSkipSpace)

kvs:
	for d.opCode != scanEndObject {
		// Focus on the property that corresponds to the field name
		var (
			p   prop.Property
			err error
		)
		{
			attrName, err := d.parseFieldName()
			if err != nil {
				return err
			}
			p = d.navigator.Dot(attrName).Current()
			if d.navigator.Error() != nil {
				return d.navigator.Error()
			}
		}

		// Parse field value
		if p.Attribute().MultiValued() {
			err = d.parseMultiValuedProperty()
		} else {
			err = d.parseSingleValuedProperty()
		}
		if err != nil {
			return err
		}

		// Exit focus on the field value property
		d.navigator.Retract()

		// Fast forward to the next field name/value pair, or exit the loop.
	fastForward:
		for {
			switch d.opCode {
			case scanEndObject:
				d.scanNext()
				break kvs
			case scanEnd:
				break kvs
			case scanSkipSpace, scanObjectValue, scanEndArray:
				d.scanNext()
			default:
				break fastForward
			}
		}
	}

	// Courtesy: skip any spaces between '}' and the next tokens
	if d.opCode == scanSkipSpace {
		d.scanWhile(scanSkipSpace)
	}

	return nil
}

// Delegate method to parse single valued field values. The caller must ensure that the currently focused property
// is indeed single valued.
func (d *deserializeState) parseSingleValuedProperty() error {
	switch d.navigator.Current().Attribute().Type() {
	case spec.TypeString, spec.TypeDateTime, spec.TypeBinary, spec.TypeReference:
		return d.parseStringProperty()
	case spec.TypeInteger:
		return d.parseIntegerProperty()
	case spec.TypeDecimal:
		return d.parseDecimalProperty()
	case spec.TypeBoolean:
		return d.parseBooleanProperty()
	case spec.TypeComplex:
		return d.parseComplexProperty(true)
	default:
		panic("invalid attribute type")
	}
}

// Parses a JSON array. This method expects '[' (appears as scanBeginArray) to be the current byte, or the literal
// null.
func (d *deserializeState) parseMultiValuedProperty() error {
	// Expect '[' or null.
	if d.opCode != scanBeginArray {
		if d.opCode == scanBeginLiteral {
			return d.parseNull()
		}
		return d.errInvalidSyntax("expects JSON array")
	}

	// Skip any spaces between '[' and the potential first element
	d.scanNext()
	if d.opCode == scanSkipSpace {
		d.scanWhile(scanSkipSpace)
	}

elements:
	for d.opCode != scanEndArray {
		// Create the place-holding element prototype and focus on it
		if mv, ok := d.navigator.Current().(interface {
			AppendElement() int
		}); !ok {
			return d.errInvalidSyntax("non-multiValued property at json array")
		} else {
			i := mv.AppendElement()
			if i < 0 {
				return fmt.Errorf("%w: failed to create property to host json array element", spec.ErrInternal)
			}
			d.navigator.At(i)
			if d.navigator.Error() != nil {
				return d.navigator.Error()
			}
		}

		// Parse the focused element property
		err := d.parseSingleValuedProperty()
		if err != nil {
			return err
		}

		// Exit the focus
		d.navigator.Retract()

		// Fast forward to the next element, or exit the loop.
	fastForward:
		for {
			switch d.opCode {
			case scanEndArray:
				d.scanNext()
				break elements
			case scanSkipSpace, scanArrayValue:
				d.scanNext()
			default:
				break fastForward
			}
		}
	}

	// Courtesy: skip any spaces between ']' and the next tokens
	if d.opCode == scanSkipSpace {
		d.scanWhile(scanSkipSpace)
	}

	return nil
}

// Parses a JSON string. This method expects a double quoted literal and the null literal.
func (d *deserializeState) parseStringProperty() error {
	p := d.navigator.Current()

	// check property type
	if p.Attribute().MultiValued() || (p.Attribute().Type() != spec.TypeString &&
		p.Attribute().Type() != spec.TypeDateTime &&
		p.Attribute().Type() != spec.TypeReference &&
		p.Attribute().Type() != spec.TypeBinary) {
		return d.errInvalidSyntax("expects string based property for '%s'", p.Attribute().Path())
	}

	// should start with literal
	if d.opCode != scanBeginLiteral {
		return d.errInvalidSyntax("expects json literal")
	}

	start := d.off - 1 // position of the first double quote
	d.scanWhile(scanContinue)
	end := d.off - 1 // position of the character after the second double quote

	if d.isNull(start, end) {
		if _, err := d.navigator.Current().Delete(); err != nil {
			return err
		}
		return nil
	}

	if d.data[start] != '"' || d.data[end-1] != '"' {
		return d.errInvalidSyntax("expects string literal value for '%s'", p.Attribute().Path())
	}

	v, ok := unquote(d.data[start:end])
	if !ok {
		return d.errInvalidSyntax("failed to unquote json string for '%s'", p.Attribute().Path())
	}

	if _, err := d.navigator.Current().Replace(v); err != nil {
		return err
	}

	return nil
}

// Parses a JSON integer. This method expects an integer literal and the null literal.
func (d *deserializeState) parseIntegerProperty() error {
	p := d.navigator.Current()

	// check property type
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeInteger {
		return d.errInvalidSyntax("expects integer property for '%s'", p.Attribute().Path())
	}

	// should start with literal
	if d.opCode != scanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(scanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if d.isNull(start, end) {
		if _, err := d.navigator.Current().Delete(); err != nil {
			return err
		}
		return nil
	}

	val, err := strconv.ParseInt(string(d.data[start:end]), 10, 64)
	if err != nil {
		return d.errInvalidSyntax("expects integer value")
	}

	if _, err := d.navigator.Current().Replace(val); err != nil {
		return err
	}

	return nil
}

// Parses a JSON boolean. This method expects the true, false, or null literal.
func (d *deserializeState) parseBooleanProperty() error {
	p := d.navigator.Current()

	// check property type
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeBoolean {
		return d.errInvalidSyntax("expects decimal property for '%s'", p.Attribute().Path())
	}

	// should start with literal
	if d.opCode != scanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(scanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if d.isNull(start, end) {
		if _, err := d.navigator.Current().Delete(); err != nil {
			return err
		}
		return nil
	}

	if d.isTrue(start, end) {
		if _, err := d.navigator.Current().Replace(true); err != nil {
			return err
		}
	} else if d.isFalse(start, end) {
		if _, err := d.navigator.Current().Replace(true); err != nil {
			return err
		}
	} else {
		return d.errInvalidSyntax("expects boolean value")
	}

	return nil
}

// Parses a JSON decimal. This method expects a decimal literal and the null literal.
func (d *deserializeState) parseDecimalProperty() error {
	p := d.navigator.Current()

	// check property type
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeDecimal {
		return d.errInvalidSyntax("expects decimal property for '%s'", p.Attribute().Path())
	}

	// should start with literal
	if d.opCode != scanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(scanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if d.isNull(start, end) {
		if _, err := d.navigator.Current().Delete(); err != nil {
			return err
		}
		return nil
	}

	val, err := strconv.ParseFloat(string(d.data[start:end]), 64)
	if err != nil {
		return d.errInvalidSyntax("expects decimal value")
	}

	if _, err := d.navigator.Current().Replace(val); err != nil {
		return err
	}

	return nil
}

// Parses the JSON null literal.
func (d *deserializeState) parseNull() error {
	// should start with literal
	if d.opCode != scanBeginLiteral {
		return d.errInvalidSyntax("expects property value")
	}

	start := d.off - 1 // position of the first character of the literal
	d.scanWhile(scanContinue)
	end := d.off - 1 // position of the character after the end of the literal

	if !d.isNull(start, end) {
		return d.errInvalidSyntax("expects null")
	}

	if _, err := d.navigator.Current().Delete(); err != nil {
		return err
	}

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

// scanWhile processes bytes in d.data[d.off:] until it
// receives a scan code not equal to op.
func (d *deserializeState) scanWhile(op int) {
	s, data, i := &d.scan, d.data, d.off
	for i < len(d.data) {
		newOp := s.step(s, data[i])
		i++
		if newOp != op {
			d.opCode = newOp
			d.off = i
			return
		}
	}

	d.off = len(d.data) + 1 // mark processed EOF with len+1
	d.opCode = d.scan.eof()
}

// scanNext processed the next byte (as in d.data[d.off])
func (d *deserializeState) scanNext() {
	s, data, i := &d.scan, d.data, d.off
	if i < len(data) {
		d.opCode = s.step(s, data[i])
		d.off = i + 1
	} else {
		d.opCode = s.eof()
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
