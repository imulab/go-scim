package protocol

import (
	"io/ioutil"
	"net/http"
)

const (
	HeaderContentType = "Content-Type"
	ContentTypeApplicationJsonScim = "application/json+scim"
	ContentTypeApplicationJson     = "application/json"
)

type (
	HttpProvider interface {
		// Return the name of the HTTP request method, as defined in RFC 7231 and RFC 5789.
		Method(r *http.Request) string
		// Return the body of the HTTP request, or an error
		Body(r *http.Request) ([]byte, error)
		// Return the header value from the request, or an empty string if header does not exist
		Header(r *http.Request, name string) string
		// Render response, or return an error
		Render(rw http.ResponseWriter, status int, headers map[string]string, body []byte) error
	}

	defaultHttpProvider struct {}
)

func NewDefaultHttpProvider() HttpProvider {
	return &defaultHttpProvider{}
}

func (h *defaultHttpProvider) Method(r *http.Request) string {
	return r.Method
}

func (h *defaultHttpProvider) Body(r *http.Request) ([]byte, error) {
	defer func() {
		_ = r.Body.Close()
	}()
	return ioutil.ReadAll(r.Body)
}

func (h *defaultHttpProvider) Header(r *http.Request, name string) string {
	return r.Header.Get(name)
}

func (h *defaultHttpProvider) Render(rw http.ResponseWriter, status int, headers map[string]string, body []byte) error {
	_, err := rw.Write(body)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers {
			rw.Header().Set(k, v)
		}
	}

	rw.WriteHeader(status)

	return nil
}
