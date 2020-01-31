package json

import (
	"bytes"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Interface to implement to be able to serialize to JSON.
type Serializable interface {
	// MainSchemaId returns the id of the resource type's main schema that describes the target
	MainSchemaId() string
	// Visit implements the order for the visitor
	Visit(visitor prop.Visitor) error
}

// Serialize the given resource to JSON bytes. The serialization process subjects to the request attributes and
// excludedAttributes from options, and the SCIM return-ability rules.
func Serialize(serializable Serializable, options ...Options) ([]byte, error) {
	s := serializer{
		Buffer:   bytes.Buffer{},
		includes: []string{},
		excludes: []string{},
		stack:    []*frame{},
		scratch:  [64]byte{},
	}
	for _, opt := range options {
		opt.apply(&s, serializable)
	}

	if len(s.includes) > 0 && len(s.excludes) > 0 {
		return nil, fmt.Errorf("%w: attributes and excludedAttributes are mutually exclusive", spec.ErrInvalidValue)
	}

	if err := serializable.Visit(&s); err != nil {
		return nil, err
	}

	return s.Bytes(), nil
}

const (
	containerObject container = iota
	containerArray
)

type (
	// type of the containing property
	container int
	// stack frame during the traversal
	frame struct {
		// the type of the containing property
		container container
		// index of the element within the container
		index int
	}
	// json serializer state
	serializer struct {
		bytes.Buffer
		includes []string
		excludes []string
		stack    []*frame
		scratch  [64]byte
	}
)

func (s *serializer) ShouldVisit(property prop.Property) bool {
	attr := property.Attribute()

	// Write only properties are never returned. It is usually coupled
	// with returned=never, but we will check it to make sure.
	if attr.Mutability() == spec.MutabilityWriteOnly {
		return false
	}

	switch attr.Returned() {
	case spec.ReturnedAlways:
		return true
	case spec.ReturnedNever:
		return false
	case spec.ReturnedDefault:
		if len(s.includes) == 0 && len(s.excludes) == 0 {
			return !property.IsUnassigned()
		} else {
			test := strings.ToLower(property.Attribute().Path())
			if len(s.includes) > 0 {
				for _, include := range s.includes {
					if include == test || strings.HasPrefix(include, test+".") || strings.HasPrefix(test, include+".") {
						return !property.IsUnassigned()
					}
				}
				return false
			} else if len(s.excludes) > 0 {
				for _, exclude := range s.excludes {
					if exclude == test || strings.HasPrefix(test, exclude+".") {
						return false
					}
				}
				return !property.IsUnassigned()
			} else {
				panic("impossible: either includeFamily or excludeFamily")
			}
		}
	case spec.ReturnedRequest:
		if len(s.includes) > 0 {
			test := strings.ToLower(property.Attribute().Path())
			for _, include := range s.includes {
				if include == test || strings.HasPrefix(include, test+".") || strings.HasPrefix(test, include+".") {
					return true
				}
			}
			return false
		}
		return false
	default:
		panic("invalid returned-ability")
	}
}

func (s *serializer) Visit(property prop.Property) error {
	if s.current().index > 0 {
		_ = s.WriteByte(',')
	}

	if s.current().container != containerArray {
		s.appendPropertyName(property.Attribute())
	}

	if property.Attribute().MultiValued() || property.Attribute().Type() == spec.TypeComplex {
		return nil
	}

	if property.IsUnassigned() {
		s.appendNull()
		return nil
	}

	switch property.Attribute().Type() {
	case spec.TypeString, spec.TypeReference, spec.TypeDateTime, spec.TypeBinary:
		s.appendString(property.Raw().(string))
	case spec.TypeInteger:
		s.appendInteger(property.Raw().(int64))
	case spec.TypeDecimal:
		s.appendFloat(property.Raw().(float64))
	case spec.TypeBoolean:
		s.appendBoolean(property.Raw().(bool))
	default:
		panic("invalid type")
	}

	s.current().index++
	return nil
}

func (s *serializer) BeginChildren(container prop.Property) {
	switch {
	case container.Attribute().MultiValued():
		_ = s.WriteByte('[')
		s.push(containerArray)
	case container.Attribute().Type() == spec.TypeComplex:
		_ = s.WriteByte('{')
		s.push(containerObject)
	default:
		panic("unknown container")
	}
}

func (s *serializer) EndChildren(container prop.Property) {
	switch {
	case container.Attribute().MultiValued():
		_ = s.WriteByte(']')
	case container.Attribute().Type() == spec.TypeComplex:
		_ = s.WriteByte('}')
	default:
		panic("unknown container")
	}
	s.pop()
	if len(s.stack) > 0 {
		s.current().index++
	}
}

func (s *serializer) appendPropertyName(attribute *spec.Attribute) {
	_ = s.WriteByte('"')
	_, _ = s.WriteString(attribute.Name())
	_ = s.WriteByte('"')
	_ = s.WriteByte(':')
}

func (s *serializer) appendNull() {
	_, _ = s.WriteString("null")
}

func (s *serializer) appendString(value string) {
	_ = s.WriteByte('"')
	start := 0
	for i := 0; i < len(value); {
		if b := value[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] {
				i++
				continue
			}
			if start < i {
				_, _ = s.WriteString(value[start:i])
			}
			_ = s.WriteByte('\\')
			switch b {
			case '\\', '"':
				_ = s.WriteByte(b)
			case '\n':
				_ = s.WriteByte('n')
			case '\r':
				_ = s.WriteByte('r')
			case '\t':
				_ = s.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				_, _ = s.WriteString(`u00`)
				_ = s.WriteByte(hex[b>>4])
				_ = s.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(value[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				_, _ = s.WriteString(value[start:i])
			}
			_, _ = s.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				_, _ = s.WriteString(value[start:i])
			}
			_, _ = s.WriteString(`\u202`)
			_ = s.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(value) {
		_, _ = s.WriteString(value[start:])
	}
	_ = s.WriteByte('"')
}

func (s *serializer) appendInteger(value int64) {
	b := strconv.AppendInt(s.scratch[:0], value, 10)
	_, _ = s.Write(b)
}

func (s *serializer) appendFloat(value float64) {
	if math.IsInf(value, 0) || math.IsNaN(value) {
		panic(fmt.Errorf("%w: invalid decimal in json serialization", spec.ErrInvalidValue))
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	b := s.scratch[:0]
	abs := math.Abs(value)
	format := byte('f')
	if abs != 0 {
		if abs < 1e-6 || abs >= 1e21 {
			format = 'e'
		}
	}
	b = strconv.AppendFloat(b, value, format, -1, 64)
	if format == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	_, _ = s.Write(b)
}

func (s *serializer) appendBoolean(value bool) {
	if value {
		_, _ = s.WriteString("true")
	} else {
		_, _ = s.WriteString("false")
	}
}

func (s *serializer) push(c container) {
	s.stack = append(s.stack, &frame{
		container: c,
		index:     0,
	})
}

func (s *serializer) pop() {
	if len(s.stack) == 0 {
		panic("cannot pop on empty stack")
	}
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *serializer) current() *frame {
	if len(s.stack) == 0 {
		panic("stack is empty")
	}
	return s.stack[len(s.stack)-1]
}
