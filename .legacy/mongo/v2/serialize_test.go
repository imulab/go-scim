package v2

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"os"
	"testing"
)

func TestSerialize(t *testing.T) {
	s := new(MongoSerializerTestSuite)
	suite.Run(t, s)
}

type MongoSerializerTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *MongoSerializerTestSuite) TestSerialize() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
	}{
		{
			name: "serialize user",
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
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			raw, err := newBsonAdapter(test.getResource(t)).MarshalBSON()
			assert.Nil(t, err)
			assert.Nil(t, bson.Raw(raw).Validate())
		})
	}
}

func (s *MongoSerializerTestSuite) SetupSuite() {
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
