package core

type (
	Type int
	Mutability int
	Returned int
	Uniqueness int
)

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

const (
	// SCIM mutability attribute defined in RFC7643
	MutabilityReadWrite Mutability = iota
	MutabilityReadOnly
	MutabilityWriteOnly
	MutabilityImmutable
)

const (
	// SCIM returned attribute defined in RFC7643
	ReturnedDefault Returned = iota
	ReturnedAlways
	ReturnedRequest
	ReturnedNever
)

const (
	// SCIM uniqueness attribute defined in RFC7643
	UniquenessNone Uniqueness = iota
	UniquenessServer
	UniquenessGlobal
)

const (
	// Datetime format
	ISO8601 = "2006-01-02T15:04:05"

	// Query tokens
	LeftParen  = "("
	RightParen = ")"
	And        = "and"
	Or         = "or"
	Not        = "not"
	Eq         = "eq"
	Ne         = "ne"
	Sw         = "sw"
	Ew         = "ew"
	Co         = "co"
	Pr         = "pr"
	Gt         = "gt"
	Ge         = "ge"
	Lt         = "lt"
	Le         = "le"
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
		panic("invalid type")
	}
}

func (m Mutability) String() string {
	switch m {
	case MutabilityReadWrite:
		return "readWrite"
	case MutabilityReadOnly:
		return "readOnly"
	case MutabilityWriteOnly:
		return "writeOnly"
	case MutabilityImmutable:
		return "immutable"
	default:
		panic("invalid mutability")
	}
}

func (r Returned) String() string {
	switch r {
	case ReturnedDefault:
		return "default"
	case ReturnedAlways:
		return "always"
	case ReturnedRequest:
		return "request"
	case ReturnedNever:
		return "never"
	default:
		panic("invalid returned")
	}
}

func (u Uniqueness) String() string {
	switch u {
	case UniquenessNone:
		return "none"
	case UniquenessServer:
		return "server"
	case UniquenessGlobal:
		return "global"
	default:
		panic("invalid uniqueness")
	}
}

func NewType(v string) Type {
	switch v {
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
		panic("invalid type")
	}
}

func NewMutability(v string) Mutability {
	switch v {
	case "readWrite", "":
		return MutabilityReadWrite
	case "readOnly":
		return MutabilityReadOnly
	case "writeOnly":
		return MutabilityWriteOnly
	case "immutable":
		return MutabilityImmutable
	default:
		panic("invalid mutability")
	}
}

func NewReturned(v string) Returned {
	switch v {
	case "default", "":
		return ReturnedDefault
	case "always":
		return ReturnedAlways
	case "request":
		return ReturnedRequest
	case "never":
		return ReturnedNever
	default:
		panic("invalid returned")
	}
}

func NewUniqueness(v string) Uniqueness {
	switch v {
	case "none", "":
		return UniquenessNone
	case "server":
		return UniquenessServer
	case "global":
		return UniquenessGlobal
	default:
		panic("invalid uniqueness")
	}
}