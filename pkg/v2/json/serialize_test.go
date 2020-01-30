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

func TestJsonSerialize(t *testing.T) {
	s := new(JsonSerializeTestSuite)
	suite.Run(t, s)
}

type JsonSerializeTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
	resourceData interface{}
}

func (s *JsonSerializeTestSuite) TestSerialize() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		options     []Options
		expect      func(t *testing.T, raw []byte, err error)
	}{
		{
			name: "default",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				_, err := r.RootProperty().Replace(s.resourceData)
				assert.Nil(t, err)
				return r
			},
			options: []Options{},
			expect: func(t *testing.T, raw []byte, err error) {
				assert.Nil(t, err)
				expect := `
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
`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name: "include attributes",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				_, err := r.RootProperty().Replace(s.resourceData)
				assert.Nil(t, err)
				return r
			},
			options: []Options{
				Include("emails.value", "emails.type"),
			},
			expect: func(t *testing.T, raw []byte, err error) {
				assert.Nil(t, err)
				expect := `
{
   "schemas":[
      "urn:ietf:params:scim:schemas:core:2.0:User"
   ],
   "id":"3cc032f5-2361-417f-9e2f-bc80adddf4a3",
   "emails":[
      {
         "value":"imulab@foo.com",
         "type":"work"
      },
      {
         "value":"imulab@bar.com",
         "type":"home"
      }
   ]
}
`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name: "exclude attributes",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				_, err := r.RootProperty().Replace(s.resourceData)
				assert.Nil(t, err)
				return r
			},
			options: []Options{
				Exclude("emails.type", "emails.value", "emails.primary"),
				Exclude("phoneNumbers", "ims", "groups", "entitlements"),
				Exclude("roles", "photos", "addresses", "x509Certificates"),
			},
			expect: func(t *testing.T, raw []byte, err error) {
				assert.Nil(t, err)
				expect := `
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
         "display":"imulab@foo.com"
      },
      {
         "display":"imulab@bar.com"
      }
   ]
}
`
				assert.JSONEq(t, expect, string(raw))
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			raw, err := Serialize(resource, test.options...)
			test.expect(t, raw, err)
		})
	}
}

func (s *JsonSerializeTestSuite) SetupSuite() {
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

	s.resourceData = map[string]interface{}{
		"schemas": []interface{}{
			"urn:ietf:params:scim:schemas:core:2.0:User",
		},
		"id": "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
		"meta": map[string]interface{}{
			"resourceType": "User",
			"created":      "2019-11-20T13:09:00",
			"lastModified": "2019-11-20T13:09:00",
			"location":     "https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
			"version":      "W/\"1\"",
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
		"phoneNumbers": []interface{}{
			map[string]interface{}{
				"value":   "123-45678",
				"type":    "work",
				"primary": true,
				"display": "123-45678",
			},
			map[string]interface{}{
				"value":   "123-45679",
				"type":    "work",
				"display": "123-45679",
			},
		},
	}
}
