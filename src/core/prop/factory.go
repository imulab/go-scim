package prop

import "github.com/imulab/go-scim/src/core"

func NewProperty(attr *core.Attribute, parent core.Container) core.Property {
	if attr.MultiValued() {
		return NewMulti(attr, parent)
	} else {
		switch attr.Type() {
		case core.TypeString:
			return NewString(attr, parent)
		case core.TypeInteger:
			return NewInteger(attr, parent)
		case core.TypeDecimal:
			return NewDecimal(attr, parent)
		case core.TypeBoolean:
			return NewBoolean(attr, parent)
		case core.TypeDateTime:
			return NewDateTime(attr, parent)
		case core.TypeReference:
			return NewReference(attr, parent)
		case core.TypeBinary:
			return NewBinary(attr, parent)
		case core.TypeComplex:
			return NewComplex(attr, parent)
		default:
			panic("invalid type")
		}
	}
}
