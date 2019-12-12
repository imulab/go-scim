package handler

import (
	"context"
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/http"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/imulab/go-scim/pkg/protocol/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetHandler(t *testing.T) {
	s := new(GetHandlerTestSuite)
	s.resourceBase = "../../tests/get_handler_test_suite"
	suite.Run(t, s)
}

type GetHandlerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *GetHandlerTestSuite) TestHandle() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name       string
		getHandler func(t *testing.T) *Get
		getReq     func(t *testing.T) http.Request
		expect     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "get existing user full representation",
			getHandler: func(t *testing.T) *Get {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &Get{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					Service: &services.GetService{
						Logger:   log.None(),
						Database: database,
					},
				}
			},
			getReq: func(t *testing.T) http.Request {
				return http.DefaultRequest(
					httptest.NewRequest("GET", "/Users/a5866759-32ca-4e2a-9808-a0fe74f94b18", nil),
					[]string{"/Users/(?P<userId>.*)"},
				)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				expect := `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "a5866759-32ca-4e2a-9808-a0fe74f94b18",
  "meta": {
    "resourceType": "User",
    "created": "2019-11-20T13:09:00",
    "lastModified": "2019-11-20T13:09:00",
    "location": "https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
    "version": "W/\\\"1\\\""
  },
  "userName": "user001",
  "name": {
    "formatted": "Mr. Weinan Qiu",
    "familyName": "Qiu",
    "givenName": "Weinan",
    "honorificPrefix": "Mr."
  },
  "displayName": "Weinan",
  "profileUrl": "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "userType": "Employee",
  "preferredLanguage": "zh_CN",
  "locale": "zh_CN",
  "timezone": "Asia/Shanghai",
  "active": true,
  "emails": [
    {
      "value": "imulab@foo.com",
      "type": "work",
      "primary": true,
      "display": "imulab@foo.com"
    },
    {
      "value": "imulab@bar.com",
      "type": "home",
      "display": "imulab@bar.com"
    }
  ],
  "phoneNumbers": [
    {
      "value": "123-45678",
      "type": "work",
      "primary": true,
      "display": "123-45678"
    },
    {
      "value": "123-45679",
      "type": "work",
      "display": "123-45679"
    }
  ],
  "ims": [
    {
      "value": "imulab",
      "type": "wechat",
      "primary": true,
      "display": "imulab (wechat)"
    }
  ],
  "addresses": [
    {
      "formatted": "123 Main. St, Shanghai, China",
      "streetAddress": "123 Main. St",
      "locality": "Shanghai",
      "postalCode": "12345",
      "country": "China",
      "type": "work",
      "primary": true
    },
    {
      "formatted": "124 Main. St, Shanghai, China",
      "streetAddress": "124 Main. St",
      "locality": "Shanghai",
      "postalCode": "12345",
      "country": "China",
      "type": "home"
    }
  ],
  "groups": [
    {
      "value": "b2bd79a2-106a-4f7f-913d-9bd2d092c3cb",
      "$ref": "https://identity.imulab.com/Groups/b2bd79a2-106a-4f7f-913d-9bd2d092c3cb",
      "type": "direct",
      "display": "interest group"
    }
  ]
}
`
				assert.Equal(t, 200, rr.Result().StatusCode)
				assert.Equal(t, "application/json+scim", rr.Header().Get("Content-Type"))
				assert.NotEmpty(t, rr.Header().Get("Location"))
				assert.NotEmpty(t, rr.Header().Get("ETag"))
				assert.JSONEq(t, expect, rr.Body.String())
			},
		},
		{
			name: "get existing user partial representation",
			getHandler: func(t *testing.T) *Get {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &Get{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					Service: &services.GetService{
						Logger:   log.None(),
						Database: database,
					},
				}
			},
			getReq: func(t *testing.T) http.Request {
				return http.DefaultRequest(
					httptest.NewRequest("GET", "/Users/a5866759-32ca-4e2a-9808-a0fe74f94b18?attributes=userName", nil),
					[]string{"/Users/(?P<userId>.*)"},
				)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				expect := `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "a5866759-32ca-4e2a-9808-a0fe74f94b18",
  "userName": "user001"
}
`
				assert.Equal(t, 200, rr.Result().StatusCode)
				assert.Equal(t, "application/json+scim", rr.Header().Get("Content-Type"))
				assert.NotEmpty(t, rr.Header().Get("Location"))
				assert.NotEmpty(t, rr.Header().Get("ETag"))
				assert.JSONEq(t, expect, rr.Body.String())
			},
		},
		{
			name: "get existing user partial representation (exclude)",
			getHandler: func(t *testing.T) *Get {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &Get{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					Service: &services.GetService{
						Logger:   log.None(),
						Database: database,
					},
				}
			},
			getReq: func(t *testing.T) http.Request {
				return http.DefaultRequest(
					httptest.NewRequest("GET", "/Users/a5866759-32ca-4e2a-9808-a0fe74f94b18?excludedAttributes=emails+phoneNumbers+groups+ims+addresses", nil),
					[]string{"/Users/(?P<userId>.*)"},
				)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				expect := `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "a5866759-32ca-4e2a-9808-a0fe74f94b18",
  "meta": {
    "resourceType": "User",
    "created": "2019-11-20T13:09:00",
    "lastModified": "2019-11-20T13:09:00",
    "location": "https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
    "version": "W/\\\"1\\\""
  },
  "userName": "user001",
  "name": {
    "formatted": "Mr. Weinan Qiu",
    "familyName": "Qiu",
    "givenName": "Weinan",
    "honorificPrefix": "Mr."
  },
  "displayName": "Weinan",
  "profileUrl": "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "userType": "Employee",
  "preferredLanguage": "zh_CN",
  "locale": "zh_CN",
  "timezone": "Asia/Shanghai",
  "active": true
}
`
				assert.Equal(t, 200, rr.Result().StatusCode)
				assert.Equal(t, "application/json+scim", rr.Header().Get("Content-Type"))
				assert.NotEmpty(t, rr.Header().Get("Location"))
				assert.NotEmpty(t, rr.Header().Get("ETag"))
				assert.JSONEq(t, expect, rr.Body.String())
			},
		},
		{
			name: "get non-existing user",
			getHandler: func(t *testing.T) *Get {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &Get{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					Service: &services.GetService{
						Logger:   log.None(),
						Database: database,
					},
				}
			},
			getReq: func(t *testing.T) http.Request {
				return http.DefaultRequest(
					httptest.NewRequest("GET", "/Users/foobar", nil),
					[]string{"/Users/(?P<userId>.*)"},
				)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 404, rr.Result().StatusCode)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			resp := http.DefaultResponse(rr)
			test.getHandler(t).Handle(test.getReq(t), resp)
			test.expect(t, rr)
		})
	}
}

func (s *GetHandlerTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *GetHandlerTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *GetHandlerTestSuite) mustSchema(filePath string) *spec.Schema {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	sch := new(spec.Schema)
	err = json.Unmarshal(raw, sch)
	s.Require().Nil(err)

	spec.SchemaHub.Put(sch)

	return sch
}
