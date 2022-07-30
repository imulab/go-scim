package core

import "encoding/json"

// Mutability describes read/write access of the Attribute.
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
	return json.Marshal(m.String())
}
