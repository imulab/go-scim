package handlers

import (
	"context"
	"fmt"
	"github.com/parsable/go-scim/shared"
	"net/http"
)

func CreateGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)

	resource, err := ParseBodyAsResource(r)
	ErrorCheck(err)

	err = server.ValidateType(resource, sch, ctx)
	ErrorCheck(err)

	err = server.CorrectCase(resource, sch, ctx)
	ErrorCheck(err)

	err = server.ValidateRequired(resource, sch, ctx)
	ErrorCheck(err)

	repo := server.Repository(shared.GroupResourceType)
	err = server.ValidateUniqueness(resource, sch, repo, ctx)
	ErrorCheck(err)

	err = server.AssignReadOnlyValue(resource, ctx)
	ErrorCheck(err)

	err = repo.Create(resource, ctx)
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

func PatchGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)
	repo := server.Repository(shared.GroupResourceType)

	id, version := ParseIdAndVersion(r)
	ctx = context.WithValue(ctx, shared.ResourceId{}, id)

	resource, err := repo.Get(id, version, ctx)
	ErrorCheck(err)

	mod, err := ParseModification(r)
	ErrorCheck(err)
	err = mod.Validate()
	ErrorCheck(err)

	for _, patch := range mod.Ops {
		err = server.ApplyPatch(patch, resource.(*shared.Resource), sch, ctx)
		ErrorCheck(err)
	}

	reference, err := repo.Get(id, version, ctx)
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

	err = repo.Update(id, version, resource, ctx)
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

func ReplaceGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)
	repo := server.Repository(shared.GroupResourceType)

	resource, err := ParseBodyAsResource(r)
	ErrorCheck(err)

	id, version := ParseIdAndVersion(r)
	ctx = context.WithValue(ctx, shared.ResourceId{}, id)
	reference, err := repo.Get(id, version, ctx)
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

	err = repo.Update(id, version, resource, ctx)
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

func QueryGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)

	attributes, excludedAttributes := ParseInclusionAndExclusionAttributes(r)

	sr, err := ParseSearchRequest(r, server)
	ErrorCheck(err)

	err = sr.Validate(sch)
	ErrorCheck(err)

	repo := server.Repository(shared.GroupResourceType)
	lr, err := repo.Search(sr, ctx)
	ErrorCheck(err)

	json, err := server.MarshalJSON(lr, sch, attributes, excludedAttributes)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	ri.Body(json)
	return
}

func DeleteGroupByIdHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	id, version := ParseIdAndVersion(r)
	repo := server.Repository(shared.GroupResourceType)

	err := repo.Delete(id, version, ctx)
	ErrorCheck(err)

	ri.Status(http.StatusNoContent)
	return
}

func GetGroupByIdHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)

	id, version := ParseIdAndVersion(r)

	if len(version) > 0 {
		count, err := server.Repository(shared.GroupResourceType).Count(fmt.Sprintf("id eq \"%s\" and meta.version eq \"%s\"", id, version), ctx, )
		if err == nil && count > 0 {
			ri.Status(http.StatusNotModified)
			return
		}
	}

	attributes, excludedAttributes := ParseInclusionAndExclusionAttributes(r)

	dp, err := server.Repository(shared.GroupResourceType).Get(id, version, ctx)
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
