package api

import (
	gojson "encoding/json"
	"errors"
	"fmt"
	"github.com/imulab/go-scim/v2/pkg/handlerutil"
	"github.com/imulab/go-scim/v2/pkg/json"
	"github.com/imulab/go-scim/v2/pkg/service"
	"github.com/imulab/go-scim/v2/pkg/spec"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"net/http"
)

// CreateHandler returns a route handler function for creating SCIM resources.
func CreateHandler(svc service.Create, log *zerolog.Logger) func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		cr, closer := handlerutil.CreateRequest(r)
		defer closer()

		resp, err := svc.Do(r.Context(), cr)
		if err != nil {
			log.
				Err(err).
				Msg("error when creating resource")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		log.Info().Msg("resource created")
		rw.WriteHeader(201)
		_ = handlerutil.WriteResourceToResponse(rw, resp.Resource)
	}
}

// GetHandler returns a route handler function for getting SCIM resource.
func GetHandler(svc service.Get, log *zerolog.Logger) func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if len(id) == 0 {
			err := fmt.Errorf("%w: id is empty", spec.ErrInvalidSyntax)
			log.
				Err(err).
				Msg("error receiving get request")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		projection, err := handlerutil.GetRequestProjection(r)
		if err != nil {
			log.
				Err(err).
				Msg("error parsing getting request")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		resp, err := svc.Do(r.Context(), &service.GetRequest{
			ResourceID: id,
			Projection: projection,
		})
		if err != nil {
			log.
				Err(err).
				Msg("error when getting resource")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		var opt []json.Options
		if projection != nil {
			if len(projection.Attributes) > 0 {
				opt = append(opt, json.Include(projection.Attributes...))
			}
			if len(projection.ExcludedAttributes) > 0 {
				opt = append(opt, json.Exclude(projection.ExcludedAttributes...))
			}
		}

		_ = handlerutil.WriteResourceToResponse(rw, resp.Resource, opt...)
	}
}

// DeleteHandler returns a route handler function for deleting SCIM resource.
func DeleteHandler(svc service.Delete, log *zerolog.Logger) func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if len(id) == 0 {
			err := fmt.Errorf("%w: id is empty", spec.ErrInvalidSyntax)
			log.
				Err(err).
				Msg("error receiving delete request")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		_, err := svc.Do(r.Context(), handlerutil.DeleteRequest(r)(id))
		if err != nil {
			log.
				Err(err).
				Msg("error when deleting resource")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		rw.WriteHeader(204)
	}
}

// ReplaceHandler returns a route handler function for replacing SCIM resource.
func ReplaceHandler(svc service.Replace, log *zerolog.Logger) func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if len(id) == 0 {
			err := fmt.Errorf("%w: id is empty", spec.ErrInvalidSyntax)
			log.
				Err(err).
				Msg("error receiving replace request")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		reqFunc, closer := handlerutil.ReplaceRequest(r)
		defer closer()

		resp, err := svc.Do(r.Context(), reqFunc(id))
		if err != nil {
			log.
				Err(err).
				Msg("error when replacing resource")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		if !resp.Replaced {
			rw.WriteHeader(204)
			return
		}

		_ = handlerutil.WriteResourceToResponse(rw, resp.Resource)
	}
}

// PatchHandler returns a route handler function for patching SCIM resource.
func PatchHandler(svc service.Patch, log *zerolog.Logger) func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if len(id) == 0 {
			err := fmt.Errorf("%w: id is empty", spec.ErrInvalidSyntax)
			log.
				Err(err).
				Msg("error receiving patching request")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		reqFunc, closer := handlerutil.PatchRequest(r)
		defer closer()

		resp, err := svc.Do(r.Context(), reqFunc(id))
		if err != nil {
			log.
				Err(err).
				Msg("error when patching resource")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		if !resp.Patched {
			rw.WriteHeader(204)
			return
		}

		_ = handlerutil.WriteResourceToResponse(rw, resp.Resource)
	}
}

// SearchHandler returns a route handler function for searching SCIM resources. This handler could be used in HTTP GET and
// HTTP POST scenarios, as defined in the SCIM specification.
func SearchHandler(svc service.Query, log *zerolog.Logger) func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var (
			req    *service.QueryRequest
			err    error
			closer func()
		)

		switch r.Method {
		case http.MethodGet:
			req, err = handlerutil.QueryRequestFromGet(r)
		case http.MethodPost:
			req, closer, err = handlerutil.QueryRequestFromPost(r)
		default:
			err = errors.New("invalid method configured for search handler")
		}

		if err != nil {
			log.
				Err(err).
				Msg("error when parsing search request")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		if closer != nil {
			defer closer()
		}

		resp, err := svc.Do(r.Context(), req)
		if err != nil {
			log.
				Err(err).
				Msg("error when searching resource")
			_ = handlerutil.WriteError(rw, err)
			return
		}

		var opt []json.Options
		if resp.Projection != nil {
			if len(resp.Projection.Attributes) > 0 {
				opt = append(opt, json.Include(resp.Projection.Attributes...))
			}
			if len(resp.Projection.ExcludedAttributes) > 0 {
				opt = append(opt, json.Exclude(resp.Projection.ExcludedAttributes...))
			}
		}

		_ = handlerutil.WriteSearchResultToResponse(rw, resp)
	}
}

// ServiceProviderConfigHandler returns a http route handler to write service provider config info.
func ServiceProviderConfigHandler(config *spec.ServiceProviderConfig) func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	raw, err := gojson.Marshal(config)
	if err != nil {
		panic(err)
	}

	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		rw.WriteHeader(200)
		rw.Header().Set("Content-Type", "application/json+scim")
		_, _ = rw.Write(raw)
	}
}

// HealthHandler returns a http handler to report service health status.
func HealthHandler() func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// This needs rework to include MongoDB and RabbitMQ status.
	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		rw.WriteHeader(200)
	}
}
