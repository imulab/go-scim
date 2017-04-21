package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/davidiamyou/go-scim/shared"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// interface for server, provides all necessary components for processing
type ScimServer interface {
	// utilities
	Property() PropertySource
	Logger() Logger

	// request
	UrlParam(name string, req *http.Request) string

	// schema
	Schema(id string) *Schema
	InternalSchema(id string) *Schema

	// validation
	// TODO add context
	ValidateType(subj *Resource, sch *Schema) error
	ValidateRequired(subj *Resource, sch *Schema) error
	ValidateMutability(subj *Resource, ref *Resource, sch *Schema) error
	ValidateUniqueness(subj *Resource, sch *Schema, repo Repository) error

	// read only generation
	AssignReadOnlyValue(r *Resource, ctx context.Context) error

	// json
	MarshalJSON(v interface{}, sch *Schema, attributes []string, excludedAttributes []string) ([]byte, error)

	// repo
	Repository(identifier string) Repository
}

// functional interface for all endpoints to implement
type EndpointHandler func(r *http.Request, server ScimServer, ctx context.Context) *ResponseInfo

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
	return func(req *http.Request, server ScimServer, ctx context.Context) (info *ResponseInfo) {
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
					switch req.Method {
					case http.MethodPut, http.MethodPatch, http.MethodDelete:
						info.Status(http.StatusPreconditionFailed)
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
	return func(req *http.Request, server ScimServer, ctx context.Context) (info *ResponseInfo) {
		ctx = context.WithValue(ctx, RequestId{}, uuid.NewV4().String())
		ctx = context.WithValue(ctx, RequestTimestamp{}, time.Now().Unix())
		ctx = context.WithValue(ctx, RequestType{}, requestType)
		return next(req, server, ctx)
	}
}

func Endpoint(next EndpointHandler, server ScimServer) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.Background()
		resp := next(req, server, ctx)
		rw.WriteHeader(resp.statusCode)
		for k, v := range resp.headers {
			rw.Header().Set(k, v)
		}
		rw.Write(resp.responseBody)
	})
}

// function to parse url param
type UrlParamParseFunc func(name string, req *http.Request) string

func ParseIdAndVersion(req *http.Request, paramExtractor UrlParamParseFunc) (id, version string) {
	id = paramExtractor("resourceId", req)
	switch req.Method {
	case http.MethodGet:
		version = req.Header.Get("If-None-Match")
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		version = req.Header.Get("If-Match")
	}
	return
}

func ParseInclusionAndExclusionAttributes(req *http.Request) (attributes, excludedAttributes []string) {
	attributes = strings.Split(req.URL.Query().Get("attributes"), ",")
	excludedAttributes = strings.Split(req.URL.Query().Get("excludedAttributes"), ",")
	return
}

func ParseBodyAsResource(req *http.Request) (*Resource, error) {
	raw, err := ioutil.ReadAll(req.Body)
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
