package prop

import (
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// NewProperty creates a new property of any legal SCIM type
func NewProperty(attr *spec.Attribute) Property {
	if attr.MultiValued() {
		return NewMulti(attr)
	}

	switch attr.Type() {
	case spec.TypeString:
		return NewString(attr)
	case spec.TypeInteger:
		return NewInteger(attr)
	case spec.TypeDecimal:
		return NewDecimal(attr)
	case spec.TypeBoolean:
		return NewBoolean(attr)
	case spec.TypeDateTime:
		return NewDateTime(attr)
	case spec.TypeReference:
		return NewReference(attr)
	case spec.TypeBinary:
		return NewBinary(attr)
	case spec.TypeComplex:
		return NewComplex(attr)
	default:
		panic("invalid type")
	}
}
