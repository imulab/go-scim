package json

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

func TestJsonDeserialize(t *testing.T) {
	s := new(JsonDeserializeTestSuite)
	suite.Run(t, s)
}

type JsonDeserializeTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *JsonDeserializeTestSuite) TestDeserializeResource() {
	tests := []struct {
		name   string
		json   string
		expect func(t *testing.T, resource *prop.Resource, err error)
	}{
		{
			name: "default",
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
				for _, each := range []struct {
					expected interface{}
					path     []interface{}
				}{
					{expected: "urn:ietf:params:scim:schemas:core:2.0:User", path: []interface{}{"schemas", 0}},
					{expected: "3cc032f5-2361-417f-9e2f-bc80adddf4a3", path: []interface{}{"id"}},
					{expected: "User", path: []interface{}{"meta", "resourceType"}},
					{expected: "2019-11-20T13:09:00", path: []interface{}{"meta", "created"}},
					{expected: "2019-11-20T13:09:00", path: []interface{}{"meta", "lastModified"}},
					{expected: "https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3", path: []interface{}{"meta", "location"}},
					{expected: "W/\"1\"", path: []interface{}{"meta", "version"}},
					{expected: "imulab", path: []interface{}{"userName"}},
					{expected: "Mr. Weinan Qiu", path: []interface{}{"name", "formatted"}},
					{expected: "Qiu", path: []interface{}{"name", "familyName"}},
					{expected: "Weinan", path: []interface{}{"name", "givenName"}},
					{expected: "Mr.", path: []interface{}{"name", "honorificPrefix"}},
					{expected: "Weinan", path: []interface{}{"displayName"}},
					{expected: "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3", path: []interface{}{"profileUrl"}},
					{expected: "Employee", path: []interface{}{"userType"}},
					{expected: "zh_CN", path: []interface{}{"preferredLanguage"}},
					{expected: "zh_CN", path: []interface{}{"locale"}},
					{expected: "Asia/Shanghai", path: []interface{}{"timezone"}},
					{expected: true, path: []interface{}{"active"}},
					{expected: "imulab@foo.com", path: []interface{}{"emails", 0, "value"}},
					{expected: "work", path: []interface{}{"emails", 0, "type"}},
					{expected: true, path: []interface{}{"emails", 0, "primary"}},
					{expected: "imulab@foo.com", path: []interface{}{"emails", 0, "display"}},
					{expected: "imulab@bar.com", path: []interface{}{"emails", 1, "value"}},
					{expected: "home", path: []interface{}{"emails", 1, "type"}},
					{expected: nil, path: []interface{}{"emails", 1, "primary"}},
					{expected: "imulab@bar.com", path: []interface{}{"emails", 1, "display"}},
					{expected: "123-45678", path: []interface{}{"phoneNumbers", 0, "value"}},
					{expected: "work", path: []interface{}{"phoneNumbers", 0, "type"}},
					{expected: true, path: []interface{}{"phoneNumbers", 0, "primary"}},
					{expected: "123-45678", path: []interface{}{"phoneNumbers", 0, "display"}},
					{expected: "123-45679", path: []interface{}{"phoneNumbers", 1, "value"}},
					{expected: "work", path: []interface{}{"phoneNumbers", 1, "type"}},
					{expected: nil, path: []interface{}{"phoneNumbers", 1, "primary"}},
					{expected: "123-45679", path: []interface{}{"phoneNumbers", 1, "display"}},
				} {
					nav := resource.Navigator()
					for _, p := range each.path {
						switch q := p.(type) {
						case string:
							nav.Dot(q)
						case int:
							nav.At(q)
						default:
							panic("unsupported")
						}
					}
					assert.Nil(t, nav.Error())
					if each.expected == nil {
						assert.Nil(t, nav.Current().Raw())
					} else {
						assert.Equal(t, each.expected, nav.Current().Raw())
					}
				}
			},
		},
		{
			name: "empty json",
			json: "{}",
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.True(t, resource.RootProperty().IsUnassigned())
			},
		},
		{
			name: "explicit nulls",
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
				assert.False(t, resource.RootProperty().IsUnassigned())
				for _, each := range []struct {
					path []interface{}
				}{
					{path: []interface{}{"id"}},
					{path: []interface{}{"name", "givenName"}},
					{path: []interface{}{"emails", 0, "value"}},
				} {
					nav := resource.Navigator()
					for _, p := range each.path {
						switch q := p.(type) {
						case string:
							nav.Dot(q)
						case int:
							nav.At(q)
						default:
							panic("unsupported")
						}
					}
					assert.Nil(t, nav.Error())
					assert.Nil(t, nav.Current().Raw())
				}
			},
		},
		{
			name: "empty array",
			json: `
{
	"id": "foobar",
	"emails":[],
	"timezone":"Asia/Shanghai"
}`,
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", resource.Navigator().Dot("id").Current().Raw())
				assert.True(t, resource.Navigator().Dot("emails").Current().IsUnassigned())
				assert.Equal(t, "Asia/Shanghai", resource.Navigator().Dot("timezone").Current().Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := prop.NewResource(s.resourceType)
			err := Deserialize([]byte(test.json), resource)
			test.expect(t, resource, err)
		})
	}
}

