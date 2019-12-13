package spec

import "encoding/json"

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

// Parse the given value to type. Empty value is defaulted to 'string'.
// This function panics if the given value is invalid.
func MustParseType(value string) Type {
	switch value {
	case "string", "":
		return TypeString
	case "integer":
		return TypeInteger
	case "decimal":
		return TypeDecimal
	case "boolean":
		return TypeBoolean
	case "dateTime":
		return TypeDateTime
	case "reference":
		return TypeReference
	case "binary":
		return TypeBinary
	case "complex":
		return TypeComplex
	default:
		panic("invalid type value")
	}
}

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

func (t Type) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

var (
	_ json.Marshaler = (Type)(0)
)