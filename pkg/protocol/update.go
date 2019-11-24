package protocol

import (
	"github.com/imulab/go-scim/pkg/core"
	"github.com/imulab/go-scim/pkg/json"
	"github.com/imulab/go-scim/pkg/persistence"
	"github.com/imulab/go-scim/pkg/protocol/stage"
	"net/http"
)

type UpdateEndpoint struct {
	ResourceIdParamName	string
	HttpProvider        HttpProvider
	ResourceType        *core.ResourceType
	PostParseHook       stage.PostParseHook
	FilterStage         stage.FilterStage
	PrePersistHook      stage.PrePersistHook
	PersistenceProvider persistence.Provider
	PostPersistHook		stage.PostPersistHook
}

func (h *UpdateEndpoint) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := h.serveHttpE(rw, r)
	if err != nil {
		RenderError(err, rw, r)
	}
}

func (h *UpdateEndpoint) serveHttpE(rw http.ResponseWriter, r *http.Request) (err error) {
	err = h.checkRequest(r)
	if err != nil {
		return
	}

	var ref *core.Resource
	{
		ref, err = h.getReference(r)
		if err != nil {
			return
		}
	}

	var resource *core.Resource
	{
		resource, err = h.parseResource(r)
		if err != nil {
			return
		}
	}

	if h.PostParseHook != nil {
		h.PostParseHook.ResourceHasBeenParsed(r.Context(), resource)
	}

	err = h.FilterStage(r.Context(), resource, ref)
	if err != nil {
		return err
	}

	if h.PrePersistHook != nil {
		h.PrePersistHook.ResourceWillBePersisted(r.Context(), resource)
	}

	err = h.PersistenceProvider.ReplaceOne(r.Context(), resource)
	if err != nil {
		return err
	}

	if h.PostPersistHook != nil {
		h.PostPersistHook.ResourceHasBeenPersisted(r.Context(), resource)
	}

	err = h.renderSuccess(rw, resource)
	if err != nil {
		return err
	}

	return nil
}

func (h *UpdateEndpoint) checkRequest(r *http.Request) error {
	if h.HttpProvider.Method(r) != http.MethodPut {
		return core.Errors.InvalidRequest("resource creation request should be submitted via HTTP PUT method")
	}

	if contentType := h.HttpProvider.Header(r, HeaderContentType); len(contentType) > 0 {
		if contentType != ContentTypeApplicationJson && contentType != ContentTypeApplicationJsonScim {
			return core.Errors.InvalidRequest("invalid %s header: accepts %s or %s",
				HeaderContentType, ContentTypeApplicationJsonScim, ContentTypeApplicationJson)
		}
	}

	return nil
}

func (h *UpdateEndpoint) getReference(r *http.Request) (*core.Resource, error) {
	resourceId := h.HttpProvider.PathParam(r, h.ResourceIdParamName)
	if len(resourceId) == 0 {
		return nil, core.Errors.InvalidRequest("resource id is required for update requests")
	}
	return h.PersistenceProvider.GetById(r.Context(), resourceId)
}

// Parse the HTTP body as a resource guided by the resource type of this endpoint.
func (h *UpdateEndpoint) parseResource(r *http.Request) (*core.Resource, error) {
	body, err := h.HttpProvider.Body(r)
	if err != nil {
		return nil, err
	}

	resource := core.Resources.New(h.ResourceType)
	err = json.Deserialize(body, resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// Render the created resource as JSON in the HTTP 201 response, if fails, try render just the 201 status code.
func (h *UpdateEndpoint) renderSuccess(rw http.ResponseWriter, resource *core.Resource) error {
	raw, err := json.Serialize(resource, []string{}, []string{})
	if err != nil {
		return err
	}

	err = h.HttpProvider.Render(rw, http.StatusCreated, map[string]string{
		HeaderContentType: ContentTypeApplicationJsonScim,
	}, raw)
	if err != nil {
		// suppress the error and try render the status code directly
		rw.WriteHeader(http.StatusCreated)
	}

	return nil
}

