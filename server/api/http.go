package api

import (
	"context"
	scimHttp "github.com/imulab/go-scim/protocol/http"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

func routeHandler(h func(request scimHttp.Request, response scimHttp.Response)) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h(&httpRequest{
			req: r,
			ps:  ps,
		}, scimHttp.DefaultResponse(rw))
	}
}

type httpRequest struct {
	req *http.Request
	ps  httprouter.Params
}

func (r *httpRequest) Context() context.Context {
	return r.req.Context()
}

func (r *httpRequest) Method() string {
	return r.req.Method
}

func (r *httpRequest) Header(key string) string {
	return r.req.Header.Get(key)
}

func (r *httpRequest) PathParam(param string) string {
	return r.ps.ByName(param)
}

func (r *httpRequest) QueryParam(param string) string {
	return r.req.URL.Query().Get(param)
}

func (r *httpRequest) ContentType() string {
	return r.req.Header.Get("Content-Type")
}

func (r *httpRequest) Body() ([]byte, error) {
	defer func() {
		_ = r.req.Body.Close()
	}()
	return ioutil.ReadAll(r.req.Body)
}
