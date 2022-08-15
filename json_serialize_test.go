package scim

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSONSerialize(t *testing.T) {
	type Omit struct{}

	rt := NewResourceType[Omit]("User").
		MainSchema(UserSchema).
		ExtendSchema(UserEnterpriseSchemaExtension, false).
		Build()

	data := map[string]any{
		"schemas": []any{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"id":      "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
		"meta": map[string]any{
			"resourceType": "User",
		},
		"userName": "imulab",
		"name": map[string]any{
			"formatted":  "Weinan Qiu",
			"familyName": "Qiu",
			"givenName":  "Weinan",
		},
		"displayName": "Weinan",
		"active":      true,
		"emails": []any{
			map[string]any{"value": "imulab@foo.com", "type": "work", "primary": true},
			map[string]any{"value": "imulab@bar.com", "type": "home"},
		},
	}

	for _, c := range []struct {
		name     string
		data     map[string]any
		includes []string
		excludes []string
		assert   func(t *testing.T, jsonBytes []byte, err error)
	}{
		{
			name: "default",
			data: data,
			assert: func(t *testing.T, jsonBytes []byte, err error) {
				if assert.NoError(t, err) {
					assert.JSONEq(t, `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "meta": {
    "resourceType": "User"
  },
  "userName": "imulab",
  "name": {
    "formatted": "Weinan Qiu",
    "familyName": "Qiu",
    "givenName": "Weinan"
  },
  "displayName": "Weinan",
  "active": true,
  "emails": [
    {
      "value": "imulab@foo.com",
      "type": "work",
      "primary": true
    },
    {
      "value": "imulab@bar.com",
      "type": "home"
    }
  ]
}
`, string(jsonBytes))
				}
			},
		},
		{
			name:     "include attributes",
			data:     data,
			includes: []string{"emails"},
			assert: func(t *testing.T, jsonBytes []byte, err error) {
				if assert.NoError(t, err) {
					assert.JSONEq(t, `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "emails": [
    {
      "value": "imulab@foo.com",
      "type": "work",
      "primary": true
    },
    {
      "value": "imulab@bar.com",
      "type": "home"
    }
  ]
}
`, string(jsonBytes))
				}
			},
		},
		{
			name:     "include attributes of sub path",
			data:     data,
			includes: []string{"emails.value"},
			assert: func(t *testing.T, jsonBytes []byte, err error) {
				if assert.NoError(t, err) {
					assert.JSONEq(t, `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "emails": [
    {
      "value": "imulab@foo.com"
    },
    {
      "value": "imulab@bar.com"
    }
  ]
}
`, string(jsonBytes))
				}
			},
		},
		{
			name:     "exclude attributes",
			data:     data,
			excludes: []string{"emails"},
			assert: func(t *testing.T, jsonBytes []byte, err error) {
				if assert.NoError(t, err) {
					assert.JSONEq(t, `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "meta": {
    "resourceType": "User"
  },
  "userName": "imulab",
  "name": {
    "formatted": "Weinan Qiu",
    "familyName": "Qiu",
    "givenName": "Weinan"
  },
  "displayName": "Weinan",
  "active": true
}
`, string(jsonBytes))
				}
			},
		},
		{
			name:     "exclude attributes with sub path",
			data:     data,
			excludes: []string{"emails.type", "emails.primary"},
			assert: func(t *testing.T, jsonBytes []byte, err error) {
				if assert.NoError(t, err) {
					assert.JSONEq(t, `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
  "meta": {
    "resourceType": "User"
  },
  "userName": "imulab",
  "name": {
    "formatted": "Weinan Qiu",
    "familyName": "Qiu",
    "givenName": "Weinan"
  },
  "displayName": "Weinan",
  "active": true,
  "emails": [
    {
      "value": "imulab@foo.com"
    },
    {
      "value": "imulab@bar.com"
    }
  ]
}
`, string(jsonBytes))
				}
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			res := rt.New()
			if nav := res.navigator(); assert.False(t, nav.replace(c.data).hasError()) {
				var (
					jsonBytes []byte
					err       error
				)
				switch {
				case len(c.includes) == 0 && len(c.excludes) == 0:
					jsonBytes, err = json.Marshal(res)
				case len(c.includes) > 0:
					jsonBytes, err = res.MarshalJSONWithAttributes(c.includes...)
				case len(c.excludes) > 0:
					jsonBytes, err = res.MarshalJSONWithExcludedAttributes(c.excludes...)
				default:
					panic("invalid test setting")
				}

				c.assert(t, jsonBytes, err)
			}
		})
	}
}
