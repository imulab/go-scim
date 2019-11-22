package protocol

import (
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/json"
	"github.com/imulab/go-scim/persistence"
	"net/http"
)

type CreateEndpoint struct {
	HttpProvider        HttpProvider
	ResourceType        *core.ResourceType
	FilterFunc          FilterFunc
	PersistenceProvider persistence.Provider
}

func (h *CreateEndpoint) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := h.serveHttpE(rw, r)
	if err != nil {
		RenderError(err, rw, r)
	}
}

func (h *CreateEndpoint) serveHttpE(rw http.ResponseWriter, r *http.Request) (err error) {
	err = h.checkRequest(r)
	if err != nil {
		return
	}

	var resource *core.Resource
	{
		resource, err = h.parseResource(r)
		if err != nil {
			return
		}
	}

	err = h.FilterFunc(r.Context(), resource, nil)
	if err != nil {
		return err
	}

	err = h.PersistenceProvider.InsertOne(r.Context(), resource)
	if err != nil {
		return err
	}

	err = h.renderSuccess(rw, resource)
	if err != nil {
		return err
	}

	return nil
}

func (h *CreateEndpoint) checkRequest(r *http.Request) error {
	if h.HttpProvider.Method(r) != http.MethodPost {
		return core.Errors.InvalidRequest("resource creation request should be submitted via HTTP POST method")
	}

	if contentType := h.HttpProvider.Header(r, HeaderContentType); len(contentType) > 0 {
		if contentType != ContentTypeApplicationJson && contentType != ContentTypeApplicationJsonScim {
			return core.Errors.InvalidRequest("invalid %s header: accepts %s or %s",
				HeaderContentType, ContentTypeApplicationJsonScim, ContentTypeApplicationJson)
		}
	}

	return nil
}

// Parse the HTTP body as a resource guided by the resource type of this endpoint.
func (h *CreateEndpoint) parseResource(r *http.Request) (*core.Resource, error) {
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
func (h *CreateEndpoint) renderSuccess(rw http.ResponseWriter, resource *core.Resource) error {
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
