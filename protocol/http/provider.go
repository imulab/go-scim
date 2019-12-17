package http

import "context"

// Abstraction of HTTP request, with respect to function related to SCIM.
type Request interface {
	// Return the context of this request
	Context() context.Context
	// Returns HTTP method names in capital letters.
	Method() string
	// Returns the HTTP header value by the key
	Header(key string) string
	// Get the URL path parameter of the name, or return empty string
	PathParam(param string) string
	// Get the URL query parameter of the name, or return empty string
	QueryParam(param string) string
	// Return the Content-Type header value, or empty string
	ContentType() string
	// Read the request body, and return content in bytes, or return an error
	Body() ([]byte, error)
}

// Abstraction of HTTP response, with respect to function related to SCIM.
type Response interface {
	// Write the response status
	WriteStatus(status int)
	// Set the Content-Type response header to application/json+scim
	WriteSCIMContentType()
	// Set the ETag response header to the given value
	WriteETag(eTag string)
	// Set the Location response header to the given value
	WriteLocation(link string)
	// Set the custom response header k to the value v
	WriteHeader(k, v string)
	// Write the given bytes to response body.
	WriteBody(body []byte)
}