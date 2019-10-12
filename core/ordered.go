package core

import (
	"strings"
	"time"
)

// An extension to the property interface to enhance the capability for
// properties that can perform order comparisons.
type OrderAware interface {
	EqualAware

	// Returns true if this property's value is greater than the other value, with consideration of
	// case sensitivity where it applies. This call directly corresponds to the 'gt' operator.
	IsGreaterThan(other interface{}) bool

	// Returns true if this property's value is less than the other value, with consideration of
	// case sensitivity where it applies. This call directly corresponds to the 'lt' operator.
	IsLessThan(other interface{}) bool

	// Returns true if this property's value is greater than or equal to the other value, with consideration of
	// case sensitivity where it applies. This call directly corresponds to the 'ge' operator.
	IsGreaterThanOrEqualTo(other interface{}) bool

	// Returns true if this property's value is less than or equal to the other value, with consideration of
	// case sensitivity where it applies. This call directly corresponds to the 'le' operator.
	IsLessThanOrEqualTo(other interface{}) bool
}

func (s *stringProperty) IsGreaterThan(other interface{}) bool {
	r, ok := s.compareWith(other)
	return ok && r > 0
}

func (s *stringProperty) IsLessThan(other interface{}) bool {
	r, ok := s.compareWith(other)
	return ok && r < 0
}

func (s *stringProperty) IsGreaterThanOrEqualTo(other interface{}) bool {
	r, ok := s.compareWith(other)
	return ok && r >= 0
}

func (s *stringProperty) IsLessThanOrEqualTo(other interface{}) bool {
	r, ok := s.compareWith(other)
	return ok && r <= 0
}

func (s *stringProperty) compareWith(other interface{}) (r int, ok bool) {
	_, ok = other.(string)
	if !ok {
		return
	}

	ok = s.v != nil
	if !ok {
		return
	}

	v1, v2 := *(s.v), other.(string)
	if s.attr.CaseExact {
		r = strings.Compare(v1, v2)
	} else {
		r = strings.Compare(strings.ToLower(v1), strings.ToLower(v2))
	}

	return
}

func (i *integerProperty) IsGreaterThan(other interface{}) bool {
	r, ok := i.compareWith(other)
	return ok && r > 0
}

func (i *integerProperty) IsLessThan(other interface{}) bool {
	r, ok := i.compareWith(other)
	return ok && r < 0
}

func (i *integerProperty) IsGreaterThanOrEqualTo(other interface{}) bool {
	r, ok := i.compareWith(other)
	return ok && r >= 0
}

func (i *integerProperty) IsLessThanOrEqualTo(other interface{}) bool {
	r, ok := i.compareWith(other)
	return ok && r <= 0
}

func (i *integerProperty) compareWith(other interface{}) (r int, ok bool) {
	ok = i.v != nil
	if !ok {
		return
	}

	var (
		v1 = *(i.v)
		v2 int64
	)
	{
		switch other.(type) {
		case int:
			v2 = int64(other.(int))
		case int32:
			v2 = int64(other.(int32))
		case int64:
			v2 = other.(int64)
		default:
			ok = false
			return
		}
	}

	if v1 == v2 {
		r = 0
	} else if v1 > v2 {
		r = 1
	} else {
		r = -1
	}

	return
}

func (d *decimalProperty) IsGreaterThan(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r > 0
}

func (d *decimalProperty) IsLessThan(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r < 0
}

func (d *decimalProperty) IsGreaterThanOrEqualTo(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r >= 0
}

func (d *decimalProperty) IsLessThanOrEqualTo(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r <= 0
}

func (d *decimalProperty) compareWith(other interface{}) (r int, ok bool) {
	ok = d.v != nil
	if !ok {
		return
	}

	var (
		v1 = *(d.v)
		v2 float64
	)
	{
		switch other.(type) {
		case float32:
			v2 = float64(other.(float32))
		case float64:
			v2 = other.(float64)
		default:
			ok = false
			return
		}
	}

	if v1 == v2 {
		r = 0
	} else if v1 > v2 {
		r = 1
	} else {
		r = -1
	}

	return
}

func (d *dateTimeProperty) IsGreaterThan(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r > 0
}

func (d *dateTimeProperty) IsLessThan(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r < 0
}

func (d *dateTimeProperty) IsGreaterThanOrEqualTo(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r >= 0
}

func (d *dateTimeProperty) IsLessThanOrEqualTo(other interface{}) bool {
	r, ok := d.compareWith(other)
	return ok && r <= 0
}

func (d *dateTimeProperty) compareWith(other interface{}) (r int, ok bool) {
	ok = d.v != nil
	if !ok {
		return
	}

	_, ok = other.(string)
	if !ok {
		return
	}

	var (
		t1 	time.Time
		t2 	time.Time
		err error
	)
	{
		t1, err = time.Parse(ISO8601, *(d.v))
		t2, err = time.Parse(ISO8601, other.(string))
		ok = err == nil
		if !ok {
			return
		}
	}

	if t1.Equal(t2) {
		r = 0
	} else if t1.After(t2) {
		r = 1
	} else {
		r = -1
	}

	return
}

// implementation checks
var (
	_ OrderAware = (*stringProperty)(nil)
	_ OrderAware = (*integerProperty)(nil)
	_ OrderAware = (*decimalProperty)(nil)
	_ OrderAware = (*dateTimeProperty)(nil)
)