package handlerutil

import (
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/service"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"net/http"
	"strconv"
	"strings"
)

const (
	paramFilter             = "filter"
	paramSortBy             = "sortBy"
	paramSortOrder          = "sortOrder"
	paramStartIndex         = "startIndex"
	paramCount              = "count"
	paramAttributes         = "attributes"
	paramExcludedAttributes = "excludedAttributes"
)

// GetRequestProjection returns a nullable *crud.Projection structure that may encapsulate the attributes or excludedAttributes
// parameters present in the HTTP GET request.
func GetRequestProjection(request *http.Request) (projection *crud.Projection, err error) {
	if attrValue := request.URL.Query().Get(paramAttributes); len(attrValue) > 0 {
		projection = &crud.Projection{
			Attributes: strings.Split(strings.TrimSpace(attrValue), " "),
		}
	}

	if exclAttrValue := request.URL.Query().Get(paramExcludedAttributes); len(exclAttrValue) > 0 {
		if projection != nil && len(projection.Attributes) > 0 {
			err = fmt.Errorf("%w: only one of attributes and excludedAttributes may be specified", spec.ErrInvalidSyntax)
			return
		}
		projection = &crud.Projection{
			ExcludedAttributes: strings.Split(strings.TrimSpace(exclAttrValue), " "),
		}
	}

	return
}

// CreateRequest returns a parsed *service.CreateRequest directly from *http.Request, and a closer function which should
// be called after resource processing is done (preferably using defer).
func CreateRequest(request *http.Request) (cr *service.CreateRequest, closer func()) {
	cr = &service.CreateRequest{PayloadSource: request.Body}
	closer = func() {
		_ = request.Body.Close()
	}
	return
}

// QueryRequestFromGet returns a parsed *service.QueryRequest from *http.Request using HTTP GET method, and any error
// during parsing.
func QueryRequestFromGet(request *http.Request) (qr *service.QueryRequest, err error) {
	qr = &service.QueryRequest{}

	if filter := request.URL.Query().Get(paramFilter); len(filter) > 0 {
		qr.Filter = filter
	}

	if sortBy := request.URL.Query().Get(paramSortBy); len(sortBy) > 0 {
		qr.Sort = &crud.Sort{
			By:    sortBy,
			Order: crud.SortOrder(request.URL.Query().Get(paramSortOrder)),
		}
	}

	if startIndexValue, countValue := request.URL.Query().Get(paramStartIndex), request.URL.Query().Get(paramCount); len(startIndexValue) > 0 || len(countValue) > 0 {

		qr.Pagination = &crud.Pagination{}

		if len(startIndexValue) > 0 {
			qr.Pagination.StartIndex, err = strconv.Atoi(startIndexValue)
			if err != nil || qr.Pagination.StartIndex < 1 {
				err = fmt.Errorf("%w: parameter startIndex must be a 1-based integer", spec.ErrInvalidSyntax)
				return
			}
		} else {
			qr.Pagination.StartIndex = 1
		}

		if len(countValue) > 0 {
			qr.Pagination.Count, err = strconv.Atoi(countValue)
			if err != nil || qr.Pagination.Count < 0 {
				err = fmt.Errorf("%w: parameter count must be a non-negative integer", spec.ErrInvalidSyntax)
				return
			}
		} else {
			qr.Pagination.Count = 0
		}
	}

	qr.Projection, err = GetRequestProjection(request)
	if err != nil {
		return
	}

	return
}

