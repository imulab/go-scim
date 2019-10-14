package core

// SCIM data types defined in RFC7643
const (
	TypeString    = "string"
	TypeInteger   = "integer"
	TypeDecimal   = "decimal"
	TypeBoolean   = "boolean"
	TypeDateTime  = "datetime"
	TypeReference = "reference"
	TypeBinary    = "binary"
	TypeComplex   = "complex"
)

// SCIM mutability attribute defined in RFC7643
const (
	MutabilityReadWrite = "readWrite"
	MutabilityReadOnly  = "readOnly"
	MutabilityWriteOnly = "writeOnly"
	MutabilityImmutable = "immutable"
)

// SCIM returned attribute defined in RFC7643
const (
	ReturnedDefault = "default"
	ReturnedAlways  = "always"
	ReturnedRequest = "request"
	ReturnedNever   = "never"
)

// SCIM uniqueness attribute defined in RFC7643
const (
	UniquenessNone   = "none"
	UniquenessServer = "server"
	UniquenessGlobal = "global"
)

// Datetime format
const ISO8601 = "2006-01-02T15:04:05"

// query tokens
const (
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
