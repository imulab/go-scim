package services

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/errors"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/imulab/go-scim/pkg/protocol/services/filter"
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
				memoryDB := db.Memory()
				return &CreateService{
					Logger: log.None(),
					Filters: []filter.ForResource{
						filter.ClearReadOnly(),
						filter.ID(),
						filter.Password(10),
						filter.Meta(),
						filter.Validation(memoryDB),
					},
					Database: memoryDB,
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
				memoryDB := db.Memory()
				return &CreateService{
					Logger: log.None(),
					Filters: []filter.ForResource{
						filter.ClearReadOnly(),
						filter.ID(),
						filter.Password(10),
						filter.Meta(),
						filter.Validation(memoryDB),
					},
					Database: memoryDB,
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
				assert.Equal(t, "invalidValue", err.(*errors.Error).Type)
			},
		},
		{
			name: "self administered readOnly fields are ignored",
			setup: func(t *testing.T) *CreateService {
				memoryDB := db.Memory()
				return &CreateService{
					Logger: log.None(),
					Filters: []filter.ForResource{
						filter.ClearReadOnly(),
						filter.ID(),
						filter.Password(10),
						filter.Meta(),
						filter.Validation(memoryDB),
					},
					Database: memoryDB,
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

func (s *CreateServiceTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *CreateServiceTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *CreateServiceTestSuite) mustSchema(filePath string) *spec.Schema {
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
