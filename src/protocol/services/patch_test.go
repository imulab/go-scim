package services

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/expr"
	scimJSON "github.com/imulab/go-scim/src/core/json"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/stretchr/testify/assert"
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

func (s *PatchServiceTestSuite) mustResource(filePath string, resourceType *core.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *PatchServiceTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *PatchServiceTestSuite) mustSchema(filePath string) *core.Schema {
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
