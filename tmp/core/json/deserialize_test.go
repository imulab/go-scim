package json

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestDeserialize(t *testing.T) {
	s := new(JSONDeserializeTestSuite)
	s.resourceBase = "../internal/json_deserialize_test_suite"
	suite.Run(t, s)
}

type JSONDeserializeTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *JSONDeserializeTestSuite) TestDeserializeProperty() {
	tests := []struct {
		name        string
		getProperty func(t *testing.T) prop.Property
		json        string
		expect      func(t *testing.T, property prop.Property, err error)
	}{
		{
			name: "deserialize string property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewString(s.mustAttribute(`
{
	"id": "urn:ietf:params:scim:schemas:core:2.0:User:userName",
	"name": "userName",
	"type": "string"
}
`), nil)
			},
			json: `"imulab"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "imulab", property.Raw())
			},
		},
		{
			name: "deserialize integer property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), nil)
			},
			json: `18`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(18), property.Raw())
			},
		},
		{
			name: "deserialize decimal property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewDecimal(s.mustAttribute(`
{
	"name": "score",
	"type": "decimal"
}
`), nil)
			},
			json: `123.123`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 123.123, property.Raw())
			},
		},
		{
			name: "deserialize boolean property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewBoolean(s.mustAttribute(`
{
	"name": "active",
	"type": "boolean"
}
`), nil)
			},
			json: `true`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, true, property.Raw())
			},
		},
		{
			name: "deserialize dateTime property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewDateTime(s.mustAttribute(`
{
	"name": "lastModified",
	"type": "dateTime"
}
`), nil)
			},
			json: `"2019-12-04T13:10:00"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "2019-12-04T13:10:00", property.Raw())
			},
		},
		{
			name: "deserialize reference property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewReference(s.mustAttribute(`
{
	"name": "link",
	"type": "reference"
}
`), nil)
			},
			json: `"http://imulab.io"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "http://imulab.io", property.Raw())
			},
		},
		{
			name: "deserialize binary property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewBinary(s.mustAttribute(`
{
	"name": "certificate",
	"type": "binary"
}
`), nil)
			},
			json: `"aGVsbG8K"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "aGVsbG8K", property.Raw())
			},
		},
		{
			name: "deserialize complex property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewComplex(s.mustAttribute(`
{
 	"id": "urn:ietf:params:scim:schemas:core:2.0:User:name",
 	"name": "name",
 	"type": "complex",
 	"subAttributes": [
		{
	  		"id": "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
	  		"name": "familyName",
	  		"type": "string",
	  		"_path": "name.familyName"
		},
		{
	  		"id": "urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
	  		"name": "givenName",
	  		"type": "string",
	  		"_path": "name.givenName"
		}
 	],
 	"_path": "name"
}
`), nil)
			},
			json: `{
	"familyName": "Qiu",
	"givenName": "David"
}`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				v := property.Raw()
				assert.IsType(t, map[string]interface{}{}, v)
				assert.Equal(t, "Qiu", v.(map[string]interface{})["familyName"])
				assert.Equal(t, "David", v.(map[string]interface{})["givenName"])
			},
		},
		{
			name: "deserialize multiValued property containing simple properties",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMulti(s.mustAttribute(`
{
 	"name": "collection",
 	"type": "string",
	"multiValued": true
}
`), nil)
			},
			json: `["A", "B", "C"]`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				v := property.Raw()
				assert.IsType(t, []interface{}{}, v)
				assert.Equal(t, "A", v.([]interface{})[0])
				assert.Equal(t, "B", v.([]interface{})[1])
				assert.Equal(t, "C", v.([]interface{})[2])
			},
		},
		{
			name: "deserialize multiValued property containing complex properties",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMulti(s.mustAttribute(`
{
  	"name": "emails",
  	"type": "complex",
	"multiValued": true,
	"subAttributes": [
		{
			"name": "value",
  			"type": "string"
		},
		{
			"name": "type",
  			"type": "string"
		}
	]
}
`), nil)
			},
			json: `[
	{
		"value": "foo@bar.com",
		"type": "work"
	},
	{
		"value": "bar@foo.com",
		"type": "home"
	}
]`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				v := property.Raw()
				assert.IsType(t, []interface{}{}, v)
				{
					assert.IsType(t, map[string]interface{}{}, v.([]interface{})[0])
					v0 := v.([]interface{})[0].(map[string]interface{})
					assert.Equal(t, "foo@bar.com", v0["value"])
					assert.Equal(t, "work", v0["type"])
				}
				{
					assert.IsType(t, map[string]interface{}{}, v.([]interface{})[1])
					v1 := v.([]interface{})[1].(map[string]interface{})
					assert.Equal(t, "bar@foo.com", v1["value"])
					assert.Equal(t, "home", v1["type"])
				}
			},
		},
		{
			name: "deserialize multiValued property with an element",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMulti(s.mustAttribute(`
{
 	"name": "collection",
 	"type": "string",
	"multiValued": true
}
`), nil)
			},
			json: `"A"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				v := property.Raw()
				assert.IsType(t, []interface{}{}, v)
				assert.Len(t, v, 1)
				assert.Equal(t, "A", v.([]interface{})[0])
			},
		},
		{
			name: "deserialize complex multiValued property with an element",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMulti(s.mustAttribute(`
{
  	"name": "emails",
  	"type": "complex",
	"multiValued": true,
	"subAttributes": [
		{
			"name": "value",
  			"type": "string"
		},
		{
			"name": "type",
  			"type": "string"
		}
	]
}
`), nil)
			},
			json: `{
	"value": "foo@bar.com",
	"type": "work"
}`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				v := property.Raw()
				assert.IsType(t, []interface{}{}, v)
				assert.Len(t, v, 1)
				{
					assert.IsType(t, map[string]interface{}{}, v.([]interface{})[0])
					v0 := v.([]interface{})[0].(map[string]interface{})
					assert.Equal(t, "foo@bar.com", v0["value"])
					assert.Equal(t, "work", v0["type"])
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			property := test.getProperty(t)
			err := DeserializeProperty([]byte(test.json), property, true)
			test.expect(t, property, err)
		})
	}
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
						assert.Equal(t, "W/\"1\"", nav.Current().Raw())
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
				assert.False(t, nav.Current().Dirty())
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
					assert.True(t, nav.Current().Dirty())
					nav.Retract()
				}
				{
					_, _ = nav.FocusName("name")
					{
						_, _ = nav.FocusName("givenName")
						assert.Nil(t, nav.Current().Raw())
						assert.True(t, nav.Current().Dirty())
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
							assert.True(t, nav.Current().Dirty())
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
					assert.False(t, nav.Current().Dirty())
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

func (s *JSONDeserializeTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *JSONDeserializeTestSuite) mustSchema(filePath string) *spec.Schema {
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

func (s *JSONDeserializeTestSuite) mustAttribute(attrJSON string) *spec.Attribute {
	raw, err := ioutil.ReadAll(strings.NewReader(attrJSON))
	s.Require().Nil(err)

	attr := new(spec.Attribute)
	err = json.Unmarshal(raw, attr)
	s.Require().Nil(err)

	return attr
}
