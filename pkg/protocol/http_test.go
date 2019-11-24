package protocol

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewDefaultHttpProvider(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/Users/1234567890?foo=bar", strings.NewReader("{\"foo\":\"bar\"}"))
	request.Header.Set(HeaderContentType, ContentTypeApplicationJson)

	provider := NewDefaultHttpProvider([]string{
		"/Users/(?P<userId>.*)",
	})

	assert.Equal(t, http.MethodPatch, provider.Method(request))
	assert.Equal(t, ContentTypeApplicationJson, provider.Header(request, HeaderContentType))
	assert.Equal(t, "bar", provider.QueryParam(request, "foo"))
	assert.Equal(t, "1234567890", provider.PathParam(request, "userId"))

	raw, err := provider.Body(request)
	assert.Nil(t, err)
	assert.Equal(t, "{\"foo\":\"bar\"}", string(raw))
}
