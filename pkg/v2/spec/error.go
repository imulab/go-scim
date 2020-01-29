package spec

// Error prototypes
var (
	// The specified filter syntax was invalid, or the specified attribute and filter comparison combination is not supported.
	ErrInvalidFilter = &Error{Status: 400, Type: "invalidFilter"}

	// The specified filter yields many more results than the server is willing to calculate or process.
	ErrTooMany = &Error{Status: 400, Type: "tooMany"}

	// One or more of the attribute values are already in use or are reserved.
	ErrUniqueness = &Error{Status: 400, Type: "uniqueness"}

	// The attempted modification is not compatible with the target attribute's mutability or current state (e.g.,
	// modification of an "immutable" attribute with an existing value).
	ErrMutability = &Error{Status: 400, Type: "mutability"}

	// The request body message structure was invalid or did not conform to the request schema.
	ErrInvalidSyntax = &Error{Status: 400, Type: "invalidSyntax"}

	// The "path" attribute was invalid or malformed.
	ErrInvalidPath = &Error{Status: 400, Type: "invalidPath"}

	// The specified "path" did not yield an attribute or attribute value that could be operated on. This occurs when
	// the specified "path" value contains a filter that yields no match.
	ErrNoTarget = &Error{Status: 400, Type: "noTarget"}

	// A required value was missing, or the value specified was not compatible with the operation or attribute type.
	ErrInvalidValue = &Error{Status: 400, Type: "invalidValue"}

	// The resource was not found from persistence store.
	ErrNotFound = &Error{Status: 404, Type: "notFound"}

	// The specified request cannot be completed, due to the passing of sensitive information in a request URI.
	ErrSensitive = &Error{Status: 400, Type: "sensitive"}

	// The resource is in conflict with some pre conditions.
	ErrConflict = &Error{Status: 412, Type: "conflict"}

	// Server encountered internal error.
	ErrInternal = &Error{Status: 500, Type: "internal"}
)

// A SCIM error message.
// The structure is left completely open for convenience, but it is not recommended to create Error directly.
// To create an error, use the error prototypes (i.e. ErrInvalidFilter). If needed, wrap the error prototype
// by fmt.Errorf("additional detail: %w", err).
type Error struct {
	Status int
	Type   string
}

func (s Error) Error() string {
	return s.Type
}

var (
	_ error = (*Error)(nil)
)
