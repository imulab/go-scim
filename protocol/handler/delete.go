package handler

import (
	"github.com/imulab/go-scim/protocol/http"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services"
)

type Delete struct {
	Log                 log.Logger
	Service             *services.DeleteService
	ResourceIDPathParam string
}

func (h *Delete) Handle(request http.Request, response http.Response) {
	dr := &services.DeleteRequest{
		ResourceID:    request.PathParam(h.ResourceIDPathParam),
		MatchCriteria: interpretConditionalHeader(request),
	}

	err := h.Service.DeleteResource(request.Context(), dr)
	if err != nil {
		WriteError(response, err)
		return
	}

	response.WriteStatus(204)
}