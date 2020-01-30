package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestBooleanProperty(t *testing.T) {
	s := new(BooleanPropertyTestSuite)

	s.NewFunc = NewBoolean
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewBooleanOf(attr, v.(bool))
	}

	suite.Run(t, s)
}

type BooleanPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *BooleanPropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:2.0:User:active",
  "name": "active",
  "type": "boolean",
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "active",
  "_index": 100
}`))
}

func (s *BooleanPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new boolean",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:2.0:User:active", p.Attribute().ID())
				assert.Equal(t, spec.TypeBoolean, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new boolean auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. BooleanPropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:2.0:User:active",
  "name": "active",
  "type": "boolean",
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "active",
  "_index": 100,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:2.0:User:active", p.Attribute().ID())
				assert.Equal(t, spec.TypeBoolean, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*booleanProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *BooleanPropertyTestSuite) TestRaw() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *bool
		expect   func(t *testing.T, raw interface{})
	}{
		{
			name: "unassigned returns nil",
			attr: s.standardAttr,
			getValue: func() *bool {
				return nil
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Nil(t, raw)
			},
		},
		{
			name: "assigned returns boolean",
			attr: s.standardAttr,
			getValue: func() *bool {
				b := true
				return &b
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, true, raw)
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

func (s *BooleanPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() *bool
		expect   func(t *testing.T, unassigned bool)
	}{
		{
			name: "unassigned returns true",
			attr: s.standardAttr,
			getValue: func() *bool {
				return nil
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "assigned returns false",
			attr: s.standardAttr,
			getValue: func() *bool {
				b := true
				return &b
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

func (s *BooleanPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewBoolean(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewBoolean(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewBooleanOf(s.standardAttr, true)
			},
			getB: func(t *testing.T) Property {
				return NewBooleanOf(s.standardAttr, true)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewBooleanOf(s.standardAttr, true)
			},
			getB: func(t *testing.T) Property {
				return NewBooleanOf(s.standardAttr, false)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewBooleanOf(s.standardAttr, true)
			},
			getB: func(t *testing.T) Property {
				return NewBoolean(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewBooleanOf(s.standardAttr, true)
			},
			getB: func(t *testing.T) Property {
				return NewBoolean(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:someBool",
  "name": "someBool",
  "type": "boolean",
  "_path": "someBool",
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

func (s *BooleanPropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewBoolean(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewBooleanOf(s.standardAttr, false),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *BooleanPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewBoolean(s.standardAttr),
			value: true,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, true, raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewBooleanOf(s.standardAttr, true),
			value: false,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, false, raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewBooleanOf(s.standardAttr, true),
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

func (s *BooleanPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewBoolean(s.standardAttr),
			value: true,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, true, raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewBooleanOf(s.standardAttr, true),
			value: false,
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, false, raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewBoolean(s.standardAttr),
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

func (s *BooleanPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewBoolean(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewBooleanOf(s.standardAttr, true),
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

func (s *BooleanPropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewBooleanOf(s.standardAttr, true),
			v:      true,
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewBooleanOf(s.standardAttr, true),
			v:      false,
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewBoolean(s.standardAttr),
			v:      true,
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewBoolean(s.standardAttr),
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

func (s *BooleanPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewBooleanOf(s.standardAttr, true),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewBoolean(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *BooleanPropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*BooleanPropertyTestSuite)(nil)
)
