package crud

import (
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
)

func TestEvaluate(t *testing.T) {
	s := new(EvaluateTestSuite)
	suite.Run(t, s)
}

type EvaluateTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *EvaluateTestSuite) TestEvaluate() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		filter      string
		expect      func(t *testing.T, result bool, err error)
	}{
		{
			name: `[id eq "foobar"] evaluates to true against {"id":"foobar"}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("id").Replace("foobar").HasError())
				return r
			},
			filter: fmt.Sprintf("id eq %s", strconv.Quote("foobar")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[id ne "foobar"] evaluates to false against {"id":"foobar"}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("id").Replace("foobar").HasError())
				return r
			},
			filter: fmt.Sprintf("id ne %s", strconv.Quote("foobar")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.False(t, result)
			},
		},
		{
			name: `[id sw "foo"] evaluates to true against {"id":"foobar"}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("id").Replace("foobar").HasError())
				return r
			},
			filter: fmt.Sprintf("id sw %s", strconv.Quote("foo")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[id ew "bar"] evaluates to true against {"id":"foobar"}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("id").Replace("foobar").HasError())
				return r
			},
			filter: fmt.Sprintf("id ew %s", strconv.Quote("bar")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[id co "xxx"] evaluates to false against {"id":"foobar"}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("id").Replace("foobar").HasError())
				return r
			},
			filter: fmt.Sprintf("id co %s", strconv.Quote("xxx")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.False(t, result)
			},
		},
		{
			name: `[meta.version gt "v1"] evaluates to true against {"meta":{"version":"v2"}}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("meta").Dot("version").Replace("v2").HasError())
				return r
			},
			filter: fmt.Sprintf("meta.version gt %s", strconv.Quote("v1")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[meta.version ge "v1"] evaluates to true against {"meta":{"version":"v1"}}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("meta").Dot("version").Replace("v1").HasError())
				return r
			},
			filter: fmt.Sprintf("meta.version ge %s", strconv.Quote("v1")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[meta.version lt "v1"] evaluates to false against {"meta":{"version":"v2"}}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("meta").Dot("version").Replace("v2").HasError())
				return r
			},
			filter: fmt.Sprintf("meta.version lt %s", strconv.Quote("v1")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.False(t, result)
			},
		},
		{
			name: `[meta.version le "v1"] evaluates to true against {"meta":{"version":"v0"}}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("meta").Dot("version").Replace("v0").HasError())
				return r
			},
			filter: fmt.Sprintf("meta.version le %s", strconv.Quote("v1")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[emails pr] evaluates to false against {}`,
			getResource: func(t *testing.T) *prop.Resource {
				return prop.NewResource(s.resourceType)
			},
			filter: "emails pr",
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.False(t, result)
			},
		},
		{
			name: `[schemas pr] evaluates to true against {"schemas": ["A", "B"]}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("schemas").Replace([]interface{}{"A", "B"}).HasError())
				return r
			},
			filter: "schemas pr",
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[schemas eq "B"] evaluates to true against {"schemas": ["A", "B"]}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("schemas").Replace([]interface{}{"A", "B"}).HasError())
				return r
			},
			filter: fmt.Sprintf("schemas eq %s", strconv.Quote("B")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[emails.value eq "bar"] evaluates to true against {"emails": [{"value": "foo"}, {"value": "bar"}]}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Replace([]interface{}{
					map[string]interface{}{"value": "foo"},
					map[string]interface{}{"value": "bar"},
				}).HasError())
				return r
			},
			filter: fmt.Sprintf("emails.value eq %s", strconv.Quote("bar")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.True(t, result)
			},
		},
		{
			name: `[emails.value eq "bar"] evaluates to false against {"emails": [{"value": "foo"}]}`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Dot("emails").Replace([]interface{}{
					map[string]interface{}{"value": "foo"},
				}).HasError())
				return r
			},
			filter: fmt.Sprintf("emails.value eq %s", strconv.Quote("bar")),
			expect: func(t *testing.T, result bool, err error) {
				assert.Nil(t, err)
				assert.False(t, result)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			result, err := Evaluate(resource, test.filter)
			test.expect(t, result, err)
		})
	}
}

// Prepares a core schema with 'schemas', 'id', 'meta'('version', 'location') attributes, and a main schema
// with 'emails'('value', 'primary') attributes. Aggregate the two schemas in the test resource type.
func (s *EvaluateTestSuite) SetupSuite() {
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
