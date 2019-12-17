package handler

import (
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/imulab/go-scim/protocol/http"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services"
)

type Create struct {
	Log          log.Logger
	Service      *services.CreateService
	ResourceType *spec.ResourceType
}

func (h *Create) Handle(request http.Request, response http.Response) {
	h.Log.Info("request to create resource")

	var payload *prop.Resource
	{
		raw, err := request.Body()
		if err != nil {
			h.Log.Error("failed to read request body: %s", err.Error())
			WriteError(response, errors.Internal("failed to read request body"))
			return
		}

		payload = prop.NewResource(h.ResourceType)
		err = json.Deserialize(raw, payload)
		if err != nil {
			h.Log.Error("failed to parse request body: %s", err.Error())
			WriteError(response, err)
			return
		}
	}

	cr, err := h.Service.CreateResource(request.Context(), &services.CreateRequest{
		Payload: payload,
	})
	if err != nil {
		WriteError(response, err)
		return
	}

	raw, err := json.Serialize(cr.Resource, json.Options())
	if err != nil {
		WriteError(response, err)
		return
	}

	response.WriteStatus(201)
	response.WriteSCIMContentType()
	response.WriteETag(cr.Version)
	response.WriteLocation(cr.Location)
	response.WriteBody(raw)
}
