package filter

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestMetaFilter(t *testing.T) {
	s := new(MetaFilterTestSuite)
	suite.Run(t, s)
}

type MetaFilterTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *MetaFilterTestSuite) TestMetaFilter() {
	tests := []struct {
		name         string
		getResource  func(t *testing.T) *prop.Resource
		getReference func(t *testing.T) *prop.Resource
		expect       func(t *testing.T, resource *prop.Resource, err error)
	}{
		{
			name: "generate meta for resource",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"id":       "C37527A1-B60F-4E30-8FD9-162A1740BDB6",
					"userName": "foobar",
				}).HasError())
				return r
			},
			getReference: func(t *testing.T) *prop.Resource {
				return nil
			},
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := resource.Navigator().Dot("meta")
				assert.False(t, nav.HasError())

				assert.Equal(t, "User", nav.Dot("resourceType").Current().Raw())
				nav.Retract()
				assert.NotNil(t, nav.Dot("created").Current().Raw())
				nav.Retract()
				assert.NotNil(t, nav.Dot("lastModified").Current().Raw())
				nav.Retract()
				assert.Equal(t, "/Users/C37527A1-B60F-4E30-8FD9-162A1740BDB6", nav.Dot("location").Current().Raw())
				nav.Retract()
				assert.NotEmpty(t, nav.Dot("version").Current().Raw())
				nav.Retract()
			},
		},
		{
			name: "update meta when resource has changed",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"id": "c37527a1-b60f-4e30-8fd9-162a1740bdb6",
					"meta": map[string]interface{}{
						"resourceType": "User",
						"created":      "2020-01-19T15:15:00",
						"lastModified": "2020-01-19T15:15:00",
						"version":      "W\"1\"",
					},
					"userName": "changed!!!",
				}).HasError())
				return r
			},
			getReference: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"id": "c37527a1-b60f-4e30-8fd9-162a1740bdb6",
					"meta": map[string]interface{}{
						"resourceType": "User",
						"created":      "2020-01-19T15:15:00",
						"lastModified": "2020-01-19T15:15:00",
						"version":      "W\"1\"",
					},
					"userName": "foobar",
				}).HasError())
				return r
			},
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := resource.Navigator().Dot("meta")
				assert.False(t, nav.HasError())

				assert.NotEqual(t, "2020-01-19T15:15:00", nav.Dot("lastModified").Current().Raw())
				nav.Retract()
				assert.NotEqual(t, "W\"1\"", nav.Dot("version").Current().Raw())
				nav.Retract()
			},
		},
		{
			name: "meta not updated when resource has not changed",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"id": "c37527a1-b60f-4e30-8fd9-162a1740bdb6",
					"meta": map[string]interface{}{
						"resourceType": "User",
						"created":      "2020-01-19T15:15:00",
						"lastModified": "2020-01-19T15:15:00",
						"version":      "W\"1\"",
					},
					"userName": "foobar",
				}).HasError())
				return r
			},
			getReference: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"id": "c37527a1-b60f-4e30-8fd9-162a1740bdb6",
					"meta": map[string]interface{}{
						"resourceType": "User",
						"created":      "2020-01-19T15:15:00",
						"lastModified": "2020-01-19T15:15:00",
						"version":      "W\"1\"",
					},
					"userName": "foobar",
				}).HasError())
				return r
			},
			expect: func(t *testing.T, resource *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := resource.Navigator().Dot("meta")
				assert.False(t, nav.HasError())

				assert.Equal(t, "2020-01-19T15:15:00", nav.Dot("lastModified").Current().Raw())
				nav.Retract()
				assert.Equal(t, "W\"1\"", nav.Dot("version").Current().Raw())
				nav.Retract()
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			filter := MetaFilter()

			resource := test.getResource(t)
			reference := test.getReference(t)

			var err error
			if reference == nil {
				err = filter.Filter(context.Background(), resource)
			} else {
				err = filter.FilterRef(context.Background(), resource, reference)
			}

			test.expect(t, resource, err)
		})
	}
}

func (s *MetaFilterTestSuite) SetupSuite() {
	for _, each := range []struct {
		filepath  string
		structure interface{}
		post      func(parsed interface{})
	}{
		{
			filepath:  "../../../../public/schemas/core_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../../public/schemas/user_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../../public/resource_types/user_resource_type.json",
			structure: new(spec.ResourceType),
			post: func(parsed interface{}) {
				s.resourceType = parsed.(*spec.ResourceType)
			},
		},
	} {
		f, err := os.Open(each.filepath)
		require.Nil(s.T(), err)

		raw, err := ioutil.ReadAll(f)
		require.Nil(s.T(), err)

		err = json.Unmarshal(raw, each.structure)
		require.Nil(s.T(), err)

		if each.post != nil {
			each.post(each.structure)
		}
	}
}
