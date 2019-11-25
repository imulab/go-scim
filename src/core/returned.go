package core

const (
	// SCIM returned attribute defined in RFC7643
	ReturnedDefault Returned = iota
	ReturnedAlways
	ReturnedRequest
	ReturnedNever
)

// SCIM returned definition
type Returned int

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