// QueryRequestFromPost returns a parsed *service.QueryRequest from *http.Request using HTTP POST method, a closer function
// to be invoked when the search is finished, and any error during the parsing.
func QueryRequestFromPost(request *http.Request) (qr *service.QueryRequest, closer func(), err error) {
	wip := new(struct {
		Schemas            []string `json:"schemas"`
		Attributes         []string `json:"attributes"`
		ExcludedAttributes []string `json:"excludedAttributes"`
		Filter             string   `json:"filter"`
		SortBy             string   `json:"sortBy"`
		SortOrder          string   `json:"sortOrder"`
		StartIndex         int      `json:"startIndex"`
		Count              int      `json:"count"`
	})
	if err = json.NewDecoder(request.Body).Decode(wip); err != nil {
		return
	}
	closer = func() {
		_ = request.Body.Close()
	}

	if len(wip.Schemas) != 1 || wip.Schemas[0] != "urn:ietf:params:scim:api:messages:2.0:SearchRequest" {
		err = fmt.Errorf("%w: invalid schema for search request", spec.ErrInvalidSyntax)
		return
	}
	qr = &service.QueryRequest{
		Filter: wip.Filter,
	}

	if len(wip.SortBy) > 0 {
		qr.Sort = &crud.Sort{
			By:    wip.SortBy,
			Order: crud.SortOrder(wip.SortOrder), // validate it later
		}
	}

	if len(wip.Attributes) > 0 || len(wip.ExcludedAttributes) > 0 {
		qr.Projection = &crud.Projection{
			Attributes:         wip.Attributes,
			ExcludedAttributes: wip.ExcludedAttributes,
		}
	}

	if wip.StartIndex > 0 || wip.Count > 0 {
		if wip.StartIndex == 0 {
			wip.StartIndex = 1
		}
		qr.Pagination = &crud.Pagination{
			StartIndex: wip.StartIndex,
			Count:      wip.Count,
		}
	}

	return
}

// ReplaceRequest returns a function that will supply a complete built *service.ReplaceRequest when given resourceId,
// and a closer function which should be called after resource processing is done (preferably using defer).
func ReplaceRequest(request *http.Request) (rr func(resourceId string) *service.ReplaceRequest, closer func()) {
	rr = func(resourceId string) *service.ReplaceRequest {
		return &service.ReplaceRequest{
			ResourceID:    resourceId,
			PayloadSource: request.Body,
			MatchCriteria: MatchCriteria(request),
		}
	}
	closer = func() {
		_ = request.Body.Close()
	}
	return
}

// PatchRequest returns a function that will supply a complete built *service.PatchRequest when given resourceId, and
// a closer function which should be called after resource processing is done (preferably using defer).
func PatchRequest(request *http.Request) (pr func(resourceId string) *service.PatchRequest, closer func()) {
	pr = func(resourceId string) *service.PatchRequest {
		return &service.PatchRequest{
			ResourceID:    resourceId,
			MatchCriteria: MatchCriteria(request),
			PayloadSource: request.Body,
		}
	}
	closer = func() {
		_ = request.Body.Close()
	}
	return
}

// DeleteRequest returns a function that will supply a complete built *service.DeleteRequest when given resourceId.
func DeleteRequest(request *http.Request) func(resourceId string) *service.DeleteRequest {
	return func(resourceId string) *service.DeleteRequest {
		return &service.DeleteRequest{
			ResourceID:    resourceId,
			MatchCriteria: MatchCriteria(request),
		}
	}
}

// MatchCriteria returns a function to be supplied as the match criteria argument in replace, patch and delete requests.
// It checks for If-Match and If-None-Match headers and supports asterisk (*) and comma delimited resource versions.
// The If-Match header takes precedence over If-None-Match header. If none of the headers are present, it returns a
// function that always returns true.
func MatchCriteria(request *http.Request) func(resource *prop.Resource) bool {
	if ifMatch := request.Header.Get("If-Match"); len(ifMatch) > 0 {
		ifMatch = strings.TrimSpace(ifMatch)
		return func(resource *prop.Resource) bool {
			version := resource.MetaVersionOrEmpty()
			if ifMatch == "*" {
				return true
			}
			for _, eachVersion := range strings.Split(ifMatch, ",") {
				if strings.TrimSpace(eachVersion) == version {
					return true
				}
			}
			return false
		}
	}

	if ifNoneMatch := request.Header.Get("If-None-Match"); len(ifNoneMatch) > 0 {
		ifNoneMatch = strings.TrimSpace(ifNoneMatch)
		return func(resource *prop.Resource) bool {
			version := resource.MetaVersionOrEmpty()
			if ifNoneMatch == "*" {
				return false
			}
			for _, eachVersion := range strings.Split(ifNoneMatch, ",") {
				if strings.TrimSpace(eachVersion) == version {
					return false
				}
			}
			return true
		}
	}

	return func(_ *prop.Resource) bool {
		return true
	}
}
