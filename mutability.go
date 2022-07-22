package scim

type Mutability int

const (
	MutabilityReadWrite Mutability = iota
	MutabilityReadOnly
	MutabilityWriteOnly
	MutabilityImmutable
)

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
		panic("invalid Mutability")
	}
}

func (m Mutability) MarshalJSON() ([]byte, error) {
	return []byte(m.String()), nil
}
