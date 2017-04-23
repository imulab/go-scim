package handlers

import (
	"context"
	"fmt"
	"github.com/davidiamyou/go-scim/shared"
	"net/http"
)

func CreateUserHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
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

func PatchUserHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.UserUrn)
	repo := server.Repository(shared.UserResourceType)

	id, version := ParseIdAndVersion(r)
	ctx = context.WithValue(ctx, shared.ResourceId{}, id)

	resource, err := repo.Get(id, version)
	ErrorCheck(err)

	mod, err := ParseModification(r)
	ErrorCheck(err)
	err = mod.Validate()
	ErrorCheck(err)

	for _, patch := range mod.Ops {
		err = server.ApplyPatch(patch, resource.(*shared.Resource), sch, ctx)
		ErrorCheck(err)
	}

	reference, err := repo.Get(id, version)
	ErrorCheck(err)

	err = server.ValidateType(resource.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.CorrectCase(resource.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.ValidateRequired(resource.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.ValidateMutability(resource.(*shared.Resource), reference.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.ValidateUniqueness(resource.(*shared.Resource), sch, repo, ctx)
	ErrorCheck(err)

	err = server.AssignReadOnlyValue(resource.(*shared.Resource), ctx)
	ErrorCheck(err)

	err = repo.Update(id, version, resource)
	ErrorCheck(err)

	json, err := server.MarshalJSON(resource, sch, []string{}, []string{})
	ErrorCheck(err)

	location := resource.GetData()["meta"].(map[string]interface{})["location"].(string)
	newVersion := resource.GetData()["meta"].(map[string]interface{})["version"].(string)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	if len(newVersion) > 0 {
		ri.ETagHeader(newVersion)
	}
	if len(location) > 0 {
		ri.LocationHeader(location)
	}
	ri.Body(json)
	return
}

func ReplaceUserHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.UserUrn)
	repo := server.Repository(shared.UserResourceType)

	resource, err := ParseBodyAsResource(r)
	ErrorCheck(err)

	id, version := ParseIdAndVersion(r)
	ctx = context.WithValue(ctx, shared.ResourceId{}, id)
	reference, err := repo.Get(id, version)
	ErrorCheck(err)

	err = server.ValidateType(resource, sch, ctx)
	ErrorCheck(err)

	err = server.CorrectCase(resource, sch, ctx)
	ErrorCheck(err)

	err = server.ValidateRequired(resource, sch, ctx)
	ErrorCheck(err)

	err = server.ValidateMutability(resource, reference.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.ValidateUniqueness(resource, sch, repo, ctx)
	ErrorCheck(err)

	err = server.AssignReadOnlyValue(resource, ctx)
	ErrorCheck(err)

	err = repo.Update(id, version, resource)
	ErrorCheck(err)

	json, err := server.MarshalJSON(resource, sch, []string{}, []string{})
	ErrorCheck(err)

	location := resource.GetData()["meta"].(map[string]interface{})["location"].(string)
	newVersion := resource.GetData()["meta"].(map[string]interface{})["version"].(string)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	if len(newVersion) > 0 {
		ri.ETagHeader(newVersion)
	}
	if len(location) > 0 {
		ri.LocationHeader(location)
	}
	ri.Body(json)
	return
}

func QueryUserHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.UserUrn)

	attributes, excludedAttributes := ParseInclusionAndExclusionAttributes(r)

	sr, err := ParseSearchRequest(r, server)
	ErrorCheck(err)

	err = sr.Validate(sch)
	ErrorCheck(err)

	repo := server.Repository(shared.UserResourceType)
	lr, err := repo.Search(sr)
	ErrorCheck(err)

	json, err := server.MarshalJSON(lr, sch, attributes, excludedAttributes)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	ri.Body(json)
	return
}

func DeleteUserByIdHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	id, version := ParseIdAndVersion(r)
	repo := server.Repository(shared.UserResourceType)

	err := repo.Delete(id, version)
	ErrorCheck(err)

	ri.Status(http.StatusNoContent)
	return
}

func GetUserByIdHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.UserUrn)

	id, version := ParseIdAndVersion(r)

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

	json, err := server.MarshalJSON(dp, sch, attributes, excludedAttributes)
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
