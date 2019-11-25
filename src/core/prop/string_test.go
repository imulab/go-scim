package prop

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestStringProperty(t *testing.T) {
	suite.Run(t, new(StringPropertyTestSuite))
}

type StringPropertyTestSuite struct {
	suite.Suite
}

func (s *StringPropertyTestSuite) TestRaw() {
	tests := []struct {
		name   string
		prop   core.Property
		expect func(t *testing.T, value interface{})
	}{
		{
			name: "unassigned property returns nil",
			prop: NewString(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`)),
			expect: func(t *testing.T, value interface{}) {
				assert.Nil(t, value)
			},
		},
		{
			name: "assigned property returns non-nil",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			expect: func(t *testing.T, value interface{}) {
				assert.Equal(t, "imulab", value)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			test.expect(t, test.prop.Raw())
		})
	}
}

func (s *StringPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name    string
		getProp func() core.Property
		expect  func(t *testing.T, unassigned bool, dirty bool)
	}{
		{
			name: "unassigned property returns true for unassigned, and false for dirty",
			getProp: func() core.Property {
				return NewString(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`))
			},
			expect: func(t *testing.T, unassigned bool, dirty bool) {
				assert.True(t, unassigned)
				assert.False(t, dirty)
			},
		},
		{
			name: "explicitly unassigned property returns true for unassigned, and true for dirty",
			getProp: func() core.Property {
				prop := NewString(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`))
				s.Require().Nil(prop.Delete())
				return prop
			},
			expect: func(t *testing.T, unassigned bool, dirty bool) {
				assert.True(t, unassigned)
				assert.True(t, dirty)
			},
		},
		{
			name: "assigned property returns false for unassigned, and true for dirty",
			getProp: func() core.Property {
				prop := NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab")
				return prop
			},
			expect: func(t *testing.T, unassigned bool, dirty bool) {
				assert.False(t, unassigned)
				assert.True(t, dirty)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			unassigned, dirty := test.getProp().IsUnassigned()
			test.expect(t, unassigned, dirty)
		})
	}
}

func (s *StringPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		p1     core.Property
		p2     core.Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "same properties match",
			p1: NewStringOf(s.mustAttribute(`
{
	"id": "userName",
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			p2: NewStringOf(s.mustAttribute(`
{
	"id": "userName",
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "same properties match (non-caseExact)",
			p1: NewStringOf(s.mustAttribute(`
{
	"id": "userName",
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			p2: NewStringOf(s.mustAttribute(`
{
	"id": "userName",
	"name": "userName",
	"type": "string"
}
`), "IMULAB"),
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "same properties does not match",
			p1: NewStringOf(s.mustAttribute(`
{
	"id": "userName",
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			p2: NewStringOf(s.mustAttribute(`
{
	"id": "title",
	"name": "title",
	"type": "string"
}
`), "imulab"),
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

func (s *StringPropertyTestSuite) TestEqualsTo() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, equal bool, err error)
	}{
		{
			name: "equal value (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "imulab",
			expect: func(t *testing.T, equal bool, err error) {
				assert.Nil(t, err)
				assert.True(t, equal)
			},
		},
		{
			name: "equal value (non-caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": false
}
`), "imulab"),
			v: "IMULAB",
			expect: func(t *testing.T, equal bool, err error) {
				assert.Nil(t, err)
				assert.True(t, equal)
			},
		},
		{
			name: "equal value different case (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "IMULAB",
			expect: func(t *testing.T, equal bool, err error) {
				assert.Nil(t, err)
				assert.False(t, equal)
			},
		},
		{
			name: "different value",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			v: "foobar",
			expect: func(t *testing.T, equal bool, err error) {
				assert.Nil(t, err)
				assert.False(t, equal)
			},
		},
		{
			name: "incompatible value",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			v: 100,
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

func (s *StringPropertyTestSuite) TestStartsWith() {
	tests := []struct {
		name   string
		prop   core.Property
		v      string
		expect func(t *testing.T, startsWith bool, err error)
	}{
		{
			name: "prefix (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "i",
			expect: func(t *testing.T, startsWith bool, err error) {
				assert.Nil(t, err)
				assert.True(t, startsWith)
			},
		},
		{
			name: "prefix (non-caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": false
}
`), "imulab"),
			v: "I",
			expect: func(t *testing.T, startsWith bool, err error) {
				assert.Nil(t, err)
				assert.True(t, startsWith)
			},
		},
		{
			name: "prefix with different case (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "I",
			expect: func(t *testing.T, startsWith bool, err error) {
				assert.Nil(t, err)
				assert.False(t, startsWith)
			},
		},
		{
			name: "non-prefix",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			v: "a",
			expect: func(t *testing.T, startsWith bool, err error) {
				assert.Nil(t, err)
				assert.False(t, startsWith)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			equal, err := test.prop.StartsWith(test.v)
			test.expect(t, equal, err)
		})
	}
}

func (s *StringPropertyTestSuite) TestEndsWith() {
	tests := []struct {
		name   string
		prop   core.Property
		v      string
		expect func(t *testing.T, contains bool, err error)
	}{
		{
			name: "suffix (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "b",
			expect: func(t *testing.T, endsWith bool, err error) {
				assert.Nil(t, err)
				assert.True(t, endsWith)
			},
		},
		{
			name: "suffix (non-caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": false
}
`), "imulab"),
			v: "B",
			expect: func(t *testing.T, endsWith bool, err error) {
				assert.Nil(t, err)
				assert.True(t, endsWith)
			},
		},
		{
			name: "suffix with different case (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "B",
			expect: func(t *testing.T, endsWith bool, err error) {
				assert.Nil(t, err)
				assert.False(t, endsWith)
			},
		},
		{
			name: "non-suffix",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			v: "a",
			expect: func(t *testing.T, endsWith bool, err error) {
				assert.Nil(t, err)
				assert.False(t, endsWith)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			equal, err := test.prop.EndsWith(test.v)
			test.expect(t, equal, err)
		})
	}
}

func (s *StringPropertyTestSuite) TestContains() {
	tests := []struct {
		name   string
		prop   core.Property
		v      string
		expect func(t *testing.T, contains bool, err error)
	}{
		{
			name: "sub string (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "mula",
			expect: func(t *testing.T, contains bool, err error) {
				assert.Nil(t, err)
				assert.True(t, contains)
			},
		},
		{
			name: "substring (non-caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": false
}
`), "imulab"),
			v: "MULA",
			expect: func(t *testing.T, contains bool, err error) {
				assert.Nil(t, err)
				assert.True(t, contains)
			},
		},
		{
			name: "substring with different case (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "imulab"),
			v: "MULA",
			expect: func(t *testing.T, contains bool, err error) {
				assert.Nil(t, err)
				assert.False(t, contains)
			},
		},
		{
			name: "non-substring",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab"),
			v: "foobar",
			expect: func(t *testing.T, contains bool, err error) {
				assert.Nil(t, err)
				assert.False(t, contains)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			equal, err := test.prop.Contains(test.v)
			test.expect(t, equal, err)
		})
	}
}

func (s *StringPropertyTestSuite) TestGreaterThan() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, greaterThan bool, err error)
	}{
		{
			name: "greater value (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "b"),
			v: "a",
			expect: func(t *testing.T, greaterThan bool, err error) {
				assert.Nil(t, err)
				assert.True(t, greaterThan)
			},
		},
		{
			name: "greater value (non-caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": false
}
`), "b"),
			v: "A",
			expect: func(t *testing.T, greaterThan bool, err error) {
				assert.Nil(t, err)
				assert.True(t, greaterThan)
			},
		},
		{
			name: "greater value different case (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "A"),
			v: "a",
			expect: func(t *testing.T, greaterThan bool, err error) {
				assert.Nil(t, err)
				assert.False(t, greaterThan)
			},
		},
		{
			name: "incompatible value",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "a"),
			v: 100,
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

func (s *StringPropertyTestSuite) TestLessThan() {
	tests := []struct {
		name   string
		prop   core.Property
		v      interface{}
		expect func(t *testing.T, lessThan bool, err error)
	}{
		{
			name: "less value (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "a"),
			v: "b",
			expect: func(t *testing.T, lessThan bool, err error) {
				assert.Nil(t, err)
				assert.True(t, lessThan)
			},
		},
		{
			name: "less value (non-caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": false
}
`), "A"),
			v: "b",
			expect: func(t *testing.T, lessThan bool, err error) {
				assert.Nil(t, err)
				assert.True(t, lessThan)
			},
		},
		{
			name: "less value different case (caseExact)",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string",
	"caseExact": true
}
`), "a"),
			v: "A",
			expect: func(t *testing.T, lessThan bool, err error) {
				assert.Nil(t, err)
				assert.False(t, lessThan)
			},
		},
		{
			name: "incompatible value",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "a"),
			v: 100,
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

func (s *StringPropertyTestSuite) TestPresent() {
	tests := []struct{
		name		string
		prop		core.Property
		expect		func(t *testing.T, present bool)
	}{
		{
			name: 	"unassigned value is not present",
			prop: 	NewString(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`)),
			expect: func(t *testing.T, present bool) {
				assert.False(t, present)
			},
		},
		{
			name: 	"unassigned value is not present",
			prop: 	NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "imulab"),
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

func (s *StringPropertyTestSuite) TestAdd() {
	tests := []struct{
		name		string
		prop		core.Property
		v			interface{}
		expect 		func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: 	"add to unassigned",
			prop: NewString(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`)),
			v: "imulab",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "imulab", raw)
			},
		},
		{
			name: 	"add to assigned replaces the value",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "foobar"),
			v: "imulab",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "imulab", raw)
			},
		},
		{
			name: 	"add incompatible returns error",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "foobar"),
			v: 100,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			prop := test.prop
			err := prop.Add(test.v)
			test.expect(t, prop.Raw(), err)
		})
	}
}

