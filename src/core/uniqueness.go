package core

const (
	// SCIM uniqueness attribute defined in RFC7643
	UniquenessNone Uniqueness = iota
	UniquenessServer
	UniquenessGlobal
)
// SCIM uniqueness definition
type Uniqueness int

// Parse the given value to uniqueness type. Empty value is defaulted to 'none'.
// This function panics if the given value is invalid.
func MustParseUniqueness(value string) Uniqueness {
	switch value {
	case "none", "":
		return UniquenessNone
	case "server":
		return UniquenessServer
	case "global":
		return UniquenessGlobal
	default:
		panic("invalid uniqueness value")
	}
}

// Return a standard definition of this uniqueness.
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

func (u Uniqueness) MarshalJSON() ([]byte, error) {
	return []byte(u.String()), nil
}
