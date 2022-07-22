package scim

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
		panic("invalid Returned-ability")
	}
}

func (r Returned) MarshalJSON() ([]byte, error) {
	return []byte(r.String()), nil
}
