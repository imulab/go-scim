package core

const (
	// SCIM mutability attribute defined in RFC7643
	MutabilityReadWrite Mutability = iota
	MutabilityReadOnly
	MutabilityWriteOnly
	MutabilityImmutable
)

// SCIM mutability definition
type Mutability int

// Parse the given value to mutability type. Empty value is defaulted to readWrite.
// This function panics if the given value is invalid.
func MustParseMutability(value string) Mutability {
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

// Return the standard string representation of this mutability
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

func (m Mutability) MarshalJSON() ([]byte, error) {
	return []byte(m.String()), nil
}
