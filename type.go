package scim

import "encoding/json"

type Type int

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
		panic("invalid attribute type")
	}
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
