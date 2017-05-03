package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/davidiamyou/go-scim/shared"
	"github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// interface for server, provides all necessary components for processing
type ScimServer interface {
	// utilities
	Property() PropertySource
	Logger() Logger
	WebRequest(r *http.Request) WebRequest

	// schema
	Schema(id string) *Schema
	InternalSchema(id string) *Schema

	// case
	CorrectCase(subj *Resource, sch *Schema, ctx context.Context) error

	// patch
	ApplyPatch(patch Patch, subj *Resource, sch *Schema, ctx context.Context) error

	// validation
	ValidateType(subj *Resource, sch *Schema, ctx context.Context) error
	ValidateRequired(subj *Resource, sch *Schema, ctx context.Context) error
	ValidateMutability(subj *Resource, ref *Resource, sch *Schema, ctx context.Context) error
	ValidateUniqueness(subj *Resource, sch *Schema, repo Repository, ctx context.Context) error

	// read only generation
	AssignReadOnlyValue(r *Resource, ctx context.Context) error

	// json
	MarshalJSON(v interface{}, sch *Schema, attributes []string, excludedAttributes []string) ([]byte, error)

	// repo
	Repository(identifier string) Repository
}

// functional interface for all endpoints to implement
type EndpointHandler func(r WebRequest, server ScimServer, ctx context.Context) *ResponseInfo

// throw the error out immediately
// error handlers are supposed to recover it and write appropriate responses
func ErrorCheck(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	errorTemplate    = `{"schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"], "Status": "%d", "scimType":"%s", "detail":"%s"}`
	errorTemplateAlt = `{"schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"], "Status": "%d", "detail":"%s"}`
)

func ErrorRecovery(next EndpointHandler) EndpointHandler {
	return func(req WebRequest, server ScimServer, ctx context.Context) (info *ResponseInfo) {
		defer func() {
			if r := recover(); r != nil {
				info = newResponse()
				info.ScimJsonHeader()

				switch r.(type) {
				case *InvalidPathError:
					info.Status(http.StatusBadRequest)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusBadRequest,
							"invalidPath",
							r.(error).Error()),
					))

				case *InvalidFilterError:
					info.Status(http.StatusBadRequest)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusBadRequest,
							"invalidFilter",
							r.(error).Error()),
					))

				case *InvalidTypeError:
					info.Status(http.StatusBadRequest)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusBadRequest,
							"invalidSyntax",
							r.(error).Error()),
					))

				case *NoAttributeError:
					info.Status(http.StatusBadRequest)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusBadRequest,
							"invalidSyntax",
							r.(error).Error()),
					))

				case *MissingRequiredPropertyError:
					info.Status(http.StatusBadRequest)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusBadRequest,
							"invalidValue",
							r.(error).Error()),
					))

				case *MutabilityViolationError:
					info.Status(http.StatusBadRequest)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusBadRequest,
							"mutability",
							r.(error).Error()),
					))

				case *InvalidParamError:
					info.Status(http.StatusBadRequest)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusBadRequest,
							"invalidValue",
							r.(error).Error()),
					))

				case *ResourceNotFoundError:
					switch req.Method() {
					case http.MethodPut, http.MethodPatch, http.MethodDelete:
						_, version := ParseIdAndVersion(req)
						if len(version) == 0 {
							info.Status(http.StatusNotFound)
						} else {
							info.Status(http.StatusPreconditionFailed)
						}
					default:
						info.Status(http.StatusNotFound)
					}
					info.Body([]byte(fmt.Sprintf(errorTemplateAlt, info.statusCode, r.(error).Error())))

				case *DuplicateError:
					info.Status(http.StatusConflict)
					info.Body([]byte(
						fmt.Sprintf(
							errorTemplate,
							http.StatusConflict,
							"uniqueness",
							r.(error).Error()),
					))

				default:
					info.Status(http.StatusInternalServerError)
					info.Body([]byte(fmt.Sprintf(
						errorTemplateAlt,
						http.StatusInternalServerError,
						r.(error).Error()),
					))
				}
			}
		}()
		return next(req, server, ctx)
	}
}

func InjectRequestScope(next EndpointHandler, requestType int) EndpointHandler {
	return func(req WebRequest, server ScimServer, ctx context.Context) (info *ResponseInfo) {
		ctx = context.WithValue(ctx, RequestId{}, uuid.NewV4().String())
		ctx = context.WithValue(ctx, RequestTimestamp{}, time.Now().Unix())
		ctx = context.WithValue(ctx, RequestType{}, requestType)
		return next(req, server, ctx)
	}
}

func Endpoint(next EndpointHandler, server ScimServer) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.Background()
		resp := next(server.WebRequest(req), server, ctx)
		rw.WriteHeader(resp.statusCode)
		for k, v := range resp.headers {
			rw.Header().Set(k, v)
		}
		rw.Write(resp.responseBody)
	})
}

func ParseIdAndVersion(req WebRequest) (id, version string) {
	id = req.Param("resourceId")
	switch req.Method() {
	case http.MethodGet:
		version = req.Header("If-None-Match")
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		version = req.Header("If-Match")
	}
	return
}

func ParseInclusionAndExclusionAttributes(req WebRequest) (attributes, excludedAttributes []string) {
	attributes = strings.Split(req.Param("attributes"), ",")
	excludedAttributes = strings.Split(req.Param("excludedAttributes"), ",")
	return
}

