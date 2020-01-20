package errors

import (
	"fmt"
)

// Defined error types. Each error type will have a corresponding constructor.
const (
	TypeInvalidRequest = "invalidRequest"
	TypeInvalidFilter  = "invalidFilter"
	TypeTooMany        = "tooMany"
	TypeUniqueness     = "uniqueness"
	TypeMutability     = "mutability"
	TypeInvalidSyntax  = "invalidSyntax"
	TypeInvalidPath    = "invalidPath"
	TypeNoTarget       = "noTarget"
	TypeInvalidValue   = "invalidValue"
	TypePreCondition   = "preCondition"
	TypeSensitive      = "sensitive"
	TypeNotFound       = "notFound"
	TypeInternal       = "internal"
)

// Returns error to describe that the request is invalid. This is a generic error description,
// use more specific ones if available.
func InvalidRequest(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeInvalidRequest,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the specified filter syntax was invalid, or the specified attribute and
// filter comparison combination is not supported.
func InvalidFilter(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeInvalidFilter,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the specified filter yields many more results than the server is willing
// to calculate or process.
func TooMany(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeTooMany,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that one or more of the attribute values are already in use or are reserved.
func Uniqueness(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeUniqueness,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the attempted modification is not compatible with the target attribute's mutability
// or current state (e.g., modification of an "immutable" attribute with an existing value).
func Mutability(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeMutability,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the request body message structure was invalid or did not conform to the request schema.
func InvalidSyntax(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeInvalidSyntax,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the "path" attribute was invalid or malformed.
func InvalidPath(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeInvalidPath,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the specified "path" did not yield an attribute or attribute value that could be
// operated on. This occurs when the specified "path" value contains a filter that yields no match.
func NoTarget(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeNoTarget,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that a required value was missing, or the value specified was not compatible with
// the operation or attribute type.
func InvalidValue(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeInvalidValue,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the specified precondition failed and request cannot proceed.
func PreConditionFailed(format string, args ...interface{}) error {
	return &Error{
		Status:  412,
		Type:    TypePreCondition,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the specified request cannot be completed, due to the passing of sensitive
// (e.g., personal) information in a request URI.
func Sensitive(format string, args ...interface{}) error {
	return &Error{
		Status:  400,
		Type:    TypeSensitive,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that the request resource is not found in the data store.
func NotFound(format string, args ...interface{}) error {
	return &Error{
		Status:  404,
		Type:    TypeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// Returns error to describe that server encountered internal error. This should be the returned error when the user
// input is not at fault.
func Internal(format string, args ...interface{}) error {
	return &Error{
		Status:  500,
		Type:    TypeInternal,
		Message: fmt.Sprintf(format, args...),
	}
}
