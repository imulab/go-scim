package core

import "encoding/json"

// Returned describes the render access of the Attribute.
type Returned int

const (
	ReturnedDefault Returned = iota
	ReturnedAlways
	ReturnedRequest
	ReturnedNever
)

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
		panic("invalid returned-ability")
	}
}

func (r Returned) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}
