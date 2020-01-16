package handler

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/errors"
	scimJSON "github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/protocol/crud"
	"github.com/imulab/go-scim/protocol/http"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services"
	"strconv"
	"strings"
)

type Query struct {
	Log     log.Logger
	Service *services.QueryService
}

func (h *Query) Handle(request http.Request, response http.Response) {
	qr, err := h.parseRequest(request)
	if err != nil {
		h.Log.Error("failed to parse query request", log.Args{
			"error": err,
		})
		WriteError(response, err)
		return
	}

	qt, err := h.Service.QueryResource(request.Context(), qr)
	if err != nil {
		WriteError(response, err)
		return
	}

	raw, err := h.serializeResponse(qr, qt)
	if err != nil {
		WriteError(response, err)
		return
	}

	response.WriteBody(raw)
	response.WriteSCIMContentType()
	response.WriteStatus(200)
}

func (h *Query) parseRequest(request http.Request) (qr *services.QueryRequest, err error) {
	qr = new(services.QueryRequest)

	switch request.Method() {
	case "GET":
		if len(request.QueryParam(sortBy)) > 0 {
			qr.Sort = &crud.Sort{
				By:    request.QueryParam(sortBy),
				Order: crud.SortOrder(request.QueryParam(sortOrder)),
			}
		}
		if len(request.QueryParam(filter)) > 0 {
			qr.Filter = request.QueryParam(filter)
		}
		if len(request.QueryParam(attributes)) > 0 || len(request.QueryParam(excludedAttributes)) > 0 {
			qr.Projection = &crud.Projection{}
			if v := strings.TrimSpace(request.QueryParam(attributes)); len(v) > 0 {
				qr.Projection.Attributes = strings.Split(v, space)
			}
			if v := strings.TrimSpace(request.QueryParam(excludedAttributes)); len(v) > 0 {
				qr.Projection.ExcludedAttributes = strings.Split(v, space)
			}
		}
		if len(request.QueryParam(startIndex)) > 0 || len(request.QueryParam(count)) > 0 {
			var i = 1
			if len(request.QueryParam(startIndex)) > 0 {
				i, err = strconv.Atoi(request.QueryParam(startIndex))
				if err != nil {
					err = errors.InvalidRequest("invalid startIndex parameter")
					return
				}
			}
			var c = 0
			if len(request.QueryParam(count)) > 0 {
				c, err = strconv.Atoi(request.QueryParam(count))
				if err != nil {
					err = errors.InvalidRequest("invalid count parameter")
					return
				}
			}
			qr.Pagination = &crud.Pagination{
				StartIndex: i,
				Count:      c,
			}
		}
	case "POST":
		raw, e := request.Body()
		if e != nil {
			err = errors.Internal("failed to read request body")
			return
		}
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
		if e := json.Unmarshal(raw, wip); e != nil {
			err = errors.InvalidSyntax("failed to parse search request: %s", e.Error())
			return
		}
		if len(wip.Schemas) != 1 || wip.Schemas[0] != "urn:ietf:params:scim:api:messages:2.0:SearchRequest" {
			err = errors.InvalidSyntax("invalid schema for search request")
			return
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
	default:
		err = errors.InvalidRequest("search request only accepts GET and POST")
	}

	return
}

func (h *Query) serializeResponse(request *services.QueryRequest, response *services.QueryResponse) ([]byte, error) {
	wip := struct {
		Schemas      []string                           `json:"schemas"`
		TotalResults int                                `json:"totalResults"`
		ItemsPerPage int                                `json:"itemsPerPage"`
		StartIndex   int                                `json:"startIndex"`
		Resources    []*scimJSON.ResourceMarshalAdapter `json:"Resources"`
	}{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: response.TotalResults,
		ItemsPerPage: response.ItemsPerPage,
		StartIndex:   response.StartIndex,
		Resources:    make([]*scimJSON.ResourceMarshalAdapter, 0, len(response.Resources)),
	}
	for _, r := range response.Resources {
		adp := &scimJSON.ResourceMarshalAdapter{
			Resource: r,
		}
		if request.Projection != nil {
			adp.Include = request.Projection.Attributes
			adp.Exclude = request.Projection.ExcludedAttributes
		}
		wip.Resources = append(wip.Resources, adp)
	}
	return json.Marshal(wip)
}
