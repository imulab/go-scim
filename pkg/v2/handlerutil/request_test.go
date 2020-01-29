package handlerutil

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/service"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetRequestProjection(t *testing.T) {
	tests := []struct {
		name        string
		requestFunc func() *http.Request
		expect      func(t *testing.T, projection *crud.Projection, err error)
	}{
		{
			name: "no projection parameters",
			requestFunc: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			expect: func(t *testing.T, projection *crud.Projection, err error) {
				assert.Nil(t, err)
				assert.Nil(t, projection)
			},
		},
		{
			name: "projection with attributes",
			requestFunc: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.URL.RawQuery = url.Values{
					paramAttributes: []string{"foo bar baz"},
				}.Encode()
				return r
			},
			expect: func(t *testing.T, projection *crud.Projection, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"foo", "bar", "baz"}, projection.Attributes)
				assert.Nil(t, projection.ExcludedAttributes)
			},
		},
		{
			name: "projection with excludedAttributes",
			requestFunc: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.URL.RawQuery = url.Values{
					paramExcludedAttributes: []string{"foo bar baz"},
				}.Encode()
				return r
			},
			expect: func(t *testing.T, projection *crud.Projection, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"foo", "bar", "baz"}, projection.ExcludedAttributes)
				assert.Nil(t, projection.Attributes)
			},
		},
		{
			name: "projection with both attributes and excludedAttributes",
			requestFunc: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.URL.RawQuery = url.Values{
					paramAttributes:         []string{"foo bar baz"},
					paramExcludedAttributes: []string{"foo bar baz"},
				}.Encode()
				return r
			},
			expect: func(t *testing.T, projection *crud.Projection, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidSyntax, errors.Unwrap(err))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := test.requestFunc()
			projection, err := GetRequestProjection(req)
			test.expect(t, projection, err)
		})
	}
}

func TestQueryRequestFromGet(t *testing.T) {
	tests := []struct {
		name        string
		requestFunc func() *http.Request
		expect      func(t *testing.T, qr *service.QueryRequest, err error)
	}{
		{
			name: "query with filter",
			requestFunc: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.URL.RawQuery = url.Values{
					paramFilter: []string{"id pr"},
				}.Encode()
				return r
			},
			expect: func(t *testing.T, qr *service.QueryRequest, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "id pr", qr.Filter)
			},
		},
		{
			name: "query with sort",
			requestFunc: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.URL.RawQuery = url.Values{
					paramSortBy:    []string{"userName"},
					paramSortOrder: []string{"ascending"},
				}.Encode()
				return r
			},
			expect: func(t *testing.T, qr *service.QueryRequest, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "userName", qr.Sort.By)
				assert.Equal(t, crud.SortAsc, qr.Sort.Order)
			},
		},
		{
			name: "query with pagination",
			requestFunc: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.URL.RawQuery = url.Values{
					paramStartIndex: []string{"2"},
					paramCount:      []string{"3"},
				}.Encode()
				return r
			},
			expect: func(t *testing.T, qr *service.QueryRequest, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 2, qr.Pagination.StartIndex)
				assert.Equal(t, 3, qr.Pagination.Count)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := test.requestFunc()
			qr, err := QueryRequestFromGet(req)
			test.expect(t, qr, err)
		})
	}
}

func TestQueryRequestFromPost(t *testing.T) {
	tests := []struct {
		name        string
		requestFunc func() *http.Request
		expect      func(t *testing.T, qr *service.QueryRequest, err error)
	}{
		{
			name: "typical",
			requestFunc: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
{
  "schemas": [
    "urn:ietf:params:scim:api:messages:2.0:SearchRequest"
  ],
  "filter": "id pr",
  "sortBy": "userName",
  "sortOrder": "ascending",
  "startIndex": 2,
  "count": 3,
  "attributes": ["id", "meta", "userName"]
}
`))
			},
			expect: func(t *testing.T, qr *service.QueryRequest, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "id pr", qr.Filter)
				assert.Equal(t, "userName", qr.Sort.By)
				assert.Equal(t, crud.SortAsc, qr.Sort.Order)
				assert.Equal(t, 2, qr.Pagination.StartIndex)
				assert.Equal(t, 3, qr.Pagination.Count)
				assert.Equal(t, []string{"id", "meta", "userName"}, qr.Projection.Attributes)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := test.requestFunc()
			qr, _, err := QueryRequestFromPost(req)
			test.expect(t, qr, err)
		})
	}
}
