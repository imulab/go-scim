package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestDateTimeProperty(t *testing.T) {
	s := new(DateTimePropertyTestSuite)

	s.NewFunc = NewDateTime
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewDateTimeOf(attr, v.(string))
	}

	suite.Run(t, s)
}

type DateTimePropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *DateTimePropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "meta.created",
  "name": "created",
  "type": "dateTime",
  "mutability": "readOnly",
  "_path": "meta.created",
  "_index": 2
}`))
}

func (s *DateTimePropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new dateTime",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "meta.created", p.Attribute().ID())
				assert.Equal(t, spec.TypeDateTime, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new dateTime auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. DateTimePropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "meta.created",
  "name": "created",
  "type": "dateTime",
  "mutability": "readOnly",
  "_path": "meta.created",
  "_index": 2,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "meta.created", p.Attribute().ID())
				assert.Equal(t, spec.TypeDateTime, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*dateTimeProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *DateTimePropertyTestSuite) TestRaw() {
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
				d := "2020-01-16T07:30:00"
				return &d
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, "2020-01-16T07:30:00", raw)
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

func (s *DateTimePropertyTestSuite) TestUnassigned() {
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
				d := "2020-01-16T07:30:00"
				return &d
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

func (s *DateTimePropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewDateTime(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewDateTime(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewDateTimeOf(s.standardAttr, "2019-01-16T07:30:00")
			},
			getB: func(t *testing.T) Property {
				return NewDateTimeOf(s.standardAttr, "2019-01-16T07:30:00")
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewDateTimeOf(s.standardAttr, "2019-01-16T07:30:00")
			},
			getB: func(t *testing.T) Property {
				return NewDateTimeOf(s.standardAttr, "2019-01-17T07:30:00")
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewDateTimeOf(s.standardAttr, "2019-01-16T07:30:00")
			},
			getB: func(t *testing.T) Property {
				return NewDateTime(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewDateTimeOf(s.standardAttr, "2019-01-16T07:30:00")
			},
			getB: func(t *testing.T) Property {
				return NewDateTime(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:someDateTime",
  "name": "someDateTime",
  "type": "dateTime",
  "_path": "someDateTime",
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

func (s *DateTimePropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewDateTime(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *DateTimePropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewDateTime(s.standardAttr),
			value: "2020-01-16T07:30:00",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "2020-01-16T07:30:00", raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value: "2020-01-17T07:30:00",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "2020-01-17T07:30:00", raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
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

func (s *DateTimePropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewDateTime(s.standardAttr),
			value: "2020-01-16T07:30:00",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "2020-01-16T07:30:00", raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value: "2020-01-17T07:30:00",
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "2020-01-17T07:30:00", raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewDateTime(s.standardAttr),
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

func (s *DateTimePropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewDateTime(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
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

func (s *DateTimePropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			v:      "2020-01-16T07:30:00",
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			v:      "2020-01-17T07:30:00",
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewDateTime(s.standardAttr),
			v:      "2020-01-16T07:30:00",
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewDateTime(s.standardAttr),
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

func (s *DateTimePropertyTestSuite) TestGreaterThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-15T07:30:00",
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-17T07:30:00",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
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

func (s *DateTimePropertyTestSuite) TestGreaterThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "greater than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-15T07:30:00",
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-16T07:30:00",
			expect: true,
		},
		{
			name:   "not greater than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-17T07:30:00",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
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

func (s *DateTimePropertyTestSuite) TestLessThan() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-17T07:30:00",
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-15T07:30:00",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
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

func (s *DateTimePropertyTestSuite) TestLessThanOrEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect bool
	}{
		{
			name:   "less than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-17T07:30:00",
			expect: true,
		},
		{
			name:   "equal",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-16T07:30:00",
			expect: true,
		},
		{
			name:   "not less than",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			value:  "2020-01-15T07:30:00",
			expect: false,
		},
		{
			name:   "incompatible",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
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

func (s *DateTimePropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewDateTimeOf(s.standardAttr, "2020-01-16T07:30:00"),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewDateTime(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *DateTimePropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*DateTimePropertyTestSuite)(nil)
)
