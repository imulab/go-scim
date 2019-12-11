package handler

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/errors"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/protocol/http"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/imulab/go-scim/pkg/protocol/services"
)

type Patch struct {
	Log                 log.Logger
	Service             *services.PatchService
	ResourceIDPathParam string
}

func (h *Patch) Handle(request http.Request, response http.Response) {
	var payload *services.PatchRequest
	{
		resourceID := request.PathParam(h.ResourceIDPathParam)

		raw, err := request.Body()
		if err != nil {
			h.Log.Error("failed to read request body for patching resource [id=%s]: %s", resourceID, err.Error())
			WriteError(response, errors.Internal("failed to read request body"))
			return
		}

		payload := new(services.PatchRequest)
		if err := json.Unmarshal(raw, payload); err != nil {
			h.Log.Error("failed to parse request body for patching resource [id=%s]: %s", resourceID, err.Error())
			WriteError(response, err)
			return
		}
		payload.ResourceID = resourceID
		payload.MatchCriteria = interpretConditionalHeader(request)
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
