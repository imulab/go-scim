package crud

import (
	"encoding/json"
	"errors"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestSeekSortTarget(t *testing.T) {
	s := new(SeekSortByTargetTestSuite)
	suite.Run(t, s)
}

type SeekSortByTargetTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *SeekSortByTargetTestSuite) TestSeekSortTarget() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		sortBy      string
		expect      func(t *testing.T, target prop.Property, err error)
	}{
		{
			name: "simple target",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("id").Replace("foobar").HasError())
				return r
			},
			sortBy: "id",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "id", target.Attribute().ID())
				assert.Equal(t, "foobar", target.Raw())
			},
		},
		{
			name: "nested target",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("meta").Dot("version").Replace("v1").HasError())
				return r
			},
			sortBy: "meta.version",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "meta.version", target.Attribute().ID())
				assert.Equal(t, "v1", target.Raw())
			},
		},
		{
			name: "multiValued simple target returns first",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("schemas").Add([]interface{}{"A", "B"}).HasError())
				return r
			},
			sortBy: "schemas",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "schemas$elem", target.Attribute().ID())
				assert.Equal(t, "A", target.Raw())
			},
		},
		{
			name: "multiValued complex target with no true-primary returns first",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value": "foo",
					},
					map[string]interface{}{
						"value": "bar",
					},
				}).HasError())
				return r
			},
			sortBy: "emails.value",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "emails.value", target.Attribute().ID())
				assert.Equal(t, "foo", target.Raw())
			},
		},
		{
			name: "multiValued complex target with true-primary returns true-primary value",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value": "foo",
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": true,
					},
				}).HasError())
				return r
			},
			sortBy: "emails.value",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "emails.value", target.Attribute().ID())
				assert.Equal(t, "bar", target.Raw())
			},
		},
		{
			name: "invalid target",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Add([]interface{}{
					map[string]interface{}{
						"value": "foo",
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": true,
					},
				}).HasError())
				return r
			},
			sortBy: "emails",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidPath, errors.Unwrap(err))
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			r := test.getResource(t)

			sortBy, err := expr.CompilePath(test.sortBy)
			assert.Nil(t, err)

			target, err := SeekSortTarget(r, sortBy)
			test.expect(t, target, err)
		})
	}
}

func (s *SeekSortByTargetTestSuite) SetupSuite() {
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