func (s *JsonDeserializeTestSuite) TestDeserializeProperty() {
	tests := []struct {
		name   string
		attr   string
		json   string
		expect func(t *testing.T, property prop.Property, err error)
	}{
		{
			name: "deserialize string property",
			attr: `
{
	"id": "urn:ietf:params:scim:schemas:core:2.0:User:userName",
	"name": "userName",
	"type": "string"
}
`,
			json: `"imulab"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "imulab", property.Raw())
			},
		},
		{
			name: "deserialize integer property",
			attr: `
{
	"name": "age",
	"type": "integer"
}
`,
			json: `18`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(18), property.Raw())
			},
		},
		{
			name: "deserialize decimal property",
			attr: `
{
	"name": "score",
	"type": "decimal"
}
`,
			json: `123.123`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 123.123, property.Raw())
			},
		},
		{
			name: "deserialize boolean property",
			attr: `
{
	"name": "active",
	"type": "boolean"
}
`,
			json: `true`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, true, property.Raw())
			},
		},
		{
			name: "deserialize dateTime property",
			attr: `
{
	"name": "lastModified",
	"type": "dateTime"
}
`,
			json: `"2019-12-04T13:10:00"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "2019-12-04T13:10:00", property.Raw())
			},
		},
		{
			name: "deserialize reference property",
			attr: `
{
	"name": "link",
	"type": "reference"
}
`,
			json: `"http://imulab.io"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "http://imulab.io", property.Raw())
			},
		},
		{
			name: "deserialize binary property",
			attr: `
{
	"name": "certificate",
	"type": "binary"
}
`,
			json: `"aGVsbG8K"`,
			expect: func(t *testing.T, property prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "aGVsbG8K", property.Raw())
			},
		},
		{
			name: "deserialize complex property",
			attr: `
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
`,
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
			attr: `
{
 	"name": "collection",
 	"type": "string",
	"multiValued": true
}
`,
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
			attr: `
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
`,
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
			attr: `
{
 	"name": "collection",
 	"type": "string",
	"multiValued": true
}
`,
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
			attr: `
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
`,
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
			property := s.propForAttr(t, test.attr)
			err := DeserializeProperty([]byte(test.json), property, true)
			test.expect(t, property, err)
		})
	}
}

func (s *JsonDeserializeTestSuite) SetupSuite() {
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
			filepath:  "../../../public/schemas/user_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../public/resource_types/user_resource_type.json",
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

func (s *JsonDeserializeTestSuite) propForAttr(t *testing.T, attrJson string) prop.Property {
	attr := new(spec.Attribute)
	assert.Nil(t, json.Unmarshal([]byte(attrJson), attr))
	return prop.NewProperty(attr)
}
