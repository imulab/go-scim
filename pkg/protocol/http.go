package protocol

import (
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	HeaderContentType              = "Content-Type"
	ContentTypeApplicationJsonScim = "application/json+scim"
	ContentTypeApplicationJson     = "application/json"
)

type (
	HttpProvider interface {
		// Return the name of the HTTP request method, as defined in RFC 7231 and RFC 5789.
		Method(r *http.Request) string
		// Return the body of the HTTP request, or an error
		Body(r *http.Request) ([]byte, error)
		// Return the query param of the HTTP request by the name, or an empty string
		QueryParam(r *http.Request, name string) string
		// Return the url path param of the HTTP request by the name, or an empty string
		PathParam(r *http.Request, name string) string
		// Return the header value from the request, or an empty string if header does not exist
		Header(r *http.Request, name string) string
		// Render response, or return an error
		Render(rw http.ResponseWriter, status int, headers map[string]string, body []byte) error
	}

	defaultHttpProvider struct {
		urlRegex []string
	}
)

// Create a new default http provider. The provider utilizes all native Golang libraries and does not require external
// dependencies. The parameter urlRegex requires a slice of regular expression patterns to help match the path variable.
// For instance, to match a userId in a URI pattern which would otherwise be defined as '/Users/:userId' in popular
// mux libraries, one would use the regular expression named match feature: '/Users/(?P<userId>.*)'. When asked for a
// path variable, this provider will match all registered patterns against the URL path and yield the first non-empty
// match as the path variable value.
func NewDefaultHttpProvider(urlRegex []string) HttpProvider {
	return &defaultHttpProvider{
		urlRegex: urlRegex,
	}
}

func (h *defaultHttpProvider) Method(r *http.Request) string {
	return r.Method
}

func (h *defaultHttpProvider) QueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func (h *defaultHttpProvider) PathParam(r *http.Request, name string) string {
	for _, pattern := range h.urlRegex {
		expr := regexp.MustCompile(pattern)
		match := expr.FindStringSubmatch(r.URL.Path)
		for i, n := range expr.SubexpNames() {
			if n == name {
				return match[i]
			}
		}
	}
	return ""
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
