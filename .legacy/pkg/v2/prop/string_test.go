package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestStringProperty(t *testing.T) {
	s := new(StringPropertyTestSuite)

	s.NewFunc = NewString
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewStringOf(attr, v.(string))
	}

	suite.Run(t, s)
}

type StringPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *StringPropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:userName",
  "name": "userName",
  "type": "string",
  "multiValued": false,
  "required": true,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "userName",
  "_index": 10
}`))
}

func (s *StringPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new string of string attribute",
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:userName",
  "name": "userName",
  "type": "string",
  "multiValued": false,
  "required": true,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "userName",
  "_index": 10
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:userName", p.Attribute().ID())
				assert.Equal(t, spec.TypeString, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new string auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. StringPropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:userName",
  "name": "userName",
  "type": "string",
  "multiValued": false,
  "required": true,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "userName",
  "_index": 10,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:userName", p.Attribute().ID())
				assert.Equal(t, spec.TypeString, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*stringProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestRaw() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *string
		expect   func(t *testing.T, raw interface{})
	}{
		{
			name: "unassigned returns nil",
			attr: s.standardAttr,
			getValue: func() *string {
				return nil
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Nil(t, raw)
			},
		},
		{
			name: "assigned returns string",
			attr: s.standardAttr,
			getValue: func() *string {
				str := "foobar"
				return &str
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, "foobar", raw)
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

func (s *StringPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *string
		expect   func(t *testing.T, unassigned bool)
	}{
		{
			name: "unassigned returns true",
			attr: s.standardAttr,
			getValue: func() *string {
				return nil
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "assigned returns false",
			attr: s.standardAttr,
			getValue: func() *string {
				str := "foobar"
				return &str
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

func (s *StringPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewString(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewString(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewStringOf(s.standardAttr, "foobar")
			},
			getB: func(t *testing.T) Property {
				return NewStringOf(s.standardAttr, "foobar")
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewStringOf(s.standardAttr, "foo")
			},
			getB: func(t *testing.T) Property {
				return NewStringOf(s.standardAttr, "bar")
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewStringOf(s.standardAttr, "foo")
			},
			getB: func(t *testing.T) Property {
				return NewString(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewStringOf(s.standardAttr, "foo")
			},
			getB: func(t *testing.T) Property {
				return NewString(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:displayName",
  "name": "displayName",
  "type": "string",
  "_path": "displayName",
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

func (s *StringPropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewString(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewStringOf(s.standardAttr, "foobar"),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *StringPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewString(s.standardAttr),
			value: "foobar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewStringOf(s.standardAttr, "foo"),
			value: "bar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "bar", raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewString(s.standardAttr),
			value: 123,
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

func (s *StringPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewString(s.standardAttr),
			value: "foobar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewStringOf(s.standardAttr, "foo"),
			value: "bar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "bar", raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewString(s.standardAttr),
			value: 123,
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

func (s *StringPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewString(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewStringOf(s.standardAttr, "foo"),
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

func (s *StringPropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "foobar",
			expect: true,
		},
		{
			name:   "equal value case insensitive",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "FOOBAR",
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewString(s.standardAttr),
			v:      "random",
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewString(s.standardAttr),
			v:      123,
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testEqualTo(t, test.prop, test.v, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestStartsWith() {
	tests := []struct {
		name   string
		prop   Property
		v      string
		expect bool
	}{
		{
			name:   "starts with prefix",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "foo",
			expect: true,
		},
		{
			name:   "starts with prefix case insensitive",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "FOO",
			expect: true,
		},
		{
			name:   "does not start with",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned",
			prop:   NewString(s.standardAttr),
			v:      "random",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testStartsWith(t, test.prop, test.v, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestEndsWith() {
	tests := []struct {
		name   string
		prop   Property
		v      string
		expect bool
	}{
		{
			name:   "ends with suffix",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "bar",
			expect: true,
		},
		{
			name:   "ends with suffix case insensitive",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "BAR",
			expect: true,
		},
		{
			name:   "does not end with",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned",
			prop:   NewString(s.standardAttr),
			v:      "random",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testEndsWith(t, test.prop, test.v, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestContains() {
	tests := []struct {
		name   string
		prop   Property
		v      string
		expect bool
	}{
		{
			name:   "contains",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "oo",
			expect: true,
		},
		{
			name:   "contains case insensitive",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "Oo",
			expect: true,
		},
		{
			name:   "does not contain",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned",
			prop:   NewString(s.standardAttr),
			v:      "random",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testContains(t, test.prop, test.v, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestGreaterThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "a",
			expect: true,
		},
		{
			name:   "greater than case insensitive",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "A",
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "z",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  123,
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testGreaterThan(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestGreaterThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "a",
			expect: true,
		},
		{
			name:   "greater than case insensitive",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "A",
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "m",
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "z",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  123,
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testGreaterThanOrEqualTo(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestLessThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "z",
			expect: true,
		},
		{
			name:   "less than case insensitive",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "Z",
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "a",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  123,
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testLessThan(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestLessThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "z",
			expect: true,
		},
		{
			name:   "less than case insensitive",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "Z",
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "m",
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  "a",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewStringOf(s.standardAttr, "m"),
			value:  123,
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testLessThanOrEqualTo(t, test.prop, test.value, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewStringOf(s.standardAttr, "foobar"),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewString(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *StringPropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*StringPropertyTestSuite)(nil)
)
