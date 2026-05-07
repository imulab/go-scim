package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestReferenceProperty(t *testing.T) {
	s := new(ReferencePropertyTestSuite)

	s.NewFunc = NewReference
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewReferenceOf(attr, v.(string))
	}

	suite.Run(t, s)
}

type ReferencePropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *ReferencePropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "meta.location",
  "name": "location",
  "type": "reference",
  "mutability": "readOnly",
  "_path": "meta.location",
  "_index": 10
}`))
}

func (s *ReferencePropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new reference",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "meta.location", p.Attribute().ID())
				assert.Equal(t, spec.TypeReference, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new reference auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. ReferencePropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "meta.location",
  "name": "location",
  "type": "reference",
  "mutability": "readOnly",
  "_path": "meta.location",
  "_index": 10,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "meta.location", p.Attribute().ID())
				assert.Equal(t, spec.TypeReference, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*referenceProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *ReferencePropertyTestSuite) TestRaw() {
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

func (s *ReferencePropertyTestSuite) TestUnassigned() {
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

func (s *ReferencePropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewReference(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewReference(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewReferenceOf(s.standardAttr, "foobar")
			},
			getB: func(t *testing.T) Property {
				return NewReferenceOf(s.standardAttr, "foobar")
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewReferenceOf(s.standardAttr, "foo")
			},
			getB: func(t *testing.T) Property {
				return NewReferenceOf(s.standardAttr, "bar")
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewReferenceOf(s.standardAttr, "foo")
			},
			getB: func(t *testing.T) Property {
				return NewReference(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewReferenceOf(s.standardAttr, "foo")
			},
			getB: func(t *testing.T) Property {
				return NewReference(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:someRef",
  "name": "someRef",
  "type": "reference",
  "_path": "someRef",
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

func (s *ReferencePropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewReference(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewReferenceOf(s.standardAttr, "foobar"),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *ReferencePropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewReference(s.standardAttr),
			value: "foobar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewReferenceOf(s.standardAttr, "foo"),
			value: "bar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "bar", raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewReference(s.standardAttr),
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

func (s *ReferencePropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewReference(s.standardAttr),
			value: "foobar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewReferenceOf(s.standardAttr, "foo"),
			value: "bar",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "bar", raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewReference(s.standardAttr),
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

func (s *ReferencePropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewReference(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewReferenceOf(s.standardAttr, "foo"),
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

func (s *ReferencePropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "foobar",
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewReference(s.standardAttr),
			v:      "random",
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewReference(s.standardAttr),
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

func (s *ReferencePropertyTestSuite) TestStartsWith() {
	tests := []struct {
		name   string
		prop   Property
		v      string
		expect bool
	}{
		{
			name:   "starts with prefix",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "foo",
			expect: true,
		},
		{
			name:   "does not start with",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned",
			prop:   NewReference(s.standardAttr),
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

func (s *ReferencePropertyTestSuite) TestEndsWith() {
	tests := []struct {
		name   string
		prop   Property
		v      string
		expect bool
	}{
		{
			name:   "ends with suffix",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "bar",
			expect: true,
		},
		{
			name:   "does not end with",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned",
			prop:   NewReference(s.standardAttr),
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

func (s *ReferencePropertyTestSuite) TestContains() {
	tests := []struct {
		name   string
		prop   Property
		v      string
		expect bool
	}{
		{
			name:   "contains",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "oo",
			expect: true,
		},
		{
			name:   "does not contain",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			v:      "random",
			expect: false,
		},
		{
			name:   "unassigned",
			prop:   NewReference(s.standardAttr),
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

func (s *ReferencePropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewReferenceOf(s.standardAttr, "foobar"),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewReference(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *ReferencePropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*ReferencePropertyTestSuite)(nil)
)
