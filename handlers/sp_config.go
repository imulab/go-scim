package handlers

import (
	"context"
	"github.com/parsable/go-scim/shared"
	"net/http"
)

func GetServiceProviderConfigHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	repo := server.Repository(shared.ServiceProviderConfigResourceType)
	spConfig, err := repo.Get("", "", ctx)
	ErrorCheck(err)
	jsonBytes, err := server.MarshalJSON(spConfig.GetData(), nil, nil, nil)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.Body(jsonBytes)
	return
}
