package services

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	scimJSON "github.com/imulab/go-scim/src/core/json"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	"github.com/imulab/go-scim/src/protocol/filters"
	"github.com/imulab/go-scim/src/protocol/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestReplaceService(t *testing.T) {
	s := new(ReplaceServiceTestSuite)
	s.resourceBase = "../../tests/replace_service_test_suite"
	suite.Run(t, s)
}

type ReplaceServiceTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *ReplaceServiceTestSuite) TestReplace() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name       string
		setup      func(t *testing.T) *ReplaceService
		getRequest func() *ReplaceRequest
		expect     func(t *testing.T, resp *ReplaceResponse, err error)
	}{
		{
			name: "replace an updated resource",
			setup: func(t *testing.T) *ReplaceService {
				memoryPersistence := persistence.Memory()
				require.Nil(t, memoryPersistence.Insert(
					context.Background(),
					s.mustResource("/user_000.json", resourceType)),
				)
				return &ReplaceService{
					Logger: protocol.NoOpLogger(),
					Lock:   protocol.DefaultLock(),
					Filters: []protocol.ResourceFilter{
						filters.NewClearReadOnlyResourceFilter(),
						filters.NewCopyReadOnlyResourceFilter(),
						filters.NewPasswordResourceFilter(10),
						filters.NewValidationResourceFilter(memoryPersistence),
						filters.NewMetaResourceFilter(),
					},
					Persistence: memoryPersistence,
				}
			},
			getRequest: func() *ReplaceRequest {
				resource := s.mustResource("/user_001.json", resourceType)
				return &ReplaceRequest{
					ResourceID: "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
					Payload:    resource,
				}
			},
			expect: func(t *testing.T, resp *ReplaceResponse, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, resp.NewVersion)
				assert.NotEmpty(t, resp.OldVersion)
				assert.NotEqual(t, resp.OldVersion, resp.NewVersion)
			},
		},
		{
			name: "replace an identical resource",
			setup: func(t *testing.T) *ReplaceService {
				memoryPersistence := persistence.Memory()
				require.Nil(t, memoryPersistence.Insert(
					context.Background(),
					s.mustResource("/user_000.json", resourceType)),
				)
				return &ReplaceService{
					Logger: protocol.NoOpLogger(),
					Lock:   protocol.DefaultLock(),
					Filters: []protocol.ResourceFilter{
						filters.NewClearReadOnlyResourceFilter(),
						filters.NewCopyReadOnlyResourceFilter(),
						filters.NewPasswordResourceFilter(10),
						filters.NewValidationResourceFilter(memoryPersistence),
						filters.NewMetaResourceFilter(),
					},
					Persistence: memoryPersistence,
				}
			},
			getRequest: func() *ReplaceRequest {
				resource := s.mustResource("/user_002.json", resourceType)
				return &ReplaceRequest{
					ResourceID: "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
					Payload:    resource,
				}
			},
			expect: func(t *testing.T, resp *ReplaceResponse, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, resp.NewVersion)
				assert.NotEmpty(t, resp.OldVersion)
				assert.Equal(t, resp.OldVersion, resp.NewVersion)
			},
		},
		{
			name: "replace an invalid resource",
			setup: func(t *testing.T) *ReplaceService {
				memoryPersistence := persistence.Memory()
				require.Nil(t, memoryPersistence.Insert(
					context.Background(),
					s.mustResource("/user_000.json", resourceType)),
				)
				return &ReplaceService{
					Logger: protocol.NoOpLogger(),
					Lock:   protocol.DefaultLock(),
					Filters: []protocol.ResourceFilter{
						filters.NewClearReadOnlyResourceFilter(),
						filters.NewCopyReadOnlyResourceFilter(),
						filters.NewPasswordResourceFilter(10),
						filters.NewValidationResourceFilter(memoryPersistence),
						filters.NewMetaResourceFilter(),
					},
					Persistence: memoryPersistence,
				}
			},
			getRequest: func() *ReplaceRequest {
				resource := s.mustResource("/user_003.json", resourceType)
				return &ReplaceRequest{
					ResourceID: "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
					Payload:    resource,
				}
			},
			expect: func(t *testing.T, resp *ReplaceResponse, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			request := test.getRequest()
			response, err := service.ReplaceResource(context.Background(), request)
			test.expect(t, response, err)
		})
	}
}

func (s *ReplaceServiceTestSuite) mustResource(filePath string, resourceType *core.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *ReplaceServiceTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *ReplaceServiceTestSuite) mustSchema(filePath string) *core.Schema {
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
