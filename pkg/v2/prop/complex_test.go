package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestComplexProperty(t *testing.T) {
	s := new(ComplexPropertyTestSuite)

	s.NewFunc = NewComplex
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewComplexOf(attr, v.(map[string]interface{}))
	}

	suite.Run(t, s)
}

type ComplexPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *ComplexPropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:name",
  "name": "name",
  "type": "complex",
  "_path": "name",
  "_index": 10,
  "subAttributes": [
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
      "name": "givenName",
      "type": "string",
      "_path": "name.givenName",
      "_index": 0
    },
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
      "name": "familyName",
      "type": "string",
      "_path": "name.familyName",
      "_index": 1
    }
  ]
}`))
}

func (s *ComplexPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new string of string attribute",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:name", p.Attribute().ID())
				assert.Equal(t, spec.TypeComplex, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new string auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. ComplexPropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:name",
  "name": "name",
  "type": "complex",
  "_path": "name",
  "_index": 10,
  "subAttributes": [
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
      "name": "givenName",
      "type": "string",
      "_path": "name.givenName",
      "_index": 0
    },
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
      "name": "familyName",
      "type": "string",
      "_path": "name.familyName",
      "_index": 1
    }
  ],
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:name", p.Attribute().ID())
				assert.Equal(t, spec.TypeComplex, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*complexProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *ComplexPropertyTestSuite) TestRaw() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() map[string]interface{}
		expect   func(t *testing.T, raw interface{})
	}{
		{
			name: "unassigned returns nil",
			attr: s.standardAttr,
			getValue: func() map[string]interface{} {
				return nil
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, map[string]interface{}{
					"givenName":  nil,
					"familyName": nil,
				}, raw)
			},
		},
		{
			name: "assigned returns string",
			attr: s.standardAttr,
			getValue: func() map[string]interface{} {
				return map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				}
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				}, raw)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			v := test.getValue()
			if v == nil {
				s.testRaw(t, test.attr, nil, test.expect)
			} else {
				s.testRaw(t, test.attr, v, test.expect)
			}
		})
	}
}

func (s *ComplexPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() map[string]interface{}
		expect   func(t *testing.T, unassigned bool)
	}{
		{
			name: "unassigned returns true",
			attr: s.standardAttr,
			getValue: func() map[string]interface{} {
				return nil
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "assigned returns false",
			attr: s.standardAttr,
			getValue: func() map[string]interface{} {
				return map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				}
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
				s.testUnassigned(t, test.attr, v, test.expect)
			}
		})
	}
}

func (s *ComplexPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewComplex(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewComplex(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewComplexOf(s.standardAttr, map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				})
			},
			getB: func(t *testing.T) Property {
				return NewComplexOf(s.standardAttr, map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				})
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewComplexOf(s.standardAttr, map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				})
			},
			getB: func(t *testing.T) Property {
				return NewComplexOf(s.standardAttr, map[string]interface{}{
					"givenName":  "Weinan",
					"familyName": "Q",
				})
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewComplexOf(s.standardAttr, map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				})
			},
			getB: func(t *testing.T) Property {
				return NewComplex(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewComplexOf(s.standardAttr, map[string]interface{}{
					"givenName":  "David",
					"familyName": "Q",
				})
			},
			getB: func(t *testing.T) Property {
				return NewComplex(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "meta",
  "name": "meta",
  "type": "complex",
  "_path": "meta",
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

func (s *ComplexPropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewComplex(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewComplexOf(s.standardAttr, map[string]interface{}{
				"givenName": "David",
			}),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *ComplexPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: "add to unassigned",
			prop: NewComplex(s.standardAttr),
			value: map[string]interface{}{
				"givenName": "David",
			},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, map[string]interface{}{
					"givenName":  "David",
					"familyName": nil,
				}, raw)
			},
		},
		{
			name: "add to assigned",
			prop: NewComplexOf(s.standardAttr, map[string]interface{}{
				"givenName": "David",
			}),
			value: map[string]interface{}{
				"givenName":  "Weinan",
				"familyName": "Q",
			},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, map[string]interface{}{
					"givenName":  "Weinan",
					"familyName": "Q",
				}, raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewComplex(s.standardAttr),
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

func (s *ComplexPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name: "replace unassigned",
			prop: NewComplex(s.standardAttr),
			value: map[string]interface{}{
				"givenName":  "Weinan",
				"familyName": "Q",
			},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, map[string]interface{}{
					"givenName":  "Weinan",
					"familyName": "Q",
				}, raw)
			},
		},
		{
			name: "replace assigned",
			prop: NewComplexOf(s.standardAttr, map[string]interface{}{
				"givenName":  "Weinan",
				"familyName": "Q",
			}),
			value: map[string]interface{}{
				"givenName": "David",
			},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, map[string]interface{}{
					"givenName":  "David",
					"familyName": nil,
				}, raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewComplex(s.standardAttr),
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

func (s *ComplexPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewComplex(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewComplexOf(s.standardAttr, map[string]interface{}{
				"givenName":  "David",
				"familyName": nil,
			}),
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

func (s *ComplexPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name: "assigned is present",
			prop: NewComplexOf(s.standardAttr, map[string]interface{}{
				"givenName":  "David",
				"familyName": nil,
			}),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewComplex(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *ComplexPropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*ComplexPropertyTestSuite)(nil)
)
