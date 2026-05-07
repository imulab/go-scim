package groupsync

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestCompare(t *testing.T) {
	s := new(CompareTestSuite)
	suite.Run(t, s)
}

type CompareTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *CompareTestSuite) TestCompare() {
	tests := []struct {
		name      string
		getBefore func(t *testing.T) *prop.Resource
		getAfter  func(t *testing.T) *prop.Resource
		expect    func(t *testing.T, diff *Diff)
	}{
		{
			name: "no modification",
			getBefore: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
					"id":      "foobar",
					"members": []interface{}{
						map[string]interface{}{
							"value":   "m1",
							"$ref":    "/Users/m1",
							"display": "m1",
						},
						map[string]interface{}{
							"value":   "m2",
							"$ref":    "/Users/m2",
							"display": "m2",
						},
					},
				}).HasError())
				return r
			},
			getAfter: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
					"id":      "foobar",
					"members": []interface{}{
						map[string]interface{}{
							"value":   "m2",
							"$ref":    "/Users/m2",
							"display": "m2",
						},
						map[string]interface{}{
							"value":   "m1",
							"$ref":    "/Users/m1",
							"display": "m1",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, diff *Diff) {
				assert.Equal(t, 0, diff.CountLeft())
				assert.Equal(t, 0, diff.CountJoined())
			},
		},
		{
			name: "someone joined",
			getBefore: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
					"id":      "foobar",
					"members": []interface{}{
						map[string]interface{}{
							"value":   "m1",
							"$ref":    "/Users/m1",
							"display": "m1",
						},
					},
				}).HasError())
				return r
			},
			getAfter: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
					"id":      "foobar",
					"members": []interface{}{
						map[string]interface{}{
							"value":   "m2",
							"$ref":    "/Users/m2",
							"display": "m2",
						},
						map[string]interface{}{
							"value":   "m1",
							"$ref":    "/Users/m1",
							"display": "m1",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, diff *Diff) {
				assert.Equal(t, 0, diff.CountLeft())
				assert.Equal(t, 1, diff.CountJoined())
				_, m2Joined := diff.joined["m2"]
				assert.True(t, m2Joined)
			},
		},
		{
			name: "someone left",
			getBefore: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
					"id":      "foobar",
					"members": []interface{}{
						map[string]interface{}{
							"value":   "m1",
							"$ref":    "/Users/m1",
							"display": "m1",
						},
						map[string]interface{}{
							"value":   "m2",
							"$ref":    "/Users/m2",
							"display": "m2",
						},
					},
				}).HasError())
				return r
			},
			getAfter: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
					"id":      "foobar",
					"members": []interface{}{
						map[string]interface{}{
							"value":   "m2",
							"$ref":    "/Users/m2",
							"display": "m2",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, diff *Diff) {
				assert.Equal(t, 1, diff.CountLeft())
				assert.Equal(t, 0, diff.CountJoined())
				_, m1Left := diff.left["m1"]
				assert.True(t, m1Left)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			before := test.getBefore(t)
			after := test.getAfter(t)
			diff := Compare(before, after)
			test.expect(t, diff)
		})
	}
}

func (s *CompareTestSuite) SetupSuite() {
	for _, each := range []struct {
		filepath  string
		structure interface{}
		post      func(parsed interface{})
	}{
		{
			filepath:  "../../../public/schemas/core_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../public/schemas/group_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../public/resource_types/group_resource_type.json",
			structure: new(spec.ResourceType),
			post: func(parsed interface{}) {
				s.resourceType = parsed.(*spec.ResourceType)
			},
		},
	} {
		f, err := os.Open(each.filepath)
		require.Nil(s.T(), err)

		raw, err := ioutil.ReadAll(f)
		require.Nil(s.T(), err)

		err = json.Unmarshal(raw, each.structure)
		require.Nil(s.T(), err)

		if each.post != nil {
			each.post(each.structure)
		}
	}
}
