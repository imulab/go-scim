package crud

import (
	"encoding/json"
	"errors"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestAddReplaceDelete(t *testing.T) {
	s := new(CrudTestSuite)
	suite.Run(t, s)
}

type CrudTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *CrudTestSuite) TestAdd() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		path        string
		value       interface{}
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "add to non existing field yields error",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path:  "foobar",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidPath, errors.Unwrap(err))
			},
		},
		{
			name: "add to top level simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path:  "id",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", r.Navigator().Dot("id").Current().Raw())
			},
		},
		{
			name: "add to nested simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path:  "meta.version",
			value: "v1",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "v1", r.Navigator().Dot("meta").Dot("version").Current().Raw())
			},
		},
		{
			name: "add to simple multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path:  "schemas",
			value: "A",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"A"}, r.Navigator().Dot("schemas").Current().Raw())
			},
		},
		{
			name: "add to complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path: "emails",
			value: map[string]interface{}{
				"value":   "foo",
				"primary": true,
			},
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
		{
			name: "add to every simple property inside a complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				}).HasError())
				return r
			},
			path:  "emails.primary",
			value: false,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": false,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": false,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
		{
			name: "add to a selected few simple property inside a complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				}).HasError())
				return r
			},
			path:  `emails[value eq "bar"].primary`,
			value: true,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": true,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			err := Add(resource, test.path, test.value)
			test.expect(t, resource, err)
		})
	}
}

func (s *CrudTestSuite) TestReplace() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		path        string
		value       interface{}
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "replace non existing field yields error",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path:  "foobar",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidPath, errors.Unwrap(err))
			},
		},
		{
			name: "replace nested simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path:  "meta.version",
			value: "v1",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "v1", r.Navigator().Dot("meta").Dot("version").Current().Raw())
			},
		},
		{
			name: "replace simple multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path:  "schemas",
			value: []interface{}{"A"},
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"A"}, r.Navigator().Dot("schemas").Current().Raw())
			},
		},
		{
			name: "replace single multiValued property element",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				}).HasError())
				return r
			},
			path: `emails[value eq "bar"]`,
			value: map[string]interface{}{
				"value": "baz",
			},
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value":   "baz",
						"primary": nil,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
		{
			name: "replace multiValued property element field with filter",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				}).HasError())
				return r
			},
			path:  `emails[value eq "bar"].primary`,
			value: true,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": true,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			err := Replace(resource, test.path, test.value)
			test.expect(t, resource, err)
		})
	}
}

func (s *CrudTestSuite) TestDelete() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		path        string
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "delete non existing field yields error",
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			path: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidPath, errors.Unwrap(err))
			},
		},
		{
			name: "delete simple property",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("id").Replace("foobar").HasError())
				return r
			},
			path: "id",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Nil(t, r.Navigator().Dot("id").Current().Raw())
			},
		},
		{
			name: "delete nested simple property",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("meta").Dot("version").Replace("v1").HasError())
				return r
			},
			path: "meta.version",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Nil(t, r.Navigator().Dot("meta").Dot("version").Current().Raw())
			},
		},
		{
			name: "delete single multiValued property element",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				}).HasError())
				return r
			},
			path: `emails[value eq "bar"]`,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
		{
			name: "delete multiValued property element field with filter",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				}).HasError())
				return r
			},
			path: `emails[value eq "foo"].primary`,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": nil,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
		{
			name: "delete all property element field with filter",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": false,
					},
				}).HasError())
				return r
			},
			path: `emails.primary`,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": nil,
					},
				}, r.Navigator().Dot("emails").Current().Raw())
			},
		},
		{
			name: "delete empty path yields error",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": false,
					},
				}).HasError())
				return r
			},
			path: "",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidPath, errors.Unwrap(err))
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			err := Delete(resource, test.path)
			test.expect(t, resource, err)
		})
	}
}

func (s *CrudTestSuite) SetupSuite() {
	core := new(spec.Schema)
	require.Nil(s.T(), json.Unmarshal([]byte(testCoreSchema), core))
	spec.Schemas().Register(core)

	schema := new(spec.Schema)
	require.Nil(s.T(), json.Unmarshal([]byte(testMainSchema), schema))
	spec.Schemas().Register(schema)

	s.resourceType = new(spec.ResourceType)
	require.Nil(s.T(), json.Unmarshal([]byte(testResourceType), s.resourceType))
	Register(s.resourceType)
}

const (
	testCoreSchema = `
{
  "id": "core",
  "name": "core",
  "attributes": [
    {
      "id": "schemas",
      "name": "schemas",
      "type": "string",
      "multiValued": true,
      "_index": 0,
      "_path": "schemas",
      "_annotations": {
        "@AutoCompact": {}
      }
    },
    {
      "id": "id",
      "name": "id",
      "type": "string",
      "_index": 1,
      "_path": "id"
    },
    {
      "id": "meta",
      "name": "meta",
      "type": "complex",
      "_index": 2,
      "_path": "meta",
      "_annotations": {
        "@StateSummary": {}
      },
      "subAttributes": [
        {
          "id": "meta.version",
          "name": "version",
          "type": "string",
          "_index": 0,
          "_path": "meta.version"
        },
        {
          "id": "meta.location",
          "name": "location",
          "type": "reference",
          "_index": 1,
          "_path": "meta.location"
        }
      ]
    }
  ]
}
`
	testMainSchema = `
{
  "id": "main",
  "name": "main",
  "attributes": [
    {
      "id": "emails",
      "name": "emails",
      "type": "complex",
      "multiValued": true,
      "_index": 100,
      "_path": "emails",
      "_annotations": {
        "@ExclusivePrimary": {},
        "@AutoCompact": {},
        "@ElementAnnotations": {
          "@StateSummary": {}
        }
      },
      "subAttributes": [
        {
          "id": "emails.value",
          "name": "value",
          "type": "string",
          "_index": 0,
          "_path": "emails.value",
          "_annotations": {
            "@Identity": {}
          }
        },
        {
          "id": "emails.primary",
          "name": "primary",
          "type": "boolean",
          "_index": 1,
          "_path": "emails.primary",
          "_annotations": {
            "@Primary": {}
          }
        }
      ]
    }
  ]
}
`
	testResourceType = `
{
  "id": "Test",
  "name": "Test",
  "schema": "main"
}
`
)
