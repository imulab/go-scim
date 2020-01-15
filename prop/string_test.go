package prop

import (
	"errors"
	"github.com/elvsn/scim.go/spec"
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
				SubscriberFactory().Register("@Test", func(params map[string]interface{}) Subscriber {
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
			s.testRaw(t, test.attr, test.getValue(), test.expect)
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
			s.testUnassigned(t, test.attr, test.getValue(), test.expect)
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

func (s *StringPropertyTestSuite) Notify(_ Property, _ []*Event) error {
	return nil
}

var (
	_ Subscriber = (*StringPropertyTestSuite)(nil)
)
