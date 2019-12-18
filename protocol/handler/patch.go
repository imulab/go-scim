package handler

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/errors"
	scimJSON "github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/protocol/http"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services"
)

type Patch struct {
	Log                 log.Logger
	Service             *services.PatchService
	ResourceIDPathParam string
}

func (h *Patch) Handle(request http.Request, response http.Response) {
	var payload *services.PatchRequest
	{
		payload = new(services.PatchRequest)

		payload.ResourceID = request.PathParam(h.ResourceIDPathParam)
		payload.MatchCriteria = interpretConditionalHeader(request)

		raw, err := request.Body()
		if err != nil {
			h.Log.Error("failed to read request body for patching resource", log.Args{
				"resourceId": payload.ResourceID,
				"error": err,
			})
			WriteError(response, errors.Internal("failed to read request body"))
			return
		}
		if err := json.Unmarshal(raw, payload); err != nil {
			h.Log.Error("failed to parse request body for patching resource", log.Args{
				"resourceId": payload.ResourceID,
				"error": err,
			})
			WriteError(response, err)
			return
		}
	}

	pr, err := h.Service.PatchResource(request.Context(), payload)
	if err != nil {
		WriteError(response, err)
		return
	}

	if pr.NewVersion == pr.OldVersion {
		response.WriteLocation(pr.Location)
		response.WriteETag(pr.NewVersion)
		response.WriteStatus(204)
	} else {
		raw, err := scimJSON.Serialize(pr.Resource, scimJSON.Options())
		if err != nil {
			WriteError(response, err)
			return
		}
		response.WriteBody(raw)
		response.WriteLocation(pr.Location)
		response.WriteETag(pr.NewVersion)
		response.WriteSCIMContentType()
		response.WriteStatus(200)
	}
}
