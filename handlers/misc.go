package handlers

import (
	"context"
	"encoding/json"
	"github.com/parsable/go-scim/shared"
	"math"
	"net/http"
	"strings"
)

func RootQueryHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema("")

	attributes, excludedAttributes := ParseInclusionAndExclusionAttributes(r)

	sr, err := ParseSearchRequest(r, server)
	ErrorCheck(err)

	err = sr.Validate(sch)
	ErrorCheck(err)

	repo := server.Repository("")
	lr, err := repo.Search(sr, ctx)
	ErrorCheck(err)

	jsonBytes, err := server.MarshalJSON(lr, sch, attributes, excludedAttributes)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	ri.Body(jsonBytes)
	return
}

func BulkHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	userUri := server.Property().GetString("scim.protocol.uri.user")
	groupUri := server.Property().GetString("scim.protocol.uri.group")

	bodyBytes, err := r.Body()
	ErrorCheck(err)

	bulkRequest := &shared.BulkReq{FailOnErrors: math.MaxInt32}
	err = json.Unmarshal(bodyBytes, bulkRequest)
	ErrorCheck(err)

	err = bulkRequest.Validate(server.Property())
	ErrorCheck(err)

	errCount := 0
	allResps := make([]*shared.BulkRespOp, 0, len(bulkRequest.Operations))
	for _, op := range bulkRequest.Operations {
		if errCount > bulkRequest.FailOnErrors {
			break
		}

		opReq := &BulkWebRequest{}
		opReq.Populate(op, server.Property())

		var handler EndpointHandler
		switch opReq.Method() {
		case http.MethodPost:
			switch {
			case strings.HasPrefix(opReq.Target(), userUri):
				handler = CreateUserHandler
			case strings.HasPrefix(opReq.Target(), groupUri):
				handler = CreateGroupHandler
			}
		case http.MethodPut:
			switch {
			case strings.HasPrefix(opReq.Target(), userUri):
				handler = ReplaceUserHandler
			case strings.HasPrefix(opReq.Target(), groupUri):
				handler = ReplaceGroupHandler
			}
		case http.MethodPatch:
			switch {
			case strings.HasPrefix(opReq.Target(), userUri):
				handler = PatchUserHandler
			case strings.HasPrefix(opReq.Target(), groupUri):
				handler = PatchGroupHandler
			}
		case http.MethodDelete:
			switch {
			case strings.HasPrefix(opReq.Target(), userUri):
				handler = DeleteUserByIdHandler
			case strings.HasPrefix(opReq.Target(), groupUri):
				handler = DeleteGroupByIdHandler
			}
		default:
			panic(shared.Error.Text("No handler found for bulk operation"))
		}

		opRi := handler(opReq, server, ctx)
		if opRi.statusCode > 299 {
			errCount++
		}

		opResp := &shared.BulkRespOp{}
		opResp.Populate(op, opRi)
		allResps = append(allResps, opResp)
	}

	respBody, err := server.MarshalJSON(&shared.BulkResp{
		Schemas:    []string{shared.BulkResponseUrn},
		Operations: allResps,
	}, nil, nil, nil)
	ErrorCheck(err)

	ri.statusCode = http.StatusOK
	ri.responseBody = respBody
	return
}
