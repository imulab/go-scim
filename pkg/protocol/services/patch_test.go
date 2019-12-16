package services

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/expr"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/lock"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/imulab/go-scim/pkg/protocol/services/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestPatchService(t *testing.T) {
	s := new(PatchServiceTestSuite)
	s.resourceBase = "../../tests/patch_service_test_suite"
	suite.Run(t, s)
}

type PatchServiceTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *PatchServiceTestSuite) TestPatchService() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	expr.Register(resourceType)
	spc := s.mustServiceProviderConfig("/service_provider_config.json")

	tests := []struct{
		name		string
		setup      func(t *testing.T) *PatchService
		getRequest func(t *testing.T) *PatchRequest
		expect     func(t *testing.T, resp *PatchResponse, err error)
	}{
		{
			name: 	"patch to make a difference",
			setup: func(t *testing.T) *PatchService {
				memoryDB := db.Memory()
				require.Nil(t, memoryDB.Insert(
					context.Background(),
					s.mustResource("/user_000.json", resourceType)),
				)
				return &PatchService{
					Logger:           log.None(),
					Lock:             lock.Default(),
					PrePatchFilters: []filter.ForResource{},
					PostPatchFilters: []filter.ForResource{
						filter.CopyReadOnly(),
						filter.Password(10),
						filter.Validation(memoryDB),
						filter.Meta(),
					},
					Database: memoryDB,
					ServiceProviderConfig: spc,
				}
			},
			getRequest: func(t *testing.T) *PatchRequest {
				req := new(PatchRequest)
				err := json.Unmarshal([]byte(`
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "userName",
			"value": "foobar"
		},
		{
			"op": "replace",
			"path": "emails[type eq \"home\"].type",
			"value": "work"
		},
		{
			"op": "remove",
			"path": "timezone"
		}
	]
}
`), req)
				require.Nil(t, err)
				req.ResourceID = "3cc032f5-2361-417f-9e2f-bc80adddf4a3"
				return req
			},
			expect: func(t *testing.T, resp *PatchResponse, err error) {
				assert.Nil(t, err)
				assert.NotEqual(t, resp.OldVersion, resp.NewVersion)

				nav := resp.Resource.NewNavigator()
				{
					_, err := nav.FocusName("userName")
					assert.Nil(t, err)
					assert.Equal(t, "foobar", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, err := nav.FocusName("timezone")
					assert.Nil(t, err)
					assert.Nil(t, nav.Current().Raw())
					nav.Retract()
				}
				{
					_, err := nav.FocusName("emails")
					assert.Nil(t, err)
					_ = nav.Current().(prop.Container).ForEachChild(func(index int, child prop.Property) error {
						assert.Equal(t, "work", child.(prop.Container).ChildAtIndex("type").Raw())
						return nil
					})
					nav.Retract()
				}
			},
		},
		{
			name: 	"patch to not make a difference",
			setup: func(t *testing.T) *PatchService {
				memoryDB := db.Memory()
				require.Nil(t, memoryDB.Insert(
					context.Background(),
					s.mustResource("/user_000.json", resourceType)),
				)
				return &PatchService{
					Logger:           log.None(),
					Lock:             lock.Default(),
					PrePatchFilters: []filter.ForResource{},
					PostPatchFilters: []filter.ForResource{
						filter.CopyReadOnly(),
						filter.Password(10),
						filter.Validation(memoryDB),
						filter.Meta(),
					},
					Database: memoryDB,
					ServiceProviderConfig: spc,
				}
			},
			getRequest: func(t *testing.T) *PatchRequest {
				req := new(PatchRequest)
				err := json.Unmarshal([]byte(`
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "userName",
			"value": "imulab"
		}
	]
}
`), req)
				require.Nil(t, err)
				req.ResourceID = "3cc032f5-2361-417f-9e2f-bc80adddf4a3"
				return req
			},
			expect: func(t *testing.T, resp *PatchResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, resp.OldVersion, resp.NewVersion)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			request := test.getRequest(t)
			response, err := service.PatchResource(context.Background(), request)
			test.expect(t, response, err)
		})
	}
}

