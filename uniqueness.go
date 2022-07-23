package scim

import "encoding/json"

type Uniqueness int

const (
	UniquenessNone Uniqueness = iota
	UniquenessServer
	UniquenessGlobal
)

func (u Uniqueness) String() string {
	switch u {
	case UniquenessNone:
		return "none"
	case UniquenessServer:
		return "server"
	case UniquenessGlobal:
		return "global"
	default:
		panic("invalid Uniqueness")
	}
}

func (u Uniqueness) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}
