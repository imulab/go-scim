package prop

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestMultiValuedProperty(t *testing.T) {
	s := new(MultiValuedPropertyTestSuite)

	s.NewFunc = NewMulti
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewMultiOf(attr, v.([]interface{}))
	}

	suite.Run(t, s)
}

type MultiValuedPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *MultiValuedPropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "schemas",
  "name": "schemas",
  "type": "string",
  "multiValued": true,
  "_path": "schemas",
  "_index": 0
}`))
}

func (s *MultiValuedPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new multiValued",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "schemas", p.Attribute().ID())
				assert.Equal(t, spec.TypeString, p.Attribute().Type())
				assert.True(t, p.Attribute().MultiValued())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new multiValued auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. MultiValuedPropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "schemas",
  "name": "schemas",
  "type": "string",
  "multiValued": true,
  "_path": "schemas",
  "_index": 0,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "schemas", p.Attribute().ID())
				assert.Equal(t, spec.TypeString, p.Attribute().Type())
				assert.True(t, p.Attribute().MultiValued())
				assert.Len(t, p.(*multiValuedProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *MultiValuedPropertyTestSuite) TestRaw() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() []interface{}
		expect   func(t *testing.T, raw interface{})
	}{
		{
			name: "unassigned returns nil",
			attr: s.standardAttr,
			getValue: func() []interface{} {
				return nil
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Nil(t, raw)
			},
		},
		{
			name: "assigned returns string",
			attr: s.standardAttr,
			getValue: func() []interface{} {
				return []interface{}{"A", "B"}
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{"A", "B"}, raw)
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

func (s *MultiValuedPropertyTestSuite) TestUnassigned() {
	tests := []struct {
		name     string
		attr     *spec.Attribute
		getValue func() []interface{}
		expect   func(t *testing.T, unassigned bool)
	}{
		{
			name: "unassigned returns true",
			attr: s.standardAttr,
			getValue: func() []interface{} {
				return nil
			},
			expect: func(t *testing.T, unassigned bool) {
				assert.True(t, unassigned)
			},
		},
		{
			name: "assigned returns false",
			attr: s.standardAttr,
			getValue: func() []interface{} {
				return []interface{}{"A", "B"}
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

func (s *MultiValuedPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewMulti(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewMulti(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewMultiOf(s.standardAttr, []interface{}{"A", "B"})
			},
			getB: func(t *testing.T) Property {
				return NewMultiOf(s.standardAttr, []interface{}{"B", "A"})
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewMultiOf(s.standardAttr, []interface{}{"A", "B"})
			},
			getB: func(t *testing.T) Property {
				return NewMultiOf(s.standardAttr, []interface{}{"A", "C"})
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewMultiOf(s.standardAttr, []interface{}{"A", "B"})
			},
			getB: func(t *testing.T) Property {
				return NewMulti(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewMultiOf(s.standardAttr, []interface{}{"A", "B"})
			},
			getB: func(t *testing.T) Property {
				return NewMulti(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "schemas2",
  "name": "schemas2",
  "type": "string",
  "multiValued": true,
  "_path": "schemas2",
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

func (s *MultiValuedPropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewMulti(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewMultiOf(s.standardAttr, []interface{}{"A"}),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *MultiValuedPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewMulti(s.standardAttr),
			value: []interface{}{"A"},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"A"}, raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewMultiOf(s.standardAttr, []interface{}{"A"}),
			value: []interface{}{"B"},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"A", "B"}, raw)
			},
		},
		{
			name:  "add existing",
			prop:  NewMultiOf(s.standardAttr, []interface{}{"A"}),
			value: []interface{}{"A"},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"A"}, raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewMulti(s.standardAttr),
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

func (s *MultiValuedPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewMulti(s.standardAttr),
			value: []interface{}{"A"},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"A"}, raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewMultiOf(s.standardAttr, []interface{}{"A"}),
			value: []interface{}{"B"},
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"B"}, raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewMulti(s.standardAttr),
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

func (s *MultiValuedPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewMulti(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewMultiOf(s.standardAttr, []interface{}{"A"}),
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

func (s *MultiValuedPropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewMultiOf(s.standardAttr, []interface{}{"A", "B"}),
			v:      "B",
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewMultiOf(s.standardAttr, []interface{}{"A", "B"}),
			v:      "C",
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewMulti(s.standardAttr),
			v:      "A",
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewMultiOf(s.standardAttr, []interface{}{"A"}),
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

func (s *MultiValuedPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewMultiOf(s.standardAttr, []interface{}{"A"}),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewMulti(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *MultiValuedPropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*MultiValuedPropertyTestSuite)(nil)
)
