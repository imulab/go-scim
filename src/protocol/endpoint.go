package protocol

import (
	"github.com/imulab/go-scim/src/core"
	"net/http"
)

type Endpoint interface {
	// Returns the resource type this endpoint is responsible for.
	ResourceType() *core.ResourceType
	// Handles the incoming HTTP traffic and returns any error. This method
	// can be adapted to http.HandlerFunc using the Render function.
	Handle(rw http.ResponseWriter, r *http.Request) error
}

// Render adapts the Handle method from the Endpoint interface to a http.HandlerFunc signature
// by rendering any returned error as a SCIM error.
func Render(handler func(rw http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		err := handler(rw, r)
		if err != nil {
			// todo
			// todo use http provider to render bytes
		}
	}
}