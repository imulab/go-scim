package handlers

import (
	"github.com/davidiamyou/go-scim/shared"
	"context"
	"net/http"
)

func GetAllSchemaHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	jsonBytes, err := server.MarshalJSON([]interface{}{
		server.Schema(shared.UserUrn),
		server.Schema(shared.GroupUrn),
	}, nil, nil, nil)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.Body(jsonBytes)
	return
}

func GetSchemaByIdHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	id, _ := ParseIdAndVersion(r)

	switch id {
	case shared.UserUrn:
		jsonBytes, err := server.MarshalJSON(server.Schema(shared.UserUrn), nil, nil, nil)
		ErrorCheck(err)
		ri.Body(jsonBytes)
	case shared.GroupUrn:
		jsonBytes, err := server.MarshalJSON(server.Schema(shared.GroupUrn), nil, nil, nil)
		ErrorCheck(err)
		ri.Body(jsonBytes)
	default:
		panic(shared.Error.ResourceNotFound(id, ""))
	}

	ri.Status(http.StatusOK)
	return
}
