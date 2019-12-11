package handler

import (
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/protocol/crud"
	"github.com/imulab/go-scim/pkg/protocol/http"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/imulab/go-scim/pkg/protocol/services"
	"strings"
)

type Get struct {
	Log                 log.Logger
	Service             *services.GetService
	ResourceIDPathParam string
}

func (h *Get) Handle(request http.Request, response http.Response) {
	var (
		resourceIDParam         string
		attributesParam         []string
		excludedAttributesParam []string
	)
	{
		resourceIDParam = request.PathParam(h.ResourceIDPathParam)
		h.Log.Info("request to get resource [id=%s]", resourceIDParam)
		if len(resourceIDParam) == 0 {
			WriteError(response, errors.InvalidRequest("missing resource id"))
			return
		}

		attributesParam = strings.Split(strings.TrimSpace(request.QueryParam(attributes)), space)
		excludedAttributesParam = strings.Split(strings.TrimSpace(request.QueryParam(excludedAttributes)), space)
		if len(attributesParam) > 0 && len(excludedAttributesParam) > 0 {
			WriteError(response, errors.InvalidRequest("only one of %s and %s parameter may be used", attributes, excludedAttributes))
			return
		}
	}

	gr, err := h.Service.GetResource(request.Context(), &services.GetRequest{
		Projection: &crud.Projection{
			Attributes:         attributesParam,
			ExcludedAttributes: excludedAttributesParam,
		},
		ResourceID: resourceIDParam,
	})
	if err != nil {
		WriteError(response, err)
		return
	}

	raw, err := json.Serialize(gr.Resource, json.Options().Include(attributesParam...).Exclude(excludedAttributesParam...))
	if err != nil {
		WriteError(response, err)
		return
	}

	response.WriteBody(raw)
	response.WriteSCIMContentType()
	response.WriteETag(gr.Version)
	response.WriteLocation(gr.Location)
	response.WriteStatus(200)
}
