package handler

import (
	"context"
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/http"
	"github.com/imulab/go-scim/pkg/protocol/lock"
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

func TestReplaceHandler(t *testing.T) {
	s := new(ReplaceHandlerTestSuite)
	s.resourceBase = "../../tests/replace_handler_test_suite"
	suite.Run(t, s)
}

type ReplaceHandlerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *ReplaceHandlerTestSuite) TestReplace() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name       string
		getHandler func(t *testing.T) *Replace
		getRequest func(t *testing.T) http.Request
		expect     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "replace existing resource",
			getHandler: func(t *testing.T) *Replace {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Replace{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					ResourceType:        resourceType,
					Service: &services.ReplaceService{
						Logger:   log.None(),
						Lock:     lock.Default(),
						Database: database,
						Filters: []filters.ForResource{
							filters.ClearReadOnly(),
							filters.CopyReadOnly(),
							filters.Password(10),
							filters.Validation(database),
							filters.Meta(),
						},
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/user_001.json")
				require.Nil(t, err)
				return http.DefaultRequest(
					httptest.NewRequest("POST", "/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3", f),
					[]string{"/Users/(?P<userId>.*)"})
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 200, rr.Result().StatusCode)
				assert.Equal(t, "application/json+scim", rr.Header().Get("Content-Type"))
				assert.NotEmpty(t, rr.Header().Get("Location"))
				assert.NotEmpty(t, rr.Header().Get("ETag"))
				assert.NotEmpty(t, rr.Body.String())
			},
		},
		{
			name: "replace non-existing resource",
			getHandler: func(t *testing.T) *Replace {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Replace{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					ResourceType:        resourceType,
					Service: &services.ReplaceService{
						Logger:   log.None(),
						Lock:     lock.Default(),
						Database: database,
						Filters: []filters.ForResource{
							filters.ClearReadOnly(),
							filters.CopyReadOnly(),
							filters.Password(10),
							filters.Validation(database),
							filters.Meta(),
						},
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/user_001.json")
				require.Nil(t, err)
				return http.DefaultRequest(
					httptest.NewRequest("POST", "/Users/foobar", f),
					[]string{"/Users/(?P<userId>.*)"})
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 404, rr.Result().StatusCode)
			},
		},
		{
			name: "replace resource that introduces no change",
			getHandler: func(t *testing.T) *Replace {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Replace{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					ResourceType:        resourceType,
					Service: &services.ReplaceService{
						Logger:   log.None(),
						Lock:     lock.Default(),
						Database: database,
						Filters: []filters.ForResource{
							filters.ClearReadOnly(),
							filters.CopyReadOnly(),
							filters.Password(10),
							filters.Validation(database),
							filters.Meta(),
						},
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/user_002.json")
				require.Nil(t, err)
				return http.DefaultRequest(
					httptest.NewRequest("POST", "/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3", f),
					[]string{"/Users/(?P<userId>.*)"})
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 204, rr.Result().StatusCode)
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

func (s *ReplaceHandlerTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *ReplaceHandlerTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *ReplaceHandlerTestSuite) mustSchema(filePath string) *spec.Schema {
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
