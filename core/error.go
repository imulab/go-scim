package core

import "fmt"

// A SCIM error.
type ScimError struct {
	Status  int    `json:"status"`
	Type    string `json:"error_type"`
	Message string `json:"error_message"`
}

// Return the formatted error information for display.
func (s ScimError) Error() string {
	if len(s.Message) == 0 {
		return s.Type
	}
	return fmt.Sprintf("%s: %s", s.Type, s.Message)
}

// Append some hint information to the end of this error
func (s ScimError) Hint(hint string) error {
	return &ScimError{
		Status:  s.Status,
		Type:    s.Type,
		Message: s.Message + " " + hint,
	}
}

// Convenience method to wrap any error with hint
func ErrAppendHint(err error, hint string) error {
	if se, ok := err.(*ScimError); ok {
		return se.Hint(hint)
	} else {
		return Errors.Internal(err.Error()).(*ScimError).Hint(hint)
	}
}

var (
	// Entry point to create a SCIM error.
	Errors = &errorFactory{}
)

// namespace for all error factory methods.
type errorFactory struct{}

// The specified filter syntax was invalid, or the specified attribute and
// filter comparison combination is not supported.
func (f errorFactory) InvalidFilter(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "InvalidFilter",
		Message: message,
	}
}

// The specified filter yields many more results than the server is willing
// to calculate or process.
func (f errorFactory) tooMany(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "tooMany",
		Message: message,
	}
}

// One or more of the attribute values are already in use or are reserved.
func (f errorFactory) uniqueness(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "uniqueness",
		Message: message,
	}
}

// The attempted modification is not compatible with the target attribute's mutability
// or current state (e.g., modification of an "immutable" attribute with an existing value).
func (f errorFactory) mutability(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "mutability",
		Message: message,
	}
}

// The request body message structure was invalid or did not conform to the request schema.
func (f errorFactory) InvalidSyntax(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "invalidSyntax",
		Message: message,
	}
}

// The "path" attribute was invalid or malformed.
func (f errorFactory) InvalidPath(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "invalidPath",
		Message: message,
	}
}

// The specified "path" did not yield an attribute or attribute value that could be
// operated on. This occurs when the specified "path" value contains a filter that yields
// no match.
func (f errorFactory) noTarget(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "noTarget",
		Message: message,
	}
}

// A required value was missing, or the value specified was not compatible with the operation
// or attribute type.
func (f errorFactory) InvalidValue(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "invalidValue",
		Message: message,
	}
}

// The specified request cannot be completed, due to the passing of sensitive (e.g., personal)
// information in a request URI.
func (f errorFactory) sensitive(message string) error {
	return &ScimError{
		Status:  400,
		Type:    "sensitive",
		Message: message,
	}
}

// Server encountered internal error
func (f errorFactory) Internal(message string) error {
	return &ScimError{
		Status:  500,
		Type:    "internal",
		Message: message,
	}
}
