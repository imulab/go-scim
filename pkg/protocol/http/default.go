package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"regexp"
)

// Default implementation of Request using Golang's net/http package. Parameter methodAndPattern specifies URL path
// regex patterns indexed to HTTP method, which is used to parse URL path parameters with the help of regex named
// matches. For instance, in order to parse 'userId' out of a pattern of '/Users/:userId', one would register a
// regex pattern of '/Users/(?P<userId>.*)'. This feature is not as easy to use as other router libraries, but serves
// its purpose as a no-dependency default implementation.
func DefaultRequest(req *http.Request, patterns []string) Request {
	return &defaultRequest{
		req:      req,
		patterns: patterns,
	}
}

// Default implementation of the Response using Golang's net/http package.
func DefaultResponse(rw http.ResponseWriter) Response {
	return &defaultResponse{
		rw: rw,
	}
}

type defaultRequest struct {
	req      *http.Request
	patterns []string
}

func (r *defaultRequest) Context() context.Context {
	return r.Context()
}

func (r *defaultRequest) Method() string {
	return r.req.Method
}

func (r *defaultRequest) Header(key string) string {
	return r.req.Header.Get(key)
}

func (r *defaultRequest) PathParam(param string) string {
	for _, pattern := range r.patterns {
		expr := regexp.MustCompile(pattern)
		match := expr.FindStringSubmatch(r.req.URL.Path)
		for i, n := range expr.SubexpNames() {
			if n == param {
				return match[i]
			}
		}
	}
	return ""
}

func (r *defaultRequest) QueryParam(param string) string {
	return r.req.URL.Query().Get(param)
}

func (r *defaultRequest) ContentType() string {
	return r.req.Header.Get(headerContentType)
}

func (r *defaultRequest) Body() ([]byte, error) {
	defer func() {
		_ = r.req.Body.Close()
	}()
	return ioutil.ReadAll(r.req.Body)
}

type defaultResponse struct {
	rw http.ResponseWriter
}

func (r *defaultResponse) WriteStatus(status int) {
	r.rw.WriteHeader(status)
}

func (r *defaultResponse) WriteSCIMContentType() {
	r.rw.Header().Set(headerContentType, applicationJSONPlusSCIM)
}

func (r *defaultResponse) WriteETag(eTag string) {
	if len(eTag) > 0 {
		r.rw.Header().Set(headerETag, eTag)
	}
}

func (r *defaultResponse) WriteLocation(link string) {
	if len(link) > 0 {
		r.rw.Header().Set(headerLocation, link)
	}
}

func (r *defaultResponse) WriteHeader(k, v string) {
	if len(k) > 0 && len(v) > 0 {
		r.rw.Header().Add(k, v)
	}
}

func (r *defaultResponse) WriteBody(body []byte) {
	_, _ = r.rw.Write(body)
}

var (
	_ Request  = (*defaultRequest)(nil)
	_ Response = (*defaultResponse)(nil)
)
