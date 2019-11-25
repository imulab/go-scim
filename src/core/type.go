package core

const (
	// SCIM data types defined in RFC7643
	TypeString Type = iota
	TypeInteger
	TypeDecimal
	TypeBoolean
	TypeDateTime
	TypeReference
	TypeBinary
	TypeComplex
)

// A SCIM data type
type Type int

// Return the standard string representation of this type
func (t Type) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInteger:
		return "integer"
	case TypeDecimal:
		return "decimal"
	case TypeBoolean:
		return "boolean"
	case TypeDateTime:
		return "dateTime"
	case TypeReference:
		return "reference"
	case TypeBinary:
		return "binary"
	case TypeComplex:
		return "complex"
	default:
		panic("invalid type")
	}
}
