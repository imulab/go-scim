package spec

import "encoding/json"

// SCIM uniqueness definition
type Uniqueness int

const (
	// SCIM uniqueness attribute defined in RFC7643
	UniquenessNone Uniqueness = iota
	UniquenessServer
	UniquenessGlobal
)

func mustParseUniqueness(value string) Uniqueness {
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

var (
	_ json.Marshaler = (Uniqueness)(0)
)
