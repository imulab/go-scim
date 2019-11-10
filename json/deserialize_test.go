package json

import (
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestDeserialize(t *testing.T) {
	// prepare: schema
	schemaRaw, err := ioutil.ReadFile("../resource/schema/test_object_schema.json")
	require.Nil(t, err)
	schema, err := core.ParseSchema(schemaRaw)
	require.Nil(t, err)
	core.Schemas.Add(schema)

	// prepare: schema companion
	schemaCompanionRaw, err := ioutil.ReadFile("../resource/companion/test_object_schema_companion.json")
	require.Nil(t, err)
	schemaCompanion, err := core.ParseSchemaCompanion(schemaCompanionRaw)
	require.Nil(t, err)
	schemaCompanion.MustLoadOntoSchema()

	// prepare: resourceType
	resourceTypeRaw, err := ioutil.ReadFile("../resource/resource_type/test_object_resource_type.json")
	require.Nil(t, err)
	resourceType, err := core.ParseResourceType(resourceTypeRaw)
	require.Nil(t, err)

	tests := []struct{
		name	string
		json	string
		assert	func(t *testing.T, original, actual string, err error)
	}{
		{
			name: "default",
			json: `
{
	"name" : "foobar",
	"age": 18,
	"score": 95.5,
	"status": true,
	"tags": ["foo","bar"],
	"courses": [
		{
			"name": "101",
			"core": true
		},
		{
			"name": "102"
		}
	]
}
`,
			assert: func(t *testing.T, original, actual string, err error) {
				assert.Nil(t, err)
				assert.JSONEq(t, original, actual)
			},
		},
		{
			name: "compact",
			json: `{"name":"foobar","age":18,"score":95.5,"status":true,"tags":["foo","bar"],"courses":[{"name":"101","core":true},{"name":"102"}]}`,
			assert: func(t *testing.T, original, actual string, err error) {
				assert.Nil(t, err)
				assert.JSONEq(t, original, actual)
			},
		},
		{
			name: "empty",
			json: `{}`,
			assert: func(t *testing.T, original, actual string, err error) {
				assert.Nil(t, err)
				assert.JSONEq(t, original, actual)
			},
		},
		{
			name: "invalid",
			json: `{"name":"foobar",}`,
			assert: func(t *testing.T, original, actual string, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := core.Resources.New(resourceType)

			err := Deserialize([]byte(test.json), resource)
			if err != nil {
				test.assert(t, test.json, "", err)
				return
			}

			r, err := Serialize(resource, nil, nil)
			test.assert(t, test.json, string(r), err)
		})
	}
}
