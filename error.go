package scim

var (
	ErrInvalidFilter = &Error{code: "invalidFilter"}
	ErrTooMany       = &Error{code: "tooMany"}
	ErrUniqueness    = &Error{code: "uniqueness"}
	ErrMutability    = &Error{code: "mutability"}
	ErrInvalidSyntax = &Error{code: "invalidSyntax"}
	ErrInvalidPath   = &Error{code: "invalidPath"}
	ErrNoTarget      = &Error{code: "noTarget"}
	ErrInvalidValue  = &Error{code: "invalidValue"}
	ErrNotFound      = &Error{code: "notFound"}
	ErrSensitive     = &Error{code: "sensitive"}
	ErrConflict      = &Error{code: "conflict"}
	ErrInternal      = &Error{code: "internal"}
)

// Error is the standard SCIM error.
type Error struct {
	code string
}

func (e *Error) Error() string {
	return e.code
}
