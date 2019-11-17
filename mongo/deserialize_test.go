package mongo

import (
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResourceUnmarshaler(t *testing.T) {
	var resourceType *core.ResourceType
	{
		_ = core.Schemas.MustLoad("../resource/schema/test_object_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/test_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/test_object_resource_type.json")
	}

	tests := []struct {
		name        string
		getResource func() *core.Resource
		assert      func(t *testing.T, result *core.Resource, err error)
	}{
		{
			name: "default",
			getResource: func() *core.Resource {
				r := core.Resources.New(resourceType)
				err := r.Replace(nil, map[string]interface{}{
					"name":        "TestUser123",
					"age":         int64(18),
					"score":       95.5,
					"status":      true,
					"certificate": "aGVsbG8gd29ybGQK",
					"secret":      "s3cret",
					"profile":     "https://test.org/results/TestUser123",
					"tags":        []interface{}{"foo", "bar"},
					"courses": []interface{}{
						map[string]interface{}{
							"name": "101",
							"core": true,
						},
						map[string]interface{}{
							"name": "102",
						},
					},
				})
				require.Nil(t, err)
				return r
			},
			assert: func(t *testing.T, result *core.Resource, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, result)

				name, err := result.Get(core.Steps.NewPath("name"))
				assert.Nil(t, err)
				assert.Equal(t, "TestUser123", name)

				age, err := result.Get(core.Steps.NewPath("age"))
				assert.Nil(t, err)
				assert.Equal(t, int64(18), age)

				score, err := result.Get(core.Steps.NewPath("score"))
				assert.Nil(t, err)
				assert.Equal(t, 95.5, score)

				status, err := result.Get(core.Steps.NewPath("status"))
				assert.Nil(t, err)
				assert.Equal(t, true, status)

				cert, err := result.Get(core.Steps.NewPath("certificate"))
				assert.Nil(t, err)
				assert.Equal(t, "aGVsbG8gd29ybGQK", cert)

				secret, err := result.Get(core.Steps.NewPath("secret"))
				assert.Nil(t, err)
				assert.Equal(t, "s3cret", secret)

				profile, err := result.Get(core.Steps.NewPath("profile"))
				assert.Nil(t, err)
				assert.Equal(t, "https://test.org/results/TestUser123", profile)

				tags, err := result.Get(core.Steps.NewPath("tags"))
				assert.Nil(t, err)
				assert.Len(t, tags, 2)
				assert.Contains(t, tags, "foo")
				assert.Contains(t, tags, "bar")

				courses, err := result.Get(core.Steps.NewPath("courses"))
				assert.Nil(t, err)
				assert.Len(t, courses, 2)
				assert.Equal(t, "101", courses.([]interface{})[0].(map[string]interface{})["name"])
				assert.Equal(t, true, courses.([]interface{})[0].(map[string]interface{})["core"])
				assert.Equal(t, "102", courses.([]interface{})[1].(map[string]interface{})["name"])
				assert.Nil(t, courses.([]interface{})[1].(map[string]interface{})["core"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := newBsonAdapter(test.getResource())
			b, err := d.MarshalBSON()
			assert.Nil(t, err)

			u := newResourceUnmarshaler(resourceType)
			err = u.UnmarshalBSON(b)
			test.assert(t, u.resource, err)
		})
	}
}
