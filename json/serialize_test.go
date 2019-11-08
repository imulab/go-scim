package json

import (
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	tests := []struct{
		name		string
		includes	[]string
		excludes	[]string
		getResource	func() *core.Resource
		assert		func(t *testing.T, json []byte, err error)
	}{
		{
			name: 	"default",
			getResource: func() *core.Resource {
				r := core.Resources.New(resourceType)
				err := r.Replace(nil, map[string]interface{}{
					"id": "f61820aa-1b37-41bb-a3d3-9ca2de83cb45",
					"name": "TestUser123",
					"age": int64(18),
					"score": 95.5,
					"status": true,
					"certificate": "aGVsbG8gd29ybGQK",
					"secret": "s3cret",
					"profile": "https://test.org/results/TestUser123",
					"tags": []interface{}{"foo", "bar"},
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
			assert: func(t *testing.T, json []byte, err error) {
				assert.Nil(t, err)
				expects := `
{
   "age":18,
   "score":95.5,
   "status":true,
   "profile":"https://test.org/results/TestUser123",
   "tags":[
      "foo",
      "bar"
   ],
   "courses":[
      {
         "core":true,
         "name":"101"
      },
      {
         "name":"102"
      }
   ],
   "name":"TestUser123"
}
`
				assert.JSONEq(t, expects, string(json))
			},
		},
		{
			name: "includes/excludes",
			getResource: func() *core.Resource {
				r := core.Resources.New(resourceType)
				err := r.Replace(nil, map[string]interface{}{
					"id": "f61820aa-1b37-41bb-a3d3-9ca2de83cb45",
					"name": "TestUser123",
					"age": int64(18),
					"score": 95.5,
					"status": true,
					"certificate": "aGVsbG8gd29ybGQK",
					"secret": "s3cret",
					"profile": "https://test.org/results/TestUser123",
					"tags": []interface{}{"foo", "bar"},
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
			includes: []string{"certificate"},
			excludes: []string{"age"},
			assert: func(t *testing.T, json []byte, err error) {
				assert.Nil(t, err)
				expects := `
{
   "score":95.5,
   "status":true,
   "profile":"https://test.org/results/TestUser123",
   "name":"TestUser123",
   "certificate":"aGVsbG8gd29ybGQK",
   "tags":[
      "foo",
      "bar"
   ],
   "courses":[
      {
         "name":"101",
         "core":true
      },
      {
         "name":"102"
      }
   ]
}
`
				assert.JSONEq(t, expects, string(json))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			json, err := Serialize(test.getResource(), test.includes, test.excludes)
			test.assert(t, json, err)
		})
	}
}
