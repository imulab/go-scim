package prop

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestIntegerProperty(t *testing.T) {
	suite.Run(t, new(IntegerPropertyTestSuite))
}

type IntegerPropertyTestSuite struct {
	suite.Suite
}

func (s *IntegerPropertyTestSuite) TestRaw() {
	tests := []struct {
		name   string
		prop   core.Property
		expect func(t *testing.T, raw interface{})
	}{
		{
			name: "unassigned property returns nil",
			prop: NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`)),
			expect: func(t *testing.T, raw interface{}) {
				assert.Nil(t, raw)
			},
		},
		{
			name: "assigned property returns int64",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 100),
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, int64(100), raw)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			test.expect(t, test.prop.Raw())
		})
	}
}

func (s *IntegerPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name    string
		getProp func() core.Property
		expect  func(t *testing.T, unassigned bool)
	}{
		{
			name: "unassigned property returns true for unassigned, and false for dirty",
			getProp: func() core.Property {
				return NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`))
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "explicitly unassigned property returns true for unassigned, and true for dirty",
			getProp: func() core.Property {
				prop := NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`))
				_, err := prop.Delete()
				s.Require().Nil(err)
				return prop
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "assigned property returns false for unassigned, and true for dirty",
			getProp: func() core.Property {
				prop := NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 18)
				return prop
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.False(t, unassigned)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			test.expect(t, test.getProp().IsUnassigned())
		})
	}
}

func (s *IntegerPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		p1     core.Property
		p2     core.Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "same properties match",
			p1: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			p2: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "different properties does not match",
			p1: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 65),
			p2: NewIntegerOf(s.mustAttribute(`
{
	"id": "score",
	"name": "score",
	"type": "integer"
}
`), 65),
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			test.expect(t, test.p1.Matches(test.p2))
		})
	}
}

func (s *IntegerPropertyTestSuite) TestEqualsTo() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, equal bool, err error)
	}{
		{
			name: "equal value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			v: 18,
			expect: func(t *testing.T, equal bool, err error) {
				assert.Nil(t, err)
				assert.True(t, equal)
			},
		},
		{
			name: "different value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			v: 19,
			expect: func(t *testing.T, equal bool, err error) {
				assert.Nil(t, err)
				assert.False(t, equal)
			},
		},
		{
			name: "incompatible value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			v: "foobar",
			expect: func(t *testing.T, equal bool, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			equal, err := test.prop.EqualsTo(test.v)
			test.expect(t, equal, err)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestGreaterThan() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, greaterThan bool, err error)
	}{
		{
			name: "greater value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			v: 17,
			expect: func(t *testing.T, greaterThan bool, err error) {
				assert.Nil(t, err)
				assert.True(t, greaterThan)
			},
		},
		{
			name: "incompatible value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			v: "foobar",
			expect: func(t *testing.T, greaterThan bool, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			equal, err := test.prop.GreaterThan(test.v)
			test.expect(t, equal, err)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestLessThan() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, lessThan bool, err error)
	}{
		{
			name: "smaller value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			v: 19,
			expect: func(t *testing.T, lessThan bool, err error) {
				assert.Nil(t, err)
				assert.True(t, lessThan)
			},
		},
		{
			name: "incompatible value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			v: "foobar",
			expect: func(t *testing.T, lessThan bool, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			equal, err := test.prop.LessThan(test.v)
			test.expect(t, equal, err)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   core.Property
		expect func(t *testing.T, present bool)
	}{
		{
			name: "unassigned value is not present",
			prop: NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`)),
			expect: func(t *testing.T, present bool) {
				assert.False(t, present)
			},
		},
		{
			name: "assigned value is present",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 18),
			expect: func(t *testing.T, present bool) {
				assert.True(t, present)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			test.expect(t, test.prop.Present())
		})
	}
}

func (s *IntegerPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: "add to unassigned",
			prop: NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`)),
			v: 18,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(18), raw)
			},
		},
		{
			name: "add to assigned replaces the value",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 18),
			v: 20,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(20), raw)
			},
		},
		{
			name: "add incompatible returns error",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 18),
			v: "100",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			prop := test.prop
			_, err := prop.Add(test.v)
			test.expect(t, prop.Raw(), err)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: "replace unassigned",
			prop: NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`)),
			v: 18,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(18), raw)
			},
		},
		{
			name: "replace assigned",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 18),
			v: 20,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(20), raw)
			},
		},
		{
			name: "replace with incompatible returns error",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 18),
			v: "100",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			prop := test.prop
			_, err := prop.Replace(test.v)
			test.expect(t, prop.Raw(), err)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   core.Property
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewInteger(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`)),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Nil(t, raw)
			},
		},
		{
			name: "delete assigned",
			prop: NewIntegerOf(s.mustAttribute(`
{
	"name": "age",
	"type": "integer"
}
`), 18),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Nil(t, raw)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			prop := test.prop
			_, err := prop.Delete()
			test.expect(t, prop.Raw(), err)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestHash() {
	tests := []struct {
		name   string
		p1     core.Property
		p2     core.Property
		expect func(t *testing.T, h1 uint64, h2 uint64)
	}{
		{
			name: "same integer property have same hash",
			p1: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			p2: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			expect: func(t *testing.T, h1 uint64, h2 uint64) {
				assert.True(t, h1 == h2)
			},
		},
		{
			name: "different property value have different hash",
			p1: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18),
			p2: NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 20),
			expect: func(t *testing.T, h1 uint64, h2 uint64) {
				assert.False(t, h1 == h2)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			h1 := test.p1.Hash()
			h2 := test.p2.Hash()
			test.expect(t, h1, h2)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestModCount() {
	tests := []struct {
		name    string
		getProp func(t *testing.T) core.Property
		expect  func(t *testing.T, count int)
	}{
		{
			name: "modifying the property increases the mod count by 1",
			getProp: func(t *testing.T) core.Property {
				prop := NewInteger(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`))
				_, err := prop.Add(18)
				require.Nil(t, err)
				return prop
			},
			expect: func(t *testing.T, count int) {
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "deleting the property immediately after creation increase mod count by 1",
			getProp: func(t *testing.T) core.Property {
				prop := NewInteger(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`))
				_, err := prop.Delete()
				require.Nil(t, err)
				return prop
			},
			expect: func(t *testing.T, count int) {
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "modifying the property with same value does not increase mod count",
			getProp: func(t *testing.T) core.Property {
				prop := NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18)
				require.Equal(t, 1, prop.ModCount())

				_, err := prop.Replace(18)
				require.Nil(t, err)
				return prop
			},
			expect: func(t *testing.T, count int) {
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "deleting an already deleted property after first attempt does not increment mod count",
			getProp: func(t *testing.T) core.Property {
				prop := NewIntegerOf(s.mustAttribute(`
{
	"id": "age",
	"name": "age",
	"type": "integer"
}
`), 18)
				_, err := prop.Delete()
				require.Nil(t, err)
				require.True(t, prop.IsUnassigned())
				require.Equal(t, 2, prop.ModCount())

				// delete already deleted
				_, err = prop.Delete()
				require.Nil(t, err)
				return prop
			},
			expect: func(t *testing.T, count int) {
				assert.Equal(t, 2, count)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			test.expect(t, test.getProp(t).ModCount())
		})
	}
}

func (s *IntegerPropertyTestSuite) mustAttribute(jsonValue string) *core.Attribute {
	attr := new(core.Attribute)
	err := json.Unmarshal([]byte(jsonValue), attr)
	s.Require().Nil(err)
	return attr
}
