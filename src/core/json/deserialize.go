package json

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/prop"
	"strconv"
)

// Entry point of JSON deserialization. Unmarshal the JSON input bytes into the unassigned
// structure of resource.
func Deserialize(json []byte, resource *prop.Resource) error {
	if err := checkValid(json, &scanner{}); err != nil {
		return err
	}

	state := &deserializeState{
		data:      json,
		off:       0,
		opCode:    scanContinue,
		scan:      scanner{},
		navigator: resource.NewNavigator(),
	}
	state.scan.reset()

	// skip the first few spaces
	state.scanWhile(scanSkipSpace)
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
	scan      scanner
	navigator *prop.Navigator
}

func (d *deserializeState) errInvalidSyntax(msg string, args ...interface{}) error {
	return errors.InvalidSyntax("failed to parse json: "+msg+" (idx: %d)", append(args, d.off)...)
}

func (d *deserializeState) errInvalidValue(msg string, args ...interface{}) error {
	return errors.InvalidValue("failed to parse json: "+msg+" (idx: %d)", append(args, d.off)...)
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
			p   core.Property
			err error
		)
		{
			attrName, err := d.parseFieldName()
			if err != nil {
				return err
			}
			p, err = d.navigator.FocusName(attrName)
			if err != nil {
				return err
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
			case scanSkipSpace, scanObjectValue:
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
	case core.TypeString, core.TypeDateTime, core.TypeBinary, core.TypeReference:
		return d.parseStringProperty()
	case core.TypeInteger:
		return d.parseIntegerProperty()
	case core.TypeDecimal:
		return d.parseDecimalProperty()
	case core.TypeBoolean:
		return d.parseBooleanProperty()
	case core.TypeComplex:
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
	d.scanWhile(scanSkipSpace)

elements:
	for d.opCode != scanEndArray {
		// Create the place-holding element prototype and focus on it
		i := d.navigator.Current().(core.Container).NewChild()
		_, err := d.navigator.FocusIndex(i)
		if err != nil {
			return err
		}

		// Parse the focused element property
		err = d.parseSingleValuedProperty()
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
	if !(p.Attribute().SingleValued() && (p.Attribute().Type() == core.TypeString ||
		p.Attribute().Type() == core.TypeDateTime ||
		p.Attribute().Type() == core.TypeReference ||
		p.Attribute().Type() == core.TypeBinary)) {
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
		return d.navigator.Current().Delete()
	}

	if d.data[start] != '"' || d.data[end-1] != '"' {
		return d.errInvalidSyntax("expects string literal value for '%s'", p.Attribute().Path())
	}

	return d.navigator.Current().Replace(string(d.data[start+1 : end-1]))
}

// Parses a JSON integer. This method expects an integer literal and the null literal.
func (d *deserializeState) parseIntegerProperty() error {
	p := d.navigator.Current()

	// check property type
	if !(p.Attribute().SingleValued() && p.Attribute().Type() == core.TypeInteger) {
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
		return d.navigator.Current().Delete()
	}

	val, err := strconv.ParseInt(string(d.data[start:end]), 10, 64)
	if err != nil {
		return errors.InvalidValue("expects integer value")
	}

	return d.navigator.Current().Replace(val)
}

// Parses a JSON boolean. This method expects the true, false, or null literal.
func (d *deserializeState) parseBooleanProperty() error {
	p := d.navigator.Current()

	// check property type
	if !(p.Attribute().SingleValued() && p.Attribute().Type() == core.TypeBoolean) {
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
		return d.navigator.Current().Delete()
	}

	if d.isTrue(start, end) {
		return d.navigator.Current().Replace(true)
	} else if d.isFalse(start, end) {
		return d.navigator.Current().Replace(false)
	} else {
		return d.errInvalidValue("expects boolean value")
	}
}

// Parses a JSON decimal. This method expects a decimal literal and the null literal.
func (d *deserializeState) parseDecimalProperty() error {
	p := d.navigator.Current()

	// check property type
	if !(p.Attribute().SingleValued() && p.Attribute().Type() == core.TypeDecimal) {
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
		return d.navigator.Current().Delete()
	}

	val, err := strconv.ParseFloat(string(d.data[start:end]), 64)
	if err != nil {
		return errors.InvalidValue("expects decimal value")
	}

	return d.navigator.Current().Replace(val)
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

	return d.navigator.Current().Delete()
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
