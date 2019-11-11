package mongo

import (
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"testing"
)

func TestSerialize(t *testing.T) {
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

	tests := []struct {
		name        string
		getResource func() *core.Resource
		assert      func(t *testing.T, raw bson.Raw)
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
			assert: func(t *testing.T, raw bson.Raw) {
				for _, key := range []string{
					"name",
					"age",
					"score",
					"status",
					"certificate",
					"secret",
					"profile",
					"tags",
					"courses",
				} {
					_, err := raw.LookupErr(key)
					assert.Nil(t, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := test.getResource()
			b, err := Serialize(r)
			assert.Nil(t, err)

			test.assert(t, bson.Raw(b))
		})
	}
}
