package handlerutil

import (
	"encoding/json"
	"errors"
	scimjson "github.com/imulab/go-scim/pkg/v2/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/service"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"net/http"
)

// WriteResourceToResponse writes the given resource to http.ResponseWriter, respecting the attributes or excludedAttributes
// specified through options. Any error during the process will be returned.
// Apart from writing the JSON representation of the resource to body, this method also sets Content-Type header to
// application/scim+json; sets Location header to resource's meta.location field, if any; and sets ETag header to
// resource's meta.version field, if any. This method does not set response status, which should be set before calling
// this method.
func WriteResourceToResponse(rw http.ResponseWriter, resource *prop.Resource, options ...scimjson.Options) error {
	raw, jsonErr := scimjson.Serialize(resource, options...)
	if jsonErr != nil {
		return jsonErr
	}

	rw.Header().Set("Content-Type", spec.ApplicationScimJson)
	if location := resource.MetaLocationOrEmpty(); len(location) > 0 {
		rw.Header().Set("Location", location)
	}
	if version := resource.MetaVersionOrEmpty(); len(version) > 0 {
		rw.Header().Set("ETag", version)
	}

	_, writeErr := rw.Write(raw)
	return writeErr
}

// WriteSearchResultToResponse writes the search result to http.ResponseWrite, respecting the attribute or excludedAttributes
// specified through options. Any error during the process will be returned.
// This method also sets Content-Type header to application/scim+json. This method does not set response status, which should
// be set before calling this method.
func WriteSearchResultToResponse(rw http.ResponseWriter, searchResult *service.QueryResponse, options ...scimjson.Options) error {
	render := SearchResultRendering{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: searchResult.TotalResults,
		StartIndex:   searchResult.StartIndex,
		ItemsPerPage: searchResult.ItemsPerPage,
		Resources:    []json.RawMessage{},
	}

	for _, resource := range searchResult.Resources {
		raw, err := scimjson.Serialize(resource, options...)
		if err != nil {
			return err
		}
		render.Resources = append(render.Resources, raw)
	}

	rw.Header().Set("Content-Type", spec.ApplicationScimJson)
	return json.NewEncoder(rw).Encode(render)
}

// WriteError writes the error to the http.ResponseWriter. Any error during the process will be returned.
// If the cause of the error (determined using errors.Unwrap) is a *spec.Error, the cause status and scimType will be
// used together with the error's message as detail. If the cause is not a *spec.Error, spec.ErrInternal is used instead.
// This method also writes the http status with the error's defined status, and set Content-Type header to application/scim+json.
func WriteError(rw http.ResponseWriter, err error) error {
	var errMsg = struct {
		Schemas  []string `json:"schemas"`
		Status   int      `json:"status"`
		ScimType string   `json:"scimType"`
		Detail   string   `json:"detail"`
	}{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Detail:  err.Error(),
	}

	cause := errors.Unwrap(err)
	if scimError, ok := cause.(*spec.Error); ok {
		errMsg.Status = scimError.Status
		errMsg.ScimType = scimError.Type
	} else {
		errMsg.Status = spec.ErrInternal.Status
		errMsg.ScimType = spec.ErrInternal.Type
	}

	rw.Header().Set("Content-Type", spec.ApplicationScimJson)
	rw.WriteHeader(errMsg.Status)

	raw, jsonErr := json.Marshal(errMsg)
	if jsonErr != nil {
		return jsonErr
	}

	_, writeErr := rw.Write(raw)
	return writeErr
}

// SearchResultRendering is the JSON rendering structure for search results. This is very similar to
// service.QueryResponse except that resources are pre-rendered to adapt for objects serialized using
// scim json mechanism or go's json mechanism.
type SearchResultRendering struct {
	Schemas      []string          `json:"schemas"`
	TotalResults int               `json:"totalResults"`
	StartIndex   int               `json:"startIndex"`
	ItemsPerPage int               `json:"itemsPerPage"`
	Resources    []json.RawMessage `json:"Resources,omitempty"`
}
