package spec

// SCIM mutability definition
type Mutability int

// SCIM mutability attribute defined in RFC7643
const (
	MutabilityReadWrite Mutability = iota
	MutabilityReadOnly
	MutabilityWriteOnly
	MutabilityImmutable
)

func mustParseMutability(value string) Mutability {
	switch value {
	case "readWrite", "":
		return MutabilityReadWrite
	case "readOnly":
		return MutabilityReadOnly
	case "immutable":
		return MutabilityImmutable
	case "writeOnly":
		return MutabilityWriteOnly
	default:
		panic("invalid mutability value")
	}
}

func (m Mutability) String() string {
	switch m {
	case MutabilityReadWrite:
		return "readWrite"
	case MutabilityReadOnly:
		return "readOnly"
	case MutabilityImmutable:
		return "immutable"
	case MutabilityWriteOnly:
		return "writeOnly"
	default:
		panic("invalid mutability")
	}
}
