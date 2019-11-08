package json

import (
	"bytes"
	"fmt"
	"github.com/imulab/go-scim/core"
	"math"
	"strconv"
)

// Entry point to serialize resources to JSON.
func Serialize(resource *core.Resource, includedAttributes []string, excludedAttributes []string, ) ([]byte, error) {
	s := new(serializer)
	s.includedAttributes = includedAttributes
	s.excludedAttributes = excludedAttributes
	s.elementIndexes = make([]int, 0)
	s.contexts = make([]int, 0)

	err := resource.Visit(s)
	if err != nil {
		return nil, err
	}

	return s.Bytes(), nil
}

const (
	contextObject = iota
	contextArray
)

// state for serializing resources to JSON.
type serializer struct {
	// fields that were requested to be included in the response
	includedAttributes []string
	// fields that were requested to be excluded in the response
	excludedAttributes []string
	// buffer
	bytes.Buffer
	// scratch buffer to assist in number conversions
	scratch [64]byte
	// stack to keep track of the index of the current element among its context
	elementIndexes []int
	// stack to keep track of the contexts
	contexts []int
}

func (v *serializer) ShouldVisit(property core.Property) bool {
	switch property.Attribute().Returned {
	case core.ReturnedAlways:
		return true
	case core.ReturnedNever:
		return false
	case core.ReturnedRequest:
		for _, included := range v.includedAttributes {
			if property.Attribute().GoesBy(included) {
				return true
			}
		}
		return false
	case core.ReturnedDefault:
		for _, excluded := range v.excludedAttributes {
			if property.Attribute().GoesBy(excluded) {
				return false
			}
		}
		return !property.IsUnassigned()
	default:
		panic("illegal returned-ability value")
	}
}

func (v *serializer) Visit(property core.Property) error {
	if v.currentIndex() > 0 {
		v.WriteByte(',')
	}

	if v.currentContext() != contextArray {
		v.encodeName(property)
	}

	attr := property.Attribute()
	if attr.MultiValued || attr.Type == core.TypeComplex {
		return nil
	}

	switch attr.Type {
	case core.TypeString, core.TypeReference, core.TypeBinary, core.TypeDateTime:
		v.encodeString(property)
	case core.TypeInteger:
		v.encodeInteger(property)
	case core.TypeDecimal:
		if err := v.encodeFloat(property); err != nil {
			return err
		}
	case core.TypeBoolean:
		v.encodeBoolean(property)
	default:
		panic("invalid property type")
	}

	v.incrementIndex()
	return nil
}

func (v *serializer) encodeName(property core.Property) {
	v.WriteByte('"')
	v.WriteString(property.Attribute().Name)
	v.WriteByte('"')
	v.WriteByte(':')
}

func (v *serializer) encodeString(property core.Property) {
	val := property.Raw()

	if val == nil {
		v.WriteString("null")
		return
	}

	v.WriteByte('"')
	v.WriteString(val.(string))
	v.WriteByte('"')
}

func (v *serializer) encodeInteger(property core.Property) {
	val := property.Raw()

	if val == nil {
		v.WriteString("null")
		return
	}

	b := strconv.AppendInt(v.scratch[:0], val.(int64), 10)
	v.Write(b)
}

func (v *serializer) encodeBoolean(property core.Property) {
	val := property.Raw()

	if val == nil {
		v.WriteString("null")
		return
	}

	if val.(bool) {
		v.WriteString("true")
	} else {
		v.WriteString("false")
	}
}

func (v *serializer) encodeFloat(property core.Property) error {
	val := property.Raw()

	if val == nil {
		v.WriteString("null")
		return nil
	}

	f := val.(float64)
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return core.Errors.InvalidValue(fmt.Sprintf(
			"invalid value during json serializaiton: value for attribute '%s' is not a proper decimal number.",
			property.Attribute().DisplayName(),
		))
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	b := v.scratch[:0]
	abs := math.Abs(f)
	format := byte('f')
	if abs != 0 {
		if abs < 1e-6 || abs >= 1e21 {
			format = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, format, -1, 64)
	if format == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}

	v.Write(b)
	return nil
}

func (v *serializer) BeginComplex() {
	v.WriteByte('{')
	v.pushContext(contextObject)
	v.pushIndex()
}

func (v *serializer) EndComplex() {
	v.WriteByte('}')
	v.popContext()
	v.popIndex()
	v.incrementIndex()
}

func (v *serializer) BeginMulti() {
	v.WriteByte('[')
	v.pushContext(contextArray)
	v.pushIndex()
}

func (v *serializer) EndMulti() {
	v.WriteByte(']')
	v.popIndex()
	v.popContext()
	v.incrementIndex()
}

func (v *serializer) pushIndex() {
	v.elementIndexes = append(v.elementIndexes, 0)
}

func (v *serializer) popIndex() {
	v.elementIndexes = v.elementIndexes[:len(v.elementIndexes)-1]
}

func (v *serializer) currentIndex() int {
	return v.elementIndexes[len(v.elementIndexes)-1]
}

func (v *serializer) incrementIndex() {
	if len(v.elementIndexes) > 0 {
		v.elementIndexes[len(v.elementIndexes)-1]++
	}
}

func (v *serializer) pushContext(ctx int) {
	v.contexts = append(v.contexts, ctx)
}

func (v *serializer) popContext() {
	v.contexts = v.contexts[:len(v.contexts)-1]
}

func (v *serializer) currentContext() int {
	return v.contexts[len(v.contexts)-1]
}