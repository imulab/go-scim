package handlers

import (
	"context"
	"fmt"
	"github.com/davidiamyou/go-scim/shared"
	"net/http"
)

func CreateUserHandler(r *http.Request, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.UserUrn)

	resource, err := ParseBodyAsResource(r)
	ErrorCheck(err)

	err = server.ValidateType(resource, sch, ctx)
	ErrorCheck(err)

	err = server.CorrectCase(resource, sch, ctx)
	ErrorCheck(err)

	err = server.ValidateRequired(resource, sch, ctx)
	ErrorCheck(err)

	repo := server.Repository(shared.UserResourceType)
	err = server.ValidateUniqueness(resource, sch, repo, ctx)
	ErrorCheck(err)

	err = server.AssignReadOnlyValue(resource, ctx)
	ErrorCheck(err)

	err = repo.Create(resource)
	ErrorCheck(err)

	json, err := server.MarshalJSON(resource, sch, []string{}, []string{})
	ErrorCheck(err)

	location := resource.GetData()["meta"].(map[string]interface{})["location"].(string)
	version := resource.GetData()["meta"].(map[string]interface{})["version"].(string)

	ri.Status(http.StatusCreated)
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

func DeleteUserByIdHandler(r *http.Request, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	id, version := ParseIdAndVersion(r, server.UrlParam)
	repo := server.Repository(shared.UserResourceType)

	err := repo.Delete(id, version)
	ErrorCheck(err)

	ri.Status(http.StatusNoContent)
	return
}

func GetUserByIdHandler(r *http.Request, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.UserUrn)

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
