package protocol

import "net/http"

type HttpProvider interface {
	// Return the name of the HTTP request method, as defined in RFC 7231 and RFC 5789.
	Method(r *http.Request) string
	// Return the body of the HTTP request, or an error
	Body(r *http.Request) ([]byte, error)
	// Return the header value from the request, or an empty string if header does not exist
	Header(r *http.Request, name string) string
	// Render response, or return an error
	Render(rw http.ResponseWriter, status int, headers map[string]string, body []byte) error
}

const (
	HeaderContentType = "Content-Type"
	ContentTypeApplicationJsonScim = "application/json+scim"
	ContentTypeApplicationJson     = "application/json"
)
