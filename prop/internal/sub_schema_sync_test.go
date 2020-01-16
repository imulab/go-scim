package internal

import (
	"encoding/json"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSchemaSyncSubscriber(t *testing.T) {
	var (
		coreSchema      = new(spec.Schema)
		mainSchema      = new(spec.Schema)
		extensionSchema = new(spec.Schema)
		resourceType    = new(spec.ResourceType)
	)
	{
		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "core",
  "name": "core",
  "attributes": [
    {
      "id": "schemas",
      "name": "schemas",
      "type": "string",
      "multiValued": true,
      "_path": "schemas",
      "_annotations": {
        "@AutoCompact": {}
      }
    }
  ]
}
`), coreSchema))
		spec.Schemas().Register(coreSchema)

		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "main",
  "name": "main",
  "attributes": [
    {
      "id": "name",
      "name": "name",
      "type": "string",
      "_path": "name"
    }
  ]
}
`), mainSchema))
		spec.Schemas().Register(mainSchema)

		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "extension",
  "name": "extension",
  "attributes": [
    {
      "id": "text",
      "name": "text",
      "type": "string",
      "_path": "extension.text"
    }
  ]
}
`), extensionSchema))
		spec.Schemas().Register(extensionSchema)

		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "Test",
  "name": "Test",
  "schema": "main",
  "schemaExtensions": [
    {
      "schema": "extension",
      "required": false
    }
  ]
}
`), resourceType))
	}

	tests := []struct {
		name        string
		getProperty func(t *testing.T) prop.Property
		modFunc     func(t *testing.T, p prop.Property)
		expect      func(t *testing.T, schemas interface{})
	}{
		{
			name: "assign to unassigned extension",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas": []interface{}{"main"},
					"name":    "foobar",
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).Dot("extension").Dot("text").Replace("hello world"))
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main", "extension"}, schemas)
			},
		},
		{
			name: "assign to assigned extension does not add extension",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas": []interface{}{"main", "extension"},
					"name":    "foobar",
					"extension": map[string]interface{}{
						"text": "hello world",
					},
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).Dot("extension").Replace(map[string]interface{}{
					"text": "see ya later",
				}))
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main", "extension"}, schemas)
			},
		},
		{
			name: "unassigned extension has schema removed",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas": []interface{}{"main", "extension"},
					"name":    "foobar",
					"extension": map[string]interface{}{
						"text": "hello world",
					},
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).Dot("extension").Delete())
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main"}, schemas)
			},
		},
		{
			name: "deleting unassigned extension does not change schema",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas":   []interface{}{"main"},
					"name":      "foobar",
					"extension": map[string]interface{}{},
				})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).Dot("extension").Delete())
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main"}, schemas)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := test.getProperty(t)
			test.modFunc(t, p)
			test.expect(t, prop.Navigate(p).Dot("schemas").Current().Raw())
		})
	}
}
