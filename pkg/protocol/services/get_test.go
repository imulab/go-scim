package services

import (
	"context"
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetService(t *testing.T) {
	s := new(GetServiceTestSuite)
	s.resourceBase = "../../tests/get_service_test_suite"
	suite.Run(t, s)
}

type GetServiceTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *GetServiceTestSuite) TestGet() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name       string
		getService func(t *testing.T) *GetService
		request    *GetRequest
		expect     func(t *testing.T, response *GetResponse, err error)
	}{
		{
			name: "get existing",
			getService: func(t *testing.T) *GetService {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &GetService{
					Logger:   log.None(),
					Database: database,
				}
			},
			request: &GetRequest{
				ResourceID: "a5866759-32ca-4e2a-9808-a0fe74f94b18",
			},
			expect: func(t *testing.T, response *GetResponse, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, response.Resource)
				assert.NotEmpty(t, response.Version)
				assert.NotEmpty(t, response.Location)
			},
		},
		{
			name: "get non-existing",
			getService: func(t *testing.T) *GetService {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &GetService{
					Logger:   log.None(),
					Database: database,
				}
			},
			request: &GetRequest{
				ResourceID: "foobar",
			},
			expect: func(t *testing.T, response *GetResponse, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.getService(t)
			response, err := service.GetResource(context.Background(), test.request)
			test.expect(t, response, err)
		})
	}
}

func (s *GetServiceTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *GetServiceTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *GetServiceTestSuite) mustSchema(filePath string) *spec.Schema {
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

