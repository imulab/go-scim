package spec

// A SCIM data type
type Type int

// SCIM data types defined in RFC7643
const (
	TypeString Type = iota
	TypeInteger
	TypeDecimal
	TypeBoolean
	TypeDateTime
	TypeReference
	TypeBinary
	TypeComplex
)

func mustParseType(value string) Type {
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
