package spec

import "encoding/json"

const (
	// SCIM returned attribute defined in RFC7643
	ReturnedDefault Returned = iota
	ReturnedAlways
	ReturnedRequest
	ReturnedNever
)

// SCIM returned definition
type Returned int

// Parse the given value to returned type. Empty value is defaulted to 'default'.
// This function panics if the given value is invalid.
func MustParseReturned(value string) Returned {
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

// Return a standard definition of this return-ability
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