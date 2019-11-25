package errors

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
)

// Error that request is invalid. This is a generic error description, use more specific ones if available.
func InvalidRequest(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "invalidRequest",
		Message: fmt.Sprintf(format, args...),
	}
}

// The specified filter syntax was invalid, or the specified attribute and
// filter comparison combination is not supported.
func InvalidFilter(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "invalidFilter",
		Message: fmt.Sprintf(format, args...),
	}
}

// The specified filter yields many more results than the server is willing
// to calculate or process.
func TooMany(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "tooMany",
		Message: fmt.Sprintf(format, args...),
	}
}

// One or more of the attribute values are already in use or are reserved.
func Uniqueness(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "uniqueness",
		Message: fmt.Sprintf(format, args...),
	}
}

// The attempted modification is not compatible with the target attribute's mutability
// or current state (e.g., modification of an "immutable" attribute with an existing value).
func Mutability(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "mutability",
		Message: fmt.Sprintf(format, args...),
	}
}

// The request body message structure was invalid or did not conform to the request schema.
func InvalidSyntax(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "invalidSyntax",
		Message: fmt.Sprintf(format, args...),
	}
}

// The "path" attribute was invalid or malformed.
func InvalidPath(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "invalidPath",
		Message: fmt.Sprintf(format, args...),
	}
}

// The specified "path" did not yield an attribute or attribute value that could be
// operated on. This occurs when the specified "path" value contains a filter that yields
// no match.
func NoTarget(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "noTarget",
		Message: fmt.Sprintf(format, args...),
	}
}

// A required value was missing, or the value specified was not compatible with the operation
// or attribute type.
func InvalidValue(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "invalidValue",
		Message: fmt.Sprintf(format, args...),
	}
}

// The specified request cannot be completed, due to the passing of sensitive (e.g., personal)
// information in a request URI.
func Sensitive(format string, args ...interface{}) error {
	return &core.Error{
		Status:  400,
		Type:    "sensitive",
		Message: fmt.Sprintf(format, args...),
	}
}

// The request resource is not found in the data store.
func NotFound(format string, args ...interface{}) error {
	return &core.Error{
		Status:  404,
		Type:    "notFound",
		Message: fmt.Sprintf(format, args...),
	}
}

// Server encountered internal error
func Internal(format string, args ...interface{}) error {
	return &core.Error{
		Status:  500,
		Type:    "internal",
		Message: fmt.Sprintf(format, args...),
	}
}
