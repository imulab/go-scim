package handlers

import (
	"net/http"
	"github.com/davidiamyou/go-scim/shared"
	"fmt"
	"context"
)

func GetUserById(r *http.Request, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.Schema(shared.UserResourceType)

	id, version := ParseIdAndVersion(r, server.UrlParam)

	if len(version) > 0 {
		count, err := server.Repository(shared.UserResourceType).Count(
			fmt.Sprintf("id eq \"%s\" and meta.version eq \"%s\"", id, version),
		)
		if err == nil && count > 0 {
			ri.Status(http.StatusNotModified)
			return
		}
	}

	attributes, excludedAttributes := ParseInclusionAndExclusionAttributes(r)

	dp, err := server.Repository(shared.UserResourceType).Get(id, version)
	ErrorCheck(err)
	location := dp.GetData()["meta"].(map[string]interface{})["location"].(string)

	json, err := server.MarshalJSON(dp.GetData(), sch, attributes, excludedAttributes)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	if len(version) > 0 {
		ri.ETagHeader(version)
	}
	if len(location) > 0 {
		ri.LocationHeader(location)
	}
	ri.Body(json)
	return
}