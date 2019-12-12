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
	filters "github.com/imulab/go-scim/pkg/protocol/services/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateHandler(t *testing.T) {
	s := new(CreateHandlerTestSuite)
	s.resourceBase = "../../tests/create_handler_test_suite"
	suite.Run(t, s)
}

type CreateHandlerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *CreateHandlerTestSuite) TestCreate() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct{
		name 		string
		getHandler	func(t *testing.T) *Create
		getRequest	func(t *testing.T) http.Request
		expect		func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: 	"create new user",
			getHandler: func(t *testing.T) *Create {
				database := db.Memory()
				return &Create{
					Log:          log.None(),
					ResourceType: resourceType,
					Service:      &services.CreateService{
						Logger:   log.None(),
						Database: database,
						Filters:  []filters.ForResource{
							filters.ClearReadOnly(),
							filters.ID(),
							filters.Password(10),
							filters.Meta(),
							filters.Validation(database),
						},
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/user_001.json")
				s.Require().Nil(err)
				return http.DefaultRequest(httptest.NewRequest("POST", "/Users", f), nil)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 201, rr.Result().StatusCode)
				assert.Equal(t, "application/json+scim", rr.Header().Get("Content-Type"))
				assert.NotEmpty(t, rr.Header().Get("Location"))
				assert.NotEmpty(t, rr.Header().Get("ETag"))
			},
		},
		{
			name: 	"create conflict user",
			getHandler: func(t *testing.T) *Create {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Create{
					Log:          log.None(),
					ResourceType: resourceType,
					Service:      &services.CreateService{
						Logger:   log.None(),
						Database: database,
						Filters:  []filters.ForResource{
							filters.ClearReadOnly(),
							filters.ID(),
							filters.Password(10),
							filters.Meta(),
							filters.Validation(database),
						},
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/user_001.json")
				s.Require().Nil(err)
				return http.DefaultRequest(httptest.NewRequest("POST", "/Users", f), nil)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 400, rr.Result().StatusCode)
			},
		},
		{
			name: 	"create bad user representation",
			getHandler: func(t *testing.T) *Create {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Create{
					Log:          log.None(),
					ResourceType: resourceType,
					Service:      &services.CreateService{
						Logger:   log.None(),
						Database: database,
						Filters:  []filters.ForResource{
							filters.ClearReadOnly(),
							filters.ID(),
							filters.Password(10),
							filters.Meta(),
							filters.Validation(database),
						},
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/user_002.json")
				s.Require().Nil(err)
				return http.DefaultRequest(httptest.NewRequest("POST", "/Users", f), nil)
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 400, rr.Result().StatusCode)
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

func (s *CreateHandlerTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *CreateHandlerTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *CreateHandlerTestSuite) mustSchema(filePath string) *spec.Schema {
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
