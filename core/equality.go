package core

import (
	"strings"
	"time"
)

// An extension to the property interface to enhance the capability for
// properties that can perform equality operations. This interface is only
// capable of equality comparison; non-equality (i.e. 'ne' operator) should
// be derived from the equality result by taking the boolean negation.
type EqualAware interface {
	Property

	// Returns true if this property's value equals to the other value, with consideration of
	// case sensitivity where it applies. This call directly corresponds to the 'eq' operator.
	IsEqualTo(other interface{}) bool
}

func (s *stringProperty) IsEqualTo(other interface{}) bool {
	if _, ok := other.(string); !ok {
		return false
	}

	if s.v == nil {
		return false
	}

	v1, v2 := *(s.v), other.(string)
	if s.attr.CaseExact {
		return v1 == v2
	} else {
		return strings.ToLower(v1) == strings.ToLower(v2)
	}
}

func (i *integerProperty) IsEqualTo(other interface{}) bool {
	if i.v == nil {
		return false
	}

	var v1 interface{}
	{
		switch other.(type) {
		case int:
			v1 = int(*(i.v))
		case int32:
			v1 = int32(*(i.v))
		case int64:
			v1 = *(i.v)
		default:
			return false
		}
	}

	return v1 == other
}

func (d *decimalProperty) IsEqualTo(other interface{}) bool {
	if d.v == nil {
		return false
	}

	var v1 interface{}
	{
		switch other.(type) {
		case float32:
			v1 = float32(*(d.v))
		case float64:
			v1 = *(d.v)
		default:
			return false
		}
	}

	return v1 == other
}

func (b *booleanProperty) IsEqualTo(other interface{}) bool {
	if other == nil {
		return b.v == nil || !*(b.v)
	}

	if _, ok := other.(bool); !ok {
		return false
	}

	var v1 interface{}
	{
		if b.v == nil {
			v1 = false
		} else {
			v1 = *(b.v)
		}
	}

	return v1 == other
}

func (d *dateTimeProperty) IsEqualTo(other interface{}) bool {
	if _, ok := other.(string); !ok {
		return false
	}

	if d.v == nil {
		return false
	}

	var (
		v1  time.Time
		v2  time.Time
		err error
	)
	{
		v1, err = time.Parse(ISO8601, *(d.v))
		v2, err = time.Parse(ISO8601, other.(string))
		if err != nil {
			return false
		}
	}

	return v1.Equal(v2)
}

func (b *binaryProperty) IsEqualTo(other interface{}) bool {
	if _, ok := other.(string); !ok {
		return false
	}

	if b.v == nil {
		return false
	}

	// binary types are always caseExact
	return *(b.v) == other
}

func (r *referenceProperty) IsEqualTo(other interface{}) bool {
	if _, ok := other.(string); !ok {
		return false
	}

	if r.v == nil {
		return false
	}

	// reference types are always caseExact
	return *(r.v) == other
}

// Implements the case where non-complex multiValued property can use 'eq' operator to match
// its elements.
func (m *multiValuedProperty) IsEqualTo(other interface{}) bool {
	if m.attr.Type == TypeComplex {
		return false
	}

	for _, prop := range m.props {
		if eqProp, ok := prop.(EqualAware); !ok {
			continue
		} else {
			if eqProp.IsEqualTo(other) {
				return true
			}
		}
	}

	return false
}

// implementation checks
var (
	_ EqualAware = (*stringProperty)(nil)
	_ EqualAware = (*integerProperty)(nil)
	_ EqualAware = (*decimalProperty)(nil)
	_ EqualAware = (*booleanProperty)(nil)
	_ EqualAware = (*dateTimeProperty)(nil)
	_ EqualAware = (*referenceProperty)(nil)
	_ EqualAware = (*binaryProperty)(nil)
	_ EqualAware = (*multiValuedProperty)(nil)
)
