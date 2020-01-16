package internal

import (
	"encoding/json"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExclusivePrimarySubscriber(t *testing.T) {
	attrFunc := func(t *testing.T) *spec.Attribute {
		attr := new(spec.Attribute)
		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "emails",
  "name": "emails",
  "type": "complex",
  "multiValued": true,
  "subAttributes": [
    {
      "id": "emails.value",
      "name": "value",
      "type": "string",
      "_path": "emails.value",
      "_index": 0,
      "_annotations": {
        "@Identity": {}
      }
    },
    {
      "id": "emails.primary",
      "name": "primary",
      "type": "boolean",
      "_path": "emails.primary",
      "_index": 1,
      "_annotations": {
        "@Primary": {}
      }
    }
  ],
  "_path": "emails",
  "_index": 0,
  "_annotations": {
    "@ExclusivePrimary": {}
  }
}
`), attr))
		return attr
	}

	tests := []struct {
		name        string
		getProperty func(t *testing.T) prop.Property
		modFunc     func(t *testing.T, p prop.Property)
		expect      func(t *testing.T, raw interface{})
	}{
		{
			name: "assigning new primary turns off old primary",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).At(1).Dot("primary").Replace(true))
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": true,
					},
				}, raw)
			},
		},
		{
			name: "assigning old primary has no side effect",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).At(0).Dot("primary").Replace(true))
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": nil,
					},
				}, raw)
			},
		},
		{
			name: "turning off old primary has no side effect",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).At(0).Dot("primary").Delete())
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": nil,
					},
				}, raw)
			},
		},
		{
			name: "assigning other primary to false has no side effect",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).At(1).Dot("primary").Replace(false))
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": false,
					},
				}, raw)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := test.getProperty(t)
			test.modFunc(t, p)
			test.expect(t, p.Raw())
		})
	}
}
