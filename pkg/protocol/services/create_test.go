package services

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol"
	"github.com/imulab/go-scim/pkg/protocol/filters"
	"github.com/imulab/go-scim/pkg/protocol/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestCreateService(t *testing.T) {
	s := new(CreateServiceTestSuite)
	s.resourceBase = "../../tests/create_service_test_suite"
	suite.Run(t, s)
}

type CreateServiceTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *CreateServiceTestSuite) TestCreate() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name       string
		setup      func(t *testing.T) *CreateService
		getRequest func() *CreateRequest
		expect     func(t *testing.T, resp *CreateResponse, err error)
	}{
		{
			name: "create a new user",
			setup: func(t *testing.T) *CreateService {
				memoryPersistence := persistence.Memory()
				return &CreateService{
					Logger: protocol.NoOpLogger(),
					Filters: []protocol.ResourceFilter{
						filters.NewClearReadOnlyResourceFilter(),
						filters.NewIDResourceFilter(),
						filters.NewPasswordResourceFilter(10),
						filters.NewMetaResourceFilter(),
						filters.NewValidationResourceFilter(memoryPersistence),
					},
					Persistence: memoryPersistence,
				}
			},
			getRequest: func() *CreateRequest {
				resource := s.mustResource("/user_001.json", resourceType)
				return &CreateRequest{
					Payload: resource,
				}
			},
			expect: func(t *testing.T, resp *CreateResponse, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, resp.Resource)
				assert.NotEmpty(t, resp.Version)
				assert.NotEmpty(t, resp.Location)
			},
		},
		{
			name: "create a user missing required userName",
			setup: func(t *testing.T) *CreateService {
				memoryPersistence := persistence.Memory()
				return &CreateService{
					Logger: protocol.NoOpLogger(),
					Filters: []protocol.ResourceFilter{
						filters.NewClearReadOnlyResourceFilter(),
						filters.NewIDResourceFilter(),
						filters.NewPasswordResourceFilter(10),
						filters.NewMetaResourceFilter(),
						filters.NewValidationResourceFilter(memoryPersistence),
					},
					Persistence: memoryPersistence,
				}
			},
			getRequest: func() *CreateRequest {
				resource := s.mustResource("/user_002.json", resourceType)
				return &CreateRequest{
					Payload: resource,
				}
			},
			expect: func(t *testing.T, resp *CreateResponse, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "invalidValue", err.(*core.Error).Type)
			},
		},
		{
			name: "self administered readOnly fields are ignored",
			setup: func(t *testing.T) *CreateService {
				memoryPersistence := persistence.Memory()
				return &CreateService{
					Logger: protocol.NoOpLogger(),
					Filters: []protocol.ResourceFilter{
						filters.NewClearReadOnlyResourceFilter(),
						filters.NewIDResourceFilter(),
						filters.NewPasswordResourceFilter(10),
						filters.NewMetaResourceFilter(),
						filters.NewValidationResourceFilter(memoryPersistence),
					},
					Persistence: memoryPersistence,
				}
			},
			getRequest: func() *CreateRequest {
				resource := s.mustResource("/user_003.json", resourceType)
				return &CreateRequest{
					Payload: resource,
				}
			},
			expect: func(t *testing.T, resp *CreateResponse, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, resp.Resource.ID())
				assert.NotEqual(t, "foobar", resp.Resource.ID())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			req := test.getRequest()
			resp, err := service.CreateResource(context.Background(), req)
			test.expect(t, resp, err)
		})
	}
}

func (s *CreateServiceTestSuite) mustResource(filePath string, resourceType *core.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *CreateServiceTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *CreateServiceTestSuite) mustSchema(filePath string) *core.Schema {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	sch := new(core.Schema)
	err = json.Unmarshal(raw, sch)
	s.Require().Nil(err)

	core.SchemaHub.Put(sch)

	return sch
}