func (s *StringPropertyTestSuite) TestReplace() {
	tests := []struct{
		name		string
		prop		core.Property
		v			interface{}
		expect 		func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: 	"replace unassigned",
			prop: NewString(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`)),
			v: "imulab",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "imulab", raw)
			},
		},
		{
			name: 	"replace assigned",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "foobar"),
			v: "imulab",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "imulab", raw)
			},
		},
		{
			name: 	"replace incompatible returns error",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "foobar"),
			v: 100,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			prop := test.prop
			err := prop.Replace(test.v)
			test.expect(t, prop.Raw(), err)
		})
	}
}

func (s *StringPropertyTestSuite) TestDelete() {
	tests := []struct{
		name		string
		prop		core.Property
		expect 		func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: 	"delete unassigned",
			prop: NewString(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`)),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Nil(t, raw)
			},
		},
		{
			name: 	"delete assigned",
			prop: NewStringOf(s.mustAttribute(`
{
	"name": "userName",
	"type": "string"
}
`), "foobar"),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Nil(t, raw)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			prop := test.prop
			err := prop.Delete()
			test.expect(t, prop.Raw(), err)
		})
	}
}

func (s *StringPropertyTestSuite) mustAttribute(jsonValue string) *core.Attribute {
	attr := new(core.Attribute)
	err := json.Unmarshal([]byte(jsonValue), attr)
	s.Require().Nil(err)
	return attr
}