func (s *PatchServiceTestSuite) TestParsePayload() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	expr.Register(resourceType)
	resource := s.mustResource("/user_001.json", resourceType)

	tests := []struct{
		name		string
		payload		string
		expect		func(t *testing.T, pr *PatchRequest)
	}{
		{
			name: 	"add patch request",
			payload: `
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "userName",
			"value": "foobar"
		}
	]
}
`,
			expect: func(t *testing.T, pr *PatchRequest) {
				assert.Len(t, pr.Operations, 1)
				v, err := pr.Operations[0].ParseValue(resource)
				assert.Nil(t, err)
				assert.Equal(t, "foobar", v)
			},
		},
		{
			name: 	"add patch request (nested)",
			payload: `
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "name.familyName",
			"value": "foobar"
		}
	]
}
`,
			expect: func(t *testing.T, pr *PatchRequest) {
				assert.Len(t, pr.Operations, 1)
				v, err := pr.Operations[0].ParseValue(resource)
				assert.Nil(t, err)
				assert.Equal(t, "foobar", v)
			},
		},
		{
			name: 	"add patch request (complex)",
			payload: `
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "emails",
			"value": {
				"value": "test@test.org",
				"display": "test",
				"type": "work"
			}
		}
	]
}
`,
			expect: func(t *testing.T, pr *PatchRequest) {
				assert.Len(t, pr.Operations, 1)
				v, err := pr.Operations[0].ParseValue(resource)
				assert.Nil(t, err)
				assert.EqualValues(t, []interface{}{
					map[string]interface{}{
						"value": "test@test.org",
						"display": "test",
						"type": "work",
						"primary": nil,
					},
				}, v)
			},
		},
		{
			name: 	"add patch request (complex filter)",
			payload: `
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "emails[value eq \"foo@bar.com\"].type",
			"value": "work"
		}
	]
}
`,
			expect: func(t *testing.T, pr *PatchRequest) {
				assert.Len(t, pr.Operations, 1)
				v, err := pr.Operations[0].ParseValue(resource)
				assert.Nil(t, err)
				assert.Equal(t, "work", v)
			},
		},
		{
			name: 	"add invalid path",
			payload: `
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "foobar",
			"value": "work"
		}
	]
}
`,
			expect: func(t *testing.T, pr *PatchRequest) {
				assert.Len(t, pr.Operations, 1)
				_, err := pr.Operations[0].ParseValue(resource)
				assert.NotNil(t, err)
			},
		},
		{
			name: 	"add invalid value",
			payload: `
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "userName",
			"value": 18
		}
	]
}
`,
			expect: func(t *testing.T, pr *PatchRequest) {
				assert.Len(t, pr.Operations, 1)
				_, err := pr.Operations[0].ParseValue(resource)
				assert.NotNil(t, err)
			},
		},
		{
			name: 	"replace patch request (complex)",
			payload: `
{
	"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	"Operations": [
		{
			"op": "add",
			"path": "emails",
			"value": [
				{
					"value": "test@test.org",
					"display": "test",
					"type": "work"
				}
			]
		}
	]
}
`,
			expect: func(t *testing.T, pr *PatchRequest) {
				assert.Len(t, pr.Operations, 1)
				v, err := pr.Operations[0].ParseValue(resource)
				assert.Nil(t, err)
				assert.EqualValues(t, []interface{}{
					map[string]interface{}{
						"value": "test@test.org",
						"display": "test",
						"type": "work",
						"primary": nil,
					},
				}, v)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			pr := new(PatchRequest)
			err := json.Unmarshal([]byte(test.payload), pr)
			assert.Nil(t, err)
			assert.Nil(t, pr.Validate())
			test.expect(t, pr)
		})
	}
}

func (s *PatchServiceTestSuite) mustServiceProviderConfig(filePath string) *spec.ServiceProviderConfig {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	spc := new(spec.ServiceProviderConfig)
	err = json.Unmarshal(raw, spc)
	s.Require().Nil(err)

	return spc
}

func (s *PatchServiceTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *PatchServiceTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *PatchServiceTestSuite) mustSchema(filePath string) *spec.Schema {
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
