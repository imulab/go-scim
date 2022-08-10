package scim

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSONDeserialize(t *testing.T) {
	type Omit struct{}

	rt := NewResourceType[Omit]("User").
		MainSchema(UserSchema).
		ExtendSchema(UserEnterpriseSchemaExtension, false).
		Build()

	for _, c := range []struct {
		name   string
		json   string
		assert func(t *testing.T, res *Resource[Omit], err error)
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
			assert: func(t *testing.T, res *Resource[Omit], err error) {
				if assert.NoError(t, err) {
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
						nav := res.navigator()
						for _, p := range each.path {
							switch q := p.(type) {
							case string:
								nav.dot(q)
							case int:
								nav.at(q)
							default:
								panic("unsupported")
							}
						}

						if assert.False(t, nav.hasError()) {
							if each.expected == nil {
								assert.Nil(t, nav.current().Value())
							} else {
								assert.Equal(t, each.expected, nav.current().Value())
							}
						}
					}
				}
			},
		},
		{
			name: "empty json",
			json: `{}`,
			assert: func(t *testing.T, res *Resource[Omit], err error) {
				if assert.NoError(t, err) {
					assert.True(t, res.root.Unassigned())
				}
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
			assert: func(t *testing.T, res *Resource[Omit], err error) {
				if assert.NoError(t, err) {
					for _, each := range []struct {
						path []interface{}
					}{
						{path: []interface{}{"id"}},
						{path: []interface{}{"name", "givenName"}},
						{path: []interface{}{"emails"}},
					} {
						nav := res.navigator()
						for _, p := range each.path {
							switch q := p.(type) {
							case string:
								nav.dot(q)
							case int:
								nav.at(q)
							default:
								panic("unsupported")
							}
						}
						if assert.False(t, nav.hasError()) {
							assert.True(t, nav.current().Unassigned())
						}
					}
				}
			},
		},
		{
			name: "empty array",
			json: `
{
	"emails":[]
}
`,
			assert: func(t *testing.T, res *Resource[Omit], err error) {
				if assert.NoError(t, err) {
					assert.True(t, res.navigator().dot("emails").current().Unassigned())
				}
			},
		},
		{
			name: "Microsoft AD boolean issue hack (pr/67)",
			json: `
{
  "schemas":[
     "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id":"3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "userName":"imulab",
  "active": "True"
}
`,
			assert: func(t *testing.T, res *Resource[Omit], err error) {
				if assert.NoError(t, err) {
					assert.Equal(t, true, res.navigator().dot("active").current().Value())
				}
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			res := rt.New()
			err := json.Unmarshal([]byte(c.json), &res)
			c.assert(t, res, err)
		})
	}
}
