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

func TestDeleteService(t *testing.T) {
	s := new(DeleteServiceTestSuite)
	s.resourceBase = "../../tests/delete_service_test_suite"
	suite.Run(t, s)
}

type DeleteServiceTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *DeleteServiceTestSuite) TestDelete() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	spc := s.mustServiceProviderConfig("/service_provider_config.json")

	tests := []struct {
		name       string
		getService func(t *testing.T) *DeleteService
		request    *DeleteRequest
		expect     func(t *testing.T, err error)
	}{
		{
			name: 	"delete existing",
			getService: func(t *testing.T) *DeleteService {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &DeleteService{
					Logger:   log.None(),
					Database: database,
					ServiceProviderConfig: spc,
				}
			},
			request: &DeleteRequest{
				ResourceID:    "a5866759-32ca-4e2a-9808-a0fe74f94b18",
			},
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: 	"delete non-existing",
			getService: func(t *testing.T) *DeleteService {
				database := db.Memory()
				err := database.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return &DeleteService{
					Logger:   log.None(),
					Database: database,
					ServiceProviderConfig: spc,
				}
			},
			request: &DeleteRequest{
				ResourceID:    "foobar",
			},
			expect: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.getService(t)
			err := service.DeleteResource(context.Background(), test.request)
			test.expect(t, err)
		})
	}
}

func (s *DeleteServiceTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *DeleteServiceTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *DeleteServiceTestSuite) mustSchema(filePath string) *spec.Schema {
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

func (s *DeleteServiceTestSuite) mustServiceProviderConfig(filePath string) *spec.ServiceProviderConfig {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	spc := new(spec.ServiceProviderConfig)
	err = json.Unmarshal(raw, spc)
	s.Require().Nil(err)

	return spc
}
