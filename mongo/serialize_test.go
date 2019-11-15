package mongo

import (
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestSerialize(t *testing.T) {
	parser := test.NewResourceParser(t,
		"../resource/schema/test_object_schema.json",
		"../resource/companion/test_object_schema_companion.json",
		"../resource/resource_type/test_object_resource_type.json",
	)
	resourceType := parser.GetResourceType()

	tests := []struct {
		name        string
		getResource func() *core.Resource
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
		},
		{
			name: "empty multiValued",
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
				})
				require.Nil(t, err)
				return r
			},
		},
		{
			name: "test object 1",
			getResource: func() *core.Resource {
				return parser.MustLoadResource(t, "../resource/test/test_object_1.json")
			},
		},
		{
			name: "test object 2",
			getResource: func() *core.Resource {
				return parser.MustLoadResource(t, "../resource/test/test_object_2.json")
			},
		},
		{
			name: "test user 1",
			getResource: func() *core.Resource {
				return test.NewResourceParser(t,
					"../resource/schema/user_schema.json",
					"../resource/companion/user_schema_companion.json",
					"../resource/resource_type/user_resource_type.json",
				).MustLoadResource(t, "../resource/test/test_user_1.json")
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			d := newBsonAdapter(each.getResource())
			b, err := d.MarshalBSON()
			assert.Nil(t, err)
			assert.Nil(t, bson.Raw(b).Validate())
		})
	}
}
