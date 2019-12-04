package json

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestDeserialize(t *testing.T) {
	s := new(JSONDeserializeTestSuite)
	s.resourceBase = "../../tests/json_deserialize_test_suite"
	suite.Run(t, s)
}

type JSONDeserializeTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *JSONDeserializeTestSuite) TestDeserialize() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		json        string
		expect      func(t *testing.T, resource *prop.Resource, err error)
	}{
		{
			name: "default json deserialization",
			getResource: func(t *testing.T) *prop.Resource {
				_ = s.mustSchema("/user_schema.json")
				return prop.NewResource(s.mustResourceType("/user_resource_type.json"))
			},
			json: `
{
   "schemas":[
      "urn:ietf:params:scim:schemas:core:2.0:User"
   ],
   "id":"3cc032f5-2361-417f-9e2f-bc80adddf4a3",
   "meta":{
      "resourceType":"User",
      "created":"2019-11-20T13:09:00",
      "lastModified":"2019-11-20T13:09:00",
      "location":"https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
      "version":"W/\"1\""
   },
   "userName":"imulab",
   "name":{
      "formatted":"Mr. Weinan Qiu",
      "familyName":"Qiu",
      "givenName":"Weinan",
      "honorificPrefix":"Mr."
   },
   "displayName":"Weinan",
   "profileUrl":"https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
   "userType":"Employee",
   "preferredLanguage":"zh_CN",
   "locale":"zh_CN",
   "timezone":"Asia/Shanghai",
   "active":true,
   "emails":[
      {
         "value":"imulab@foo.com",
         "type":"work",
         "primary":true,
         "display":"imulab@foo.com"
      },
      {
         "value":"imulab@bar.com",
         "type":"home",
         "display":"imulab@bar.com"
      }
   ],
   "phoneNumbers":[
      {
         "value":"123-45678",
         "type":"work",
         "primary":true,
         "display":"123-45678"
      },
      {
         "value":"123-45679",
         "type":"work",
         "display":"123-45679"
      }
   ]
}
`,
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)
				nav := resource.NewNavigator()

				{
					_, _ = nav.FocusName("schemas")
					{
						_, _ = nav.FocusIndex(0)
						assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", nav.Current().Raw())
						nav.Retract()
					}
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("id")
					assert.Equal(t, "3cc032f5-2361-417f-9e2f-bc80adddf4a3", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("meta")
					{
						_, _ = nav.FocusName("resourceType")
						assert.Equal(t, "User", nav.Current().Raw())
						nav.Retract()
					}
					{
						_, _ = nav.FocusName("created")
						assert.Equal(t, "2019-11-20T13:09:00", nav.Current().Raw())
						nav.Retract()
					}
					{
						_, _ = nav.FocusName("lastModified")
						assert.Equal(t, "2019-11-20T13:09:00", nav.Current().Raw())
						nav.Retract()
					}
					{
						_, _ = nav.FocusName("location")
						assert.Equal(t, "https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3", nav.Current().Raw())
						nav.Retract()
					}
					{
						_, _ = nav.FocusName("version")
						assert.Equal(t, "W/\\\"1\\\"", nav.Current().Raw())
						nav.Retract()
					}
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("userName")
					assert.Equal(t, "imulab", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("name")
					{
						_, _ = nav.FocusName("formatted")
						assert.Equal(t, "Mr. Weinan Qiu", nav.Current().Raw())
						nav.Retract()
					}
					{
						_, _ = nav.FocusName("familyName")
						assert.Equal(t, "Qiu", nav.Current().Raw())
						nav.Retract()
					}
					{
						_, _ = nav.FocusName("givenName")
						assert.Equal(t, "Weinan", nav.Current().Raw())
						nav.Retract()
					}
					{
						_, _ = nav.FocusName("honorificPrefix")
						assert.Equal(t, "Mr.", nav.Current().Raw())
						nav.Retract()
					}
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("displayName")
					assert.Equal(t, "Weinan", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("profileUrl")
					assert.Equal(t, "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("userType")
					assert.Equal(t, "Employee", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("preferredLanguage")
					assert.Equal(t, "zh_CN", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("locale")
					assert.Equal(t, "zh_CN", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("timezone")
					assert.Equal(t, "Asia/Shanghai", nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("active")
					assert.Equal(t, true, nav.Current().Raw())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("emails")
					{
						_, _ = nav.FocusIndex(0)
						{
							{
								_, _ = nav.FocusName("value")
								assert.Equal(t, "imulab@foo.com", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("type")
								assert.Equal(t, "work", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("primary")
								assert.Equal(t, true, nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("display")
								assert.Equal(t, "imulab@foo.com", nav.Current().Raw())
								nav.Retract()
							}
						}
						nav.Retract()
					}
					{
						_, _ = nav.FocusIndex(1)
						{
							{
								_, _ = nav.FocusName("value")
								assert.Equal(t, "imulab@bar.com", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("type")
								assert.Equal(t, "home", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("display")
								assert.Equal(t, "imulab@bar.com", nav.Current().Raw())
								nav.Retract()
							}
						}
						nav.Retract()
					}
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("phoneNumbers")
					{
						_, _ = nav.FocusIndex(0)
						{
							{
								_, _ = nav.FocusName("value")
								assert.Equal(t, "123-45678", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("type")
								assert.Equal(t, "work", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("primary")
								assert.Equal(t, true, nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("display")
								assert.Equal(t, "123-45678", nav.Current().Raw())
								nav.Retract()
							}
						}
						nav.Retract()
					}
					{
						_, _ = nav.FocusIndex(1)
						{
							{
								_, _ = nav.FocusName("value")
								assert.Equal(t, "123-45679", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("type")
								assert.Equal(t, "work", nav.Current().Raw())
								nav.Retract()
							}
							{
								_, _ = nav.FocusName("display")
								assert.Equal(t, "123-45679", nav.Current().Raw())
								nav.Retract()
							}
						}
						nav.Retract()
					}
					nav.Retract()
				}
			},
		},
		{
			name: "empty json",
			getResource: func(t *testing.T) *prop.Resource {
				_ = s.mustSchema("/user_schema.json")
				return prop.NewResource(s.mustResourceType("/user_resource_type.json"))
			},
			json: "{}",
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)
				nav := resource.NewNavigator()
				_, _ = nav.FocusName("id")
				assert.Nil(t, nav.Current().Raw())
				assert.False(t, nav.Current().Touched())
			},
		},
		{
			name: "explicit nulls",
			getResource: func(t *testing.T) *prop.Resource {
				_ = s.mustSchema("/user_schema.json")
				return prop.NewResource(s.mustResourceType("/user_resource_type.json"))
			},
			json: `
{
	"id": null,
	"name": {
		"givenName": null
	},
	"emails": [
		{
			"value": null
		}
	]
}
`,
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)
				nav := resource.NewNavigator()
				{
					_, _ = nav.FocusName("id")
					assert.Nil(t, nav.Current().Raw())
					assert.True(t, nav.Current().Touched())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("name")
					{
						_, _ = nav.FocusName("givenName")
						assert.Nil(t, nav.Current().Raw())
						assert.True(t, nav.Current().Touched())
						nav.Retract()
					}
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("emails")
					{
						_, _ = nav.FocusIndex(0)
						{
							_, _ = nav.FocusName("value")
							assert.Nil(t, nav.Current().Raw())
							assert.True(t, nav.Current().Touched())
							nav.Retract()
						}
						nav.Retract()
					}
					nav.Retract()
				}
				{
					// in contrast, other fields was not touched
					_, _ = nav.FocusName("userName")
					assert.Nil(t, nav.Current().Raw())
					assert.False(t, nav.Current().Touched())
					nav.Retract()
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			err := Deserialize([]byte(test.json), resource)
			test.expect(t, resource, err)
		})
	}
}

func (s *JSONDeserializeTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *JSONDeserializeTestSuite) mustSchema(filePath string) *core.Schema {
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
