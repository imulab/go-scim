package spec

import "encoding/json"

// SCIM returned definition
type Returned int

const (
	// SCIM returned attribute defined in RFC7643
	ReturnedDefault Returned = iota
	ReturnedAlways
	ReturnedRequest
	ReturnedNever
)

func mustParseReturned(value string) Returned {
	switch value {
	case "default", "":
		return ReturnedDefault
	case "always":
		return ReturnedAlways
	case "never":
		return ReturnedNever
	case "request":
		return ReturnedRequest
	default:
		panic("invalid returned value")
	}
}

func (r Returned) String() string {
	switch r {
	case ReturnedDefault:
		return "default"
	case ReturnedAlways:
		return "always"
	case ReturnedNever:
		return "never"
	case ReturnedRequest:
		return "request"
	default:
		panic("invalid return-ability")
	}
}

func (r Returned) MarshalJSON() ([]byte, error) {
	return []byte(r.String()), nil
}

var (
	_ json.Marshaler = (Returned)(0)
)
