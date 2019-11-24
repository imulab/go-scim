package protocol

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core"
	"net/http"
)

// Render the error as a JSON http response. All error will be attempted to convert back to *core.ScimError. If type
// is incompatible, an internal SCIM error will be constructed and rendered instead.
func RenderError(err error, rw http.ResponseWriter, _ *http.Request) {
	var se *core.ScimError
	{
		if _, ok := err.(*core.ScimError); ok {
			se = err.(*core.ScimError)
		} else {
			se = core.Errors.Internal(err.Error()).(*core.ScimError)
		}
	}

	raw, err := json.Marshal(se)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write(raw)
		rw.WriteHeader(se.Status)
	}
}
