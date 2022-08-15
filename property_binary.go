package scim

import (
	"bytes"
	"encoding/base64"
)

// TODO complete the rest of properties
// TODO remove simpleProperty
// TODO evaluator
// TODO traverseQualifiedElements
// TODO crud methods on resources

type binaryProperty struct {
	*stringProperty
}

func (p *binaryProperty) Set(value any) error {
	s, ok := value.(string)
	if !ok {
		return ErrInvalidValue
	}

	if _, err := base64.StdEncoding.DecodeString(s); err != nil {
		return ErrInvalidValue
	}

	return p.stringProperty.Set(value)
}

func (p *binaryProperty) equalsTo(value any) bool {
	if !p.Unassigned() || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	if b1, err := base64.StdEncoding.DecodeString(*p.vs); err != nil {
		return false
	} else if b2, err := base64.StdEncoding.DecodeString(s); err != nil {
		return false
	} else {
		return bytes.Compare(b1, b2) == 0
	}
}
