package handlers

import (
	"github.com/davidiamyou/go-scim/shared"
	"context"
	"net/http"
)

func GetServiceProviderConfigHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	repo := server.Repository(shared.ServiceProviderConfigResourceType)
	spConfig, err := repo.Get("", "")
	ErrorCheck(err)
	jsonBytes, err := server.MarshalJSON(spConfig.GetData(), nil, nil, nil)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.Body(jsonBytes)
	return
}
