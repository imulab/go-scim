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

func TestPatchHandler(t *testing.T) {
	s := new(PatchHandlerTestSuite)
	s.resourceBase = "../../tests/patch_handler_test_suite"
	suite.Run(t, s)
}

type PatchHandlerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *PatchHandlerTestSuite) TestPatch() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	spc := s.mustServiceProviderConfig("/service_provider_config.json")

	tests := []struct {
		name       string
		getHandler func(t *testing.T) *Patch
		getRequest func(t *testing.T) http.Request
		expect     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "patch existing resource",
			getHandler: func(t *testing.T) *Patch {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Patch{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					Service: &services.PatchService{
						Logger:          log.None(),
						Database:        database,
						PrePatchFilters: []filters.ForResource{},
						PostPatchFilters: []filters.ForResource{
							filters.CopyReadOnly(),
							filters.Password(10),
							filters.Validation(database),
							filters.Meta(),
						},
						ServiceProviderConfig: spc,
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/patch_001.json")
				require.Nil(t, err)
				return http.DefaultRequest(httptest.NewRequest("PATCH", "/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3", f),
					[]string{"/Users/(?P<userId>.*)"})
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 200, rr.Result().StatusCode)

				patched := prop.NewResource(resourceType)
				err := scimJSON.Deserialize(rr.Body.Bytes(), patched)
				assert.Nil(t, err)
				{
					// userName changed to davidiamyou
					p, _ := patched.NewNavigator().FocusName("userName")
					assert.Equal(t, "davidiamyou", p.Raw())
				}
				{
					// name.givenName changed to David
					nav := patched.NewNavigator()
					_, _ = nav.FocusName("name")
					_, _ = nav.FocusName("givenName")
					assert.Equal(t, "David", nav.Current().Raw())
				}
				{
					// first email (imulab@foo.com) is no longer primary
					nav := patched.NewNavigator()
					_, _ = nav.FocusName("emails")
					_, _ = nav.FocusIndex(0)
					_, _ = nav.FocusName("value")
					assert.Equal(t, "imulab@foo.com", nav.Current().Raw())
					nav.Retract()
					_, _ = nav.FocusName("primary")
					assert.Nil(t, nav.Current().Raw())
				}
				{
					// second email (imulab@bar.com) is now primary
					nav := patched.NewNavigator()
					_, _ = nav.FocusName("emails")
					_, _ = nav.FocusIndex(1)
					_, _ = nav.FocusName("value")
					assert.Equal(t, "imulab@bar.com", nav.Current().Raw())
					nav.Retract()
					_, _ = nav.FocusName("primary")
					assert.Equal(t, true, nav.Current().Raw())
				}
				{
					// only one phone number left
					nav := patched.NewNavigator()
					_, _ = nav.FocusName("phoneNumbers")
					assert.Equal(t, 1, nav.Current().(prop.Container).CountChildren())
				}
			},
		},
		{
			name: "patch non-existing resource",
			getHandler: func(t *testing.T) *Patch {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Patch{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					Service: &services.PatchService{
						Logger:          log.None(),
						Database:        database,
						PrePatchFilters: []filters.ForResource{},
						PostPatchFilters: []filters.ForResource{
							filters.CopyReadOnly(),
							filters.Password(10),
							filters.Validation(database),
							filters.Meta(),
						},
						ServiceProviderConfig: spc,
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/patch_001.json")
				require.Nil(t, err)
				return http.DefaultRequest(httptest.NewRequest("PATCH", "/Users/foobar", f),
					[]string{"/Users/(?P<userId>.*)"})
			},
			expect: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, 404, rr.Result().StatusCode)
			},
		},
		{
			name: "invalid patch should fail",
			getHandler: func(t *testing.T) *Patch {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_000.json", resourceType))
				require.Nil(t, err)
				return &Patch{
					Log:                 log.None(),
					ResourceIDPathParam: "userId",
					Service: &services.PatchService{
						Logger:          log.None(),
						Database:        database,
						PrePatchFilters: []filters.ForResource{},
						PostPatchFilters: []filters.ForResource{
							filters.CopyReadOnly(),
							filters.Password(10),
							filters.Validation(database),
							filters.Meta(),
						},
						ServiceProviderConfig: spc,
					},
				}
			},
			getRequest: func(t *testing.T) http.Request {
				f, err := os.Open(s.resourceBase + "/patch_002.json")
				require.Nil(t, err)
				return http.DefaultRequest(httptest.NewRequest("PATCH", "/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3", f),
					[]string{"/Users/(?P<userId>.*)"})
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

func (s *PatchHandlerTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *PatchHandlerTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *PatchHandlerTestSuite) mustSchema(filePath string) *spec.Schema {
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

func (s *PatchHandlerTestSuite) mustServiceProviderConfig(filePath string) *spec.ServiceProviderConfig {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	spc := new(spec.ServiceProviderConfig)
	err = json.Unmarshal(raw, spc)
	s.Require().Nil(err)

	return spc
}
