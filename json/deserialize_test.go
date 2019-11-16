package json

import (
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeserialize(t *testing.T) {
	var resourceType *core.ResourceType
	{
		_ = core.Schemas.MustLoad("../resource/schema/test_object_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/test_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/test_object_resource_type.json")
	}

	tests := []struct {
		name   string
		json   string
		assert func(t *testing.T, original, actual string, err error)
	}{
		{
			name: "default",
			json: `
{
	"schemas": ["urn:imulab:scim:TestObject"],
	"id": "5EE4A343-B8F8-4403-B87F-F44DB7133480",
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
			json: `{"schemas":["urn:imulab:scim:TestObject"],"id":"5EE4A343-B8F8-4403-B87F-F44DB7133480","name":"foobar","age":18,"score":95.5,"status":true,"tags":["foo","bar"],"courses":[{"name":"101","core":true},{"name":"102"}]}`,
			assert: func(t *testing.T, original, actual string, err error) {
				assert.Nil(t, err)
				assert.JSONEq(t, original, actual)
			},
		},
		{
			name: "empty",
			json: `{"schemas":["urn:imulab:scim:TestObject"],"id":"5EE4A343-B8F8-4403-B87F-F44DB7133480"}`,
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
