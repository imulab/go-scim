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