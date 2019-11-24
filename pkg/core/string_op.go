package core

import "strings"

// An extension to the property interface to enhance the property to perform
// a series of string operations
type StringOpCapable interface {
	Property

	// Return true if this property's value starts with the other value, with consideration of
	// case sensitivity, if applicable.
	StartsWith(other interface{}) bool

	// Return true if this property's value ends with the other value, with consideration of
	// case sensitivity, if applicable.
	EndsWith(other interface{}) bool

	// Return true if this property's value contains the other value, with consideration of
	// case sensitivity, if applicable.
	Contains(other interface{}) bool
}

func (s *stringProperty) StartsWith(other interface{}) bool {
	v1, v2, ok := s.formatStringOp(other)
	return ok && strings.HasPrefix(v1, v2)
}

func (s *stringProperty) EndsWith(other interface{}) bool {
	v1, v2, ok := s.formatStringOp(other)
	return ok && strings.HasSuffix(v1, v2)
}

func (s *stringProperty) Contains(other interface{}) bool {
	v1, v2, ok := s.formatStringOp(other)
	return ok && strings.Contains(v1, v2)
}

func (s *stringProperty) formatStringOp(other interface{}) (v1 string, v2 string, ok bool) {
	ok = s.v != nil
	if !ok {
		return
	}

	_, ok = other.(string)
	if !ok {
		return
	}

	v1 = *(s.v)
	v2 = other.(string)

	if !s.attr.CaseExact {
		v1 = strings.ToLower(v1)
		v2 = strings.ToLower(v2)
	}

	return
}

func (r *referenceProperty) StartsWith(other interface{}) bool {
	v1, v2, ok := r.formatStringOp(other)
	return ok && strings.HasPrefix(v1, v2)
}

func (r *referenceProperty) EndsWith(other interface{}) bool {
	v1, v2, ok := r.formatStringOp(other)
	return ok && strings.HasSuffix(v1, v2)
}

func (r *referenceProperty) Contains(other interface{}) bool {
	v1, v2, ok := r.formatStringOp(other)
	return ok && strings.Contains(v1, v2)
}

func (r *referenceProperty) formatStringOp(other interface{}) (v1 string, v2 string, ok bool) {
	ok = r.v != nil
	if !ok {
		return
	}

	_, ok = other.(string)
	if !ok {
		return
	}

	// reference property is always caseExact
	v1 = *(r.v)
	v2 = other.(string)
	return
}

// implementation checks
var (
	_ StringOpCapable = (*stringProperty)(nil)
	_ StringOpCapable = (*referenceProperty)(nil)
)
