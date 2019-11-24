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

	// Return true if the property matches another property. Two properties match if and only if
	// the attributes match, and the value that determines equality match each other. For simple
	// properties, the value that determines equality is itself; for complex properties, the value
	// that determines equality is its sub attributes marked as identity attributes; for multiValued
	// properties, two property matches if and only if all their properties match.
	Matches(another Property) bool
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

func (s *stringProperty) Matches(another Property) bool {
	if !s.attr.Equals(another.Attribute()) {
		return false
	}

	if s.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		return s.IsEqualTo(another.Raw())
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

func (i *integerProperty) Matches(another Property) bool {
	if !i.attr.Equals(another.Attribute()) {
		return false
	}

	if i.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		return i.IsEqualTo(another.Raw())
	}
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

func (d *decimalProperty) Matches(another Property) bool {
	if !d.attr.Equals(another.Attribute()) {
		return false
	}

	if d.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		return d.IsEqualTo(another.Raw())
	}
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

func (b *booleanProperty) Matches(another Property) bool {
	if !b.attr.Equals(another.Attribute()) {
		return false
	}

	if b.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		return b.IsEqualTo(another.Raw())
	}
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

func (d *dateTimeProperty) Matches(another Property) bool {
	if !d.attr.Equals(another.Attribute()) {
		return false
	}

	if d.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		return d.IsEqualTo(another.Raw())
	}
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

func (b *binaryProperty) Matches(another Property) bool {
	if !b.attr.Equals(another.Attribute()) {
		return false
	}

	if b.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		return b.IsEqualTo(another.Raw())
	}
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

func (r *referenceProperty) Matches(another Property) bool {
	if !r.attr.Equals(another.Attribute()) {
		return false
	}

	if r.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		return r.IsEqualTo(another.Raw())
	}
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

func (m *multiValuedProperty) Matches(another Property) bool {
	if !m.attr.Equals(another.Attribute()) {
		return false
	}

	if m.IsUnassigned() {
		return another.IsUnassigned()
	} else {
		m1 := another.(*multiValuedProperty)
		if len(m.props) != len(m1.props) {
			return false
		}

		// short circuit: bet element order has not changed, give a chance
		// for the comparison to run in O(N).
		for i, elem := range m.props {
			if !elem.(EqualAware).Matches(m1.props[i]) {
				goto SlowCheck
			}
		}
		return true

	SlowCheck:
		// have to resort to O(N^2) check now.
		for _, e1 := range m.props {
			for _, e2 := range m1.props {
				if e1.(EqualAware).Matches(e2) {
					continue
				}
			}
			return false
		}
		return true
	}
}

func (c *complexProperty) IsEqualTo(other interface{}) bool {
	// complex properties do not participate in 'eq' operations
	return false
}

func (c *complexProperty) Matches(another Property) bool {
	if !c.attr.Equals(another.Attribute()) {
		return false
	}

	identityProps := c.getIdentityProps()
	for _, idProp := range identityProps {
		p2, err := another.(*complexProperty).getSubProperty(idProp.Attribute().Name)
		if err != nil {
			return false
		}

		if !idProp.(EqualAware).Matches(p2) {
			return false
		}
	}

	return true
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
	_ EqualAware = (*complexProperty)(nil)
)
