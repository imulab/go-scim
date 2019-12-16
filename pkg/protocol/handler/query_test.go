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
	"strings"
	"testing"
)

func TestQueryHandler(t *testing.T) {
	s := new(QueryHandlerTestSuite)
	s.resourceBase = "../../tests/query_handler_test_suite"
	suite.Run(t, s)
}

type QueryHandlerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *QueryHandlerTestSuite) TestQuery() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	spc := s.mustServiceProviderConfig("/service_provider_config.json")

	tests := []struct {
		name       string
		getHandler func(t *testing.T) *Query
		getRequest func(t *testing.T) http.Request
		expect     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "query with filter, sort, paginate and projection (GET)",
			getHandler: func(t *testing.T) *Query {
				database := db.Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return &Query{
					Log: log.None(),
					Service: &services.QueryService{
						Logger:           log.None(),
						Database:         database,
						ServiceProviderConfig: spc,
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				return http.DefaultRequest(httptest.NewRequest(
					"GET",
					"/Users?filter=userName+pr&sortBy=userName&sortOrder=ascending&startIndex=3&count=2&attributes=userName",
					nil), nil)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 200, rr.Result().StatusCode)
				expect := `
{
   "schemas":[
      "urn:ietf:params:scim:api:messages:2.0:ListResponse"
   ],
   "totalResults":10,
   "itemsPerPage":2,
   "startIndex":3,
   "Resources":[
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"bf5d7cbd-396a-4460-b59c-579250e1a81a",
         "userName":"user003"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"1509dd67-2cf2-4e83-9945-02259403e889",
         "userName":"user004"
      }
   ]
}
`
				assert.JSONEq(t, expect, rr.Body.String())
			},
		},
		{
			name: "query with filter, sort, paginate and projection (POST)",
			getHandler: func(t *testing.T) *Query {
				database := db.Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return &Query{
					Log: log.None(),
					Service: &services.QueryService{
						Logger:           log.None(),
						Database:         database,
						ServiceProviderConfig: spc,
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				body := strings.NewReader(`
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:SearchRequest"],
	"filter": "userName pr",
	"sortBy": "userName",
	"sortOrder": "ascending",
	"startIndex": 3,
	"count": 2,
	"attributes": ["userName"]
}
`)
				return http.DefaultRequest(httptest.NewRequest(
					"POST",
					"/Users/.search",
					body), nil)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 200, rr.Result().StatusCode)
				expect := `
{
   "schemas":[
      "urn:ietf:params:scim:api:messages:2.0:ListResponse"
   ],
   "totalResults":10,
   "itemsPerPage":2,
   "startIndex":3,
   "Resources":[
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"bf5d7cbd-396a-4460-b59c-579250e1a81a",
         "userName":"user003"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"1509dd67-2cf2-4e83-9945-02259403e889",
         "userName":"user004"
      }
   ]
}
`
				assert.JSONEq(t, expect, rr.Body.String())
			},
		},
		{
			name: 	"search with no filter",
			getHandler: func(t *testing.T) *Query {
				database := db.Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return &Query{
					Log: log.None(),
					Service: &services.QueryService{
						Logger:           log.None(),
						Database:         database,
						ServiceProviderConfig: spc,
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				body := strings.NewReader(`
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:SearchRequest"],
	"startIndex": 1,
	"count": 100,
	"sortBy": "userName",
	"sortOrder": "descending",
	"attributes": ["userName"]
}
`)
				return http.DefaultRequest(httptest.NewRequest(
					"POST",
					"/Users/.search",
					body), nil)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 200, rr.Result().StatusCode)
				expect := `
{
   "schemas":[
      "urn:ietf:params:scim:api:messages:2.0:ListResponse"
   ],
   "totalResults":10,
   "itemsPerPage":10,
   "startIndex":1,
   "Resources":[
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"66fd936f-c0e2-428e-b626-5bb164b57518",
         "userName":"user010"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"375109a2-f0a1-434e-a220-a10f81271ad8",
         "userName":"user009"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"d012cb14-77e8-4125-b2e1-9880046d834a",
         "userName":"user008"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"b1d7f21e-a81f-4a3e-99d9-68183a54d412",
         "userName":"user007"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"8e04c6c6-3c27-4f61-8978-165e5420b091",
         "userName":"user006"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"42c1f39d-136f-4b47-bc56-e3775b22a9b0",
         "userName":"user005"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"1509dd67-2cf2-4e83-9945-02259403e889",
         "userName":"user004"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"bf5d7cbd-396a-4460-b59c-579250e1a81a",
         "userName":"user003"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"23d22b2d-4fc4-49f9-90fb-ee10882c69ed",
         "userName":"user002"
      },
      {
         "schemas":[
            "urn:ietf:params:scim:schemas:core:2.0:User"
         ],
         "id":"a5866759-32ca-4e2a-9808-a0fe74f94b18",
         "userName":"user001"
      }
   ]
}
`
				assert.JSONEq(t, expect, rr.Body.String())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			resp := http.DefaultResponse(rr)
			test.getHandler(t).Handle(test.getRequest(t), resp)
			test.expect(t, rr)
		})
	}
}

func (s *QueryHandlerTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *QueryHandlerTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *QueryHandlerTestSuite) mustSchema(filePath string) *spec.Schema {
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

func (s *QueryHandlerTestSuite) mustServiceProviderConfig(filePath string) *spec.ServiceProviderConfig {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	spc := new(spec.ServiceProviderConfig)
	err = json.Unmarshal(raw, spc)
	s.Require().Nil(err)

	return spc
}
