package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestIntegerProperty(t *testing.T) {
	s := new(IntegerPropertyTestSuite)

	s.NewFunc = NewInteger
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewIntegerOf(attr, v.(int64))
	}

	suite.Run(t, s)
}

type IntegerPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *IntegerPropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:extension:test:2.0:User:age",
  "name": "age",
  "type": "integer",
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "age",
  "_index": 100
}`))
}

func (s *IntegerPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new integer of integer attribute",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:extension:test:2.0:User:age", p.Attribute().ID())
				assert.Equal(t, spec.TypeInteger, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new integer auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. IntegerPropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:extension:test:2.0:User:age",
  "name": "age",
  "type": "integer",
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "age",
  "_index": 100,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:extension:test:2.0:User:age", p.Attribute().ID())
				assert.Equal(t, spec.TypeInteger, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*integerProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestRaw() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *int64
		expect   func(t *testing.T, raw interface{})
	}{
		{
			name: "unassigned returns nil",
			attr: s.standardAttr,
			getValue: func() *int64 {
				return nil
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Nil(t, raw)
			},
		},
		{
			name: "assigned returns string",
			attr: s.standardAttr,
			getValue: func() *int64 {
				i := int64(64)
				return &i
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, int64(64), raw)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			v := test.getValue()
			if v == nil {
				s.testRaw(t, test.attr, nil, test.expect)
			} else {
				s.testRaw(t, test.attr, *v, test.expect)
			}
		})
	}
}

func (s *IntegerPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *int64
		expect   func(t *testing.T, unassigned bool)
	}{
		{
			name: "unassigned returns true",
			attr: s.standardAttr,
			getValue: func() *int64 {
				return nil
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "assigned returns false",
			attr: s.standardAttr,
			getValue: func() *int64 {
				i := int64(64)
				return &i
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.False(t, unassigned)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			v := test.getValue()
			if v == nil {
				s.testUnassigned(t, test.attr, nil, test.expect)
			} else {
				s.testUnassigned(t, test.attr, *v, test.expect)
			}
		})
	}
}

func (s *IntegerPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewInteger(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewInteger(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewIntegerOf(s.standardAttr, 64)
			},
			getB: func(t *testing.T) Property {
				return NewIntegerOf(s.standardAttr, 64)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewIntegerOf(s.standardAttr, 64)
			},
			getB: func(t *testing.T) Property {
				return NewIntegerOf(s.standardAttr, 32)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewIntegerOf(s.standardAttr, 64)
			},
			getB: func(t *testing.T) Property {
				return NewInteger(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewIntegerOf(s.standardAttr, 64)
			},
			getB: func(t *testing.T) Property {
				return NewInteger(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:someNumber",
  "name": "someNumber",
  "type": "integer",
  "_path": "someNumber",
  "_index": 10
}`)))
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testMatches(t, test.getA(t), test.getB(t), test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewInteger(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewIntegerOf(s.standardAttr, 64),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewInteger(s.standardAttr),
			value: int64(64),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(64), raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewIntegerOf(s.standardAttr, 64),
			value: int64(128),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(128), raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewIntegerOf(s.standardAttr, 64),
			value: "123",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidValue, errors.Unwrap(err))
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testAdd(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewInteger(s.standardAttr),
			value: int64(64),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(64), raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewIntegerOf(s.standardAttr, int64(64)),
			value: int64(128),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(128), raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewInteger(s.standardAttr),
			value: "123",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidValue, errors.Unwrap(err))
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testReplace(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewInteger(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewIntegerOf(s.standardAttr, 64),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testDelete(t, test.prop, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewIntegerOf(s.standardAttr, 64),
			v:      64,
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewIntegerOf(s.standardAttr, 64),
			v:      128,
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewInteger(s.standardAttr),
			v:      64,
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewInteger(s.standardAttr),
			v:      "123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testEqualTo(t, test.prop, test.v, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestGreaterThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  32,
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  128,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  "123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testGreaterThan(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestGreaterThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  32,
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  64,
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  128,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  "123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testGreaterThanOrEqualTo(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestLessThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  128,
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  32,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  "123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testLessThan(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestLessThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  128,
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  64,
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  32,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewIntegerOf(s.standardAttr, 64),
			value:  "123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testLessThanOrEqualTo(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewIntegerOf(s.standardAttr, 64),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewInteger(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *IntegerPropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*IntegerPropertyTestSuite)(nil)
)
