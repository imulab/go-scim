package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestDecimalProperty(t *testing.T) {
	s := new(DecimalPropertyTestSuite)

	s.NewFunc = NewDecimal
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewDecimalOf(attr, v.(float64))
	}

	suite.Run(t, s)
}

type DecimalPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *DecimalPropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:extension:test:2.0:User:score",
  "name": "score",
  "type": "decimal",
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "score",
  "_index": 100
}`))
}

func (s *DecimalPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new decimal of decimal attribute",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:extension:test:2.0:User:score", p.Attribute().ID())
				assert.Equal(t, spec.TypeDecimal, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new decimal auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. DecimalPropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:extension:test:2.0:User:score",
  "name": "score",
  "type": "decimal",
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "score",
  "_index": 100,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:extension:test:2.0:User:score", p.Attribute().ID())
				assert.Equal(t, spec.TypeDecimal, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*decimalProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *DecimalPropertyTestSuite) TestRaw() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *float64
		expect   func(t *testing.T, raw interface{})
	}{
		{
			name: "unassigned returns nil",
			attr: s.standardAttr,
			getValue: func() *float64 {
				return nil
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Nil(t, raw)
			},
		},
		{
			name: "assigned returns decimal",
			attr: s.standardAttr,
			getValue: func() *float64 {
				f := 100.123
				return &f
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, 100.123, raw)
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

func (s *DecimalPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *float64
		expect   func(t *testing.T, unassigned bool)
	}{
		{
			name: "unassigned returns true",
			attr: s.standardAttr,
			getValue: func() *float64 {
				return nil
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "assigned returns false",
			attr: s.standardAttr,
			getValue: func() *float64 {
				f := 100.123
				return &f
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

func (s *DecimalPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewDecimal(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewDecimal(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewDecimalOf(s.standardAttr, 100.123)
			},
			getB: func(t *testing.T) Property {
				return NewDecimalOf(s.standardAttr, 100.123)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewDecimalOf(s.standardAttr, 100.123)
			},
			getB: func(t *testing.T) Property {
				return NewDecimalOf(s.standardAttr, 200.123)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewDecimalOf(s.standardAttr, 100.123)
			},
			getB: func(t *testing.T) Property {
				return NewDecimal(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewDecimalOf(s.standardAttr, 100.123)
			},
			getB: func(t *testing.T) Property {
				return NewDecimal(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:someDecimal",
  "name": "someDecimal",
  "type": "decimal",
  "_path": "someDecimal",
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

func (s *DecimalPropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewDecimal(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewDecimalOf(s.standardAttr, 100.123),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *DecimalPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewDecimal(s.standardAttr),
			value: 100.123,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 100.123, raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewDecimalOf(s.standardAttr, 100.123),
			value: 200.123,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 200.123, raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewDecimalOf(s.standardAttr, 100.123),
			value: "200.123",
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

func (s *DecimalPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewDecimal(s.standardAttr),
			value: 100.123,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 100.123, raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewDecimalOf(s.standardAttr, 100.123),
			value: 200.123,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 200.123, raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewDecimal(s.standardAttr),
			value: "100.123",
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

func (s *DecimalPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewDecimal(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewDecimalOf(s.standardAttr, 100.123),
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

func (s *DecimalPropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			v:      100.123,
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			v:      200.123,
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewDecimal(s.standardAttr),
			v:      100.123,
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewDecimal(s.standardAttr),
			v:      "100.123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testEqualTo(t, test.prop, test.v, test.expect)
		})
	}
}

func (s *DecimalPropertyTestSuite) TestGreaterThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  50.123,
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  200.123,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  "200.123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testGreaterThan(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *DecimalPropertyTestSuite) TestGreaterThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  50.123,
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  100.123,
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  200.123,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  "200.123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testGreaterThanOrEqualTo(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *DecimalPropertyTestSuite) TestLessThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  200.123,
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  50.123,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  "100.123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testLessThan(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *DecimalPropertyTestSuite) TestLessThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  200.123,
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  100.123,
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  50.123,
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDecimalOf(s.standardAttr, 100.123),
			value:  "100.123",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testLessThanOrEqualTo(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *DecimalPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewDecimalOf(s.standardAttr, 64.123),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewDecimal(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *DecimalPropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*DecimalPropertyTestSuite)(nil)
)