func ParseBodyAsResource(req WebRequest) (*Resource, error) {
	raw, err := req.Body()
	if err != nil {
		return nil, Error.Text("failed to read request: %s", err.Error())
	}

	data := make(map[string]interface{}, 0)
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, Error.InvalidParam("request body", "json conforming to resource syntax", err.Error())
	}

	return &Resource{Complex: Complex(data)}, nil
}

func ParseModification(req WebRequest) (Modification, error) {
	m := Modification{}
	reqBody, err := req.Body()
	if err != nil {
		return Modification{}, err
	}
	err = json.Unmarshal(reqBody, &m)
	if err != nil {
		return Modification{}, err
	}
	return m, nil
}

func ParseSearchRequest(req WebRequest, server ScimServer) (SearchRequest, error) {
	switch req.Method() {
	case http.MethodGet:
		sr := SearchRequest{
			Schemas:    []string{SearchUrn},
			StartIndex: 1,
			Count:      server.Property().GetInt("scim.protocol.itemsPerPage"),
		}
		sr.Attributes = strings.Split(req.Param("attributes"), ",")
		sr.ExcludedAttributes = strings.Split(req.Param("excludedAttributes"), ",")
		sr.Filter = req.Param("filter")
		sr.SortBy = req.Param("sortBy")
		sr.SortOrder = req.Param("sortOrder")
		if v := req.Param("startIndex"); len(v) > 0 {
			if i, err := strconv.Atoi(v); err != nil {
				return SearchRequest{}, Error.InvalidParam("startIndex", "1-based integer", v)
			} else {
				if i < 1 {
					sr.StartIndex = 1
				} else {
					sr.StartIndex = i
				}
			}
		}
		if v := req.Param("count"); len(v) > 0 {
			if i, err := strconv.Atoi(v); err != nil {
				return SearchRequest{}, Error.InvalidParam("count", "non-negative integer", v)
			} else {
				if i < 0 {
					sr.Count = 0
				} else {
					sr.Count = i
				}
			}
		}
		return sr, nil

	case http.MethodPost:
		sr := SearchRequest{
			StartIndex: 1,
			Count:      server.Property().GetInt("scim.protocol.itemsPerPage"),
		}
		reqBody, err := req.Body()
		if err != nil {
			return SearchRequest{}, err
		}
		err = json.Unmarshal(reqBody, &sr)
		if err != nil {
			return SearchRequest{}, err
		}
		return sr, nil

	default:
		return SearchRequest{}, Error.Text("%s method is not supported for search request", req.Method)
	}
}

// response info
type ResponseInfo struct {
	statusCode   int
	headers      map[string]string
	responseBody []byte
}

func newResponse() *ResponseInfo {
	return &ResponseInfo{
		statusCode:   http.StatusOK,
		headers:      map[string]string{},
		responseBody: nil,
	}
}

func (ri *ResponseInfo) GetStatus() int {
	return ri.statusCode
}

func (ri *ResponseInfo) GetHeader(name string) string {
	return ri.headers[name]
}

func (ri *ResponseInfo) GetBody() []byte {
	return ri.responseBody
}

func (ri *ResponseInfo) Status(statusCode int) *ResponseInfo {
	ri.statusCode = statusCode
	return ri
}

func (ri *ResponseInfo) ScimJsonHeader() *ResponseInfo {
	ri.headers["Content-Type"] = "application/scim+json"
	return ri
}

func (ri *ResponseInfo) LocationHeader(location string) *ResponseInfo {
	ri.headers["Location"] = location
	return ri
}

func (ri *ResponseInfo) ETagHeader(version string) *ResponseInfo {
	ri.headers["ETag"] = version
	return ri
}

func (ri *ResponseInfo) Header(k, v string) *ResponseInfo {
	ri.headers[k] = v
	return ri
}

func (ri *ResponseInfo) Body(content []byte) *ResponseInfo {
	ri.responseBody = content
	return ri
}

// bulk web request, implements WebRequest
type BulkWebRequest struct {
	target  string
	method  string
	headers map[string]string
	params  map[string]string
	body    []byte
}

func (bwr BulkWebRequest) Target() string            { return bwr.target }
func (bwr BulkWebRequest) Method() string            { return bwr.method }
func (bwr BulkWebRequest) Header(name string) string { return bwr.headers[name] }
func (bwr BulkWebRequest) Param(name string) string  { return bwr.params[name] }
func (bwr BulkWebRequest) Body() ([]byte, error)     { return bwr.body, nil }
func (bwr BulkWebRequest) Populate(op BulkReqOp, ps PropertySource) {
	userUri := ps.GetString("scim.protocol.uri.user")
	groupUri := ps.GetString("scim.protocol.uri.user")

	bwr.target = op.Path
	bwr.method = strings.ToUpper(op.Method)
	bwr.headers = make(map[string]string, 0)
	bwr.params = make(map[string]string, 0)
	switch bwr.method {
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		if strings.HasPrefix(bwr.target, userUri+"/") {
			bwr.params["resourceId"] = strings.TrimPrefix(bwr.target, userUri+"/")
		} else if strings.HasPrefix(bwr.target, groupUri+"/") {
			bwr.params["resourceId"] = strings.TrimPrefix(bwr.target, groupUri+"/")
		}
	}
	switch bwr.method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		bwr.body = []byte(op.Data)
	}
}
