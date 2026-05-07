package v2

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

func TestDeserialize(t *testing.T) {
	s := new(MongoDeserializerTestSuite)
	suite.Run(t, s)
}

type MongoDeserializerTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *MongoDeserializerTestSuite) TestDeserialize() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "user resource",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{
						"urn:ietf:params:scim:schemas:core:2.0:User",
					},
					"userName": "imulab",
					"name": map[string]interface{}{
						"formatted":       "Mr. Weinan Qiu",
						"familyName":      "Qiu",
						"givenName":       "Weinan",
						"honorificPrefix": "Mr.",
					},
					"displayName":       "Weinan",
					"profileUrl":        "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
					"userType":          "Employee",
					"preferredLanguage": "zh_CN",
					"locale":            "zh_CN",
					"timezone":          "Asia/Shanghai",
					"active":            true,
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "imulab@foo.com",
							"type":    "work",
							"primary": true,
							"display": "imulab@foo.com",
						},
						map[string]interface{}{
							"value":   "imulab@bar.com",
							"type":    "home",
							"display": "imulab@bar.com",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				for _, each := range []struct {
					expected interface{}
					paths    []interface{}
				}{
					{expected: "urn:ietf:params:scim:schemas:core:2.0:User", paths: []interface{}{"schemas", 0}},
					{expected: nil, paths: []interface{}{"id"}},
					{expected: "imulab", paths: []interface{}{"userName"}},
					{expected: "Mr. Weinan Qiu", paths: []interface{}{"name", "formatted"}},
					{expected: "Weinan", paths: []interface{}{"name", "givenName"}},
					{expected: "Qiu", paths: []interface{}{"name", "familyName"}},
					{expected: "Mr.", paths: []interface{}{"name", "honorificPrefix"}},
					{expected: "Weinan", paths: []interface{}{"displayName"}},
					{expected: "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3", paths: []interface{}{"profileUrl"}},
					{expected: "Employee", paths: []interface{}{"userType"}},
					{expected: "zh_CN", paths: []interface{}{"preferredLanguage"}},
					{expected: "zh_CN", paths: []interface{}{"locale"}},
					{expected: "Asia/Shanghai", paths: []interface{}{"timezone"}},
					{expected: true, paths: []interface{}{"active"}},
					{expected: "imulab@foo.com", paths: []interface{}{"emails", 0, "value"}},
					{expected: "work", paths: []interface{}{"emails", 0, "type"}},
					{expected: true, paths: []interface{}{"emails", 0, "primary"}},
					{expected: "imulab@bar.com", paths: []interface{}{"emails", 1, "value"}},
					{expected: "home", paths: []interface{}{"emails", 1, "type"}},
					{expected: nil, paths: []interface{}{"emails", 1, "primary"}},
				} {
					nav := r.Navigator()
					for _, path := range each.paths {
						switch p := path.(type) {
						case string:
							nav.Dot(p)
						case int:
							nav.At(p)
						default:
							panic("unsupported")
						}
					}
					if each.expected == nil {
						assert.True(t, nav.Current().IsUnassigned())
					} else {
						assert.Equal(t, each.expected, nav.Current().Raw())
					}
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			raw, err := newBsonAdapter(resource).MarshalBSON()
			assert.Nil(t, err)
			um := newResourceUnmarshaler(resource.ResourceType())
			err = um.UnmarshalBSON(raw)
			test.expect(t, um.Resource(), err)
		})
	}
}

func (s *MongoDeserializerTestSuite) SetupSuite() {
	for _, each := range []struct {
		filepath  string
		structure interface{}
		post      func(parsed interface{})
	}{
		{
			filepath:  "../../public/schemas/core_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../public/schemas/user_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../public/resource_types/user_resource_type.json",
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
