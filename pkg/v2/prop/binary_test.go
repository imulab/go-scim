package prop

import (
	"encoding/base64"
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestBinaryProperty(t *testing.T) {
	s := new(BinaryPropertyTestSuite)

	s.NewFunc = NewBinary
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewBinaryOf(attr, v.(string))
	}

	suite.Run(t, s)
}

type BinaryPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite
	OperatorTestSuite
	standardAttr *spec.Attribute
}

func (s *BinaryPropertyTestSuite) SetupSuite() {
	s.standardAttr = s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:2.0:User:x509Certificates.value",
  "name": "value",
  "type": "binary",
  "_path": "x509Certificates.value",
  "_index": 100
}`))
}

func (s *BinaryPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)
	}{
		{
			name: "new binary",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:2.0:User:x509Certificates.value", p.Attribute().ID())
				assert.Equal(t, spec.TypeBinary, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "new binary auto load subscribers",
			description: `
This test confirms the capability to automatically load registered subscribers associated with annotations in the
attribute. BinaryPropertyTestSuite is a dummy implementation of Subscriber, which is registered to annotation @Test.
`,
			before: func() {
				SubscriberFactory().Register("@Test", func(_ Property, _ map[string]interface{}) Subscriber {
					return s
				})
			},
			attr: s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:2.0:User:x509Certificates.value",
  "name": "value",
  "type": "binary",
  "_path": "x509Certificates.value",
  "_index": 100,
  "_annotations": {
    "@Test": {}
  }
}`)),
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "urn:ietf:params:scim:schemas:2.0:User:x509Certificates.value", p.Attribute().ID())
				assert.Equal(t, spec.TypeBinary, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
				assert.Len(t, p.(*binaryProperty).subscribers, 1)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testNew(t, test.before, test.attr, test.expect)
		})
	}
}

func (s *BinaryPropertyTestSuite) TestRaw() {
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
			name: "assigned returns boolean",
			attr: s.standardAttr,
			getValue: func() *string {
				b := s.base64("hello world")
				return &b
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, s.base64("hello world"), raw)
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

func (s *BinaryPropertyTestSuite) TestUnassigned() {
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
				b := s.base64("hello")
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

func (s *BinaryPropertyTestSuite) TestMatches() {
	tests := []struct {
		name   string
		getA   func(t *testing.T) Property
		getB   func(t *testing.T) Property
		expect func(t *testing.T, match bool)
	}{
		{
			name: "unassigned property of same attribute matches",
			getA: func(t *testing.T) Property {
				return NewBinary(s.standardAttr)
			},
			getB: func(t *testing.T) Property {
				return NewBinary(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute and value matches",
			getA: func(t *testing.T) Property {
				return NewBinaryOf(s.standardAttr, s.base64("hello"))
			},
			getB: func(t *testing.T) Property {
				return NewBinaryOf(s.standardAttr, s.base64("hello"))
			},
			expect: func(t *testing.T, match bool) {
				assert.True(t, match)
			},
		},
		{
			name: "assigned property of same attribute but different value does not match",
			getA: func(t *testing.T) Property {
				return NewBinaryOf(s.standardAttr, s.base64("hello"))
			},
			getB: func(t *testing.T) Property {
				return NewBinaryOf(s.standardAttr, s.base64("world"))
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "assigned property does not match with unassigned property",
			getA: func(t *testing.T) Property {
				return NewBinaryOf(s.standardAttr, s.base64("hello"))
			},
			getB: func(t *testing.T) Property {
				return NewBinary(s.standardAttr)
			},
			expect: func(t *testing.T, match bool) {
				assert.False(t, match)
			},
		},
		{
			name: "different attribute does not match",
			getA: func(t *testing.T) Property {
				return NewBinaryOf(s.standardAttr, s.base64("hello"))
			},
			getB: func(t *testing.T) Property {
				return NewBinary(s.mustAttribute(s.T(), strings.NewReader(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:someBinary",
  "name": "someBinary",
  "type": "binary",
  "_path": "someBinary",
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

func (s *BinaryPropertyTestSuite) TestClone() {
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "clone unassigned property",
			prop: NewBinary(s.standardAttr),
		},
		{
			name: "clone assigned property",
			prop: NewBinaryOf(s.standardAttr, s.base64("hello")),
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testClone(t, test.prop)
		})
	}
}

func (s *BinaryPropertyTestSuite) TestAdd() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "add to unassigned",
			prop:  NewBinary(s.standardAttr),
			value: s.base64("hello"),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, s.base64("hello"), raw)
			},
		},
		{
			name:  "add to assigned",
			prop:  NewBinaryOf(s.standardAttr, s.base64("hello")),
			value: s.base64("world"),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, s.base64("world"), raw)
			},
		},
		{
			name:  "add incompatible value",
			prop:  NewBinaryOf(s.standardAttr, s.base64("hello")),
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

func (s *BinaryPropertyTestSuite) TestReplace() {
	tests := []struct {
		name   string
		prop   Property
		value  interface{}
		expect func(t *testing.T, raw interface{}, err error)
	}{
		{
			name:  "replace unassigned",
			prop:  NewBinary(s.standardAttr),
			value: s.base64("hello"),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, s.base64("hello"), raw)
			},
		},
		{
			name:  "replace assigned",
			prop:  NewBinaryOf(s.standardAttr, s.base64("hello")),
			value: s.base64("world"),
			expect: func(t *testing.T, raw interface{}, err error) {
				assert.Nil(t, err)
				assert.Equal(t, s.base64("world"), raw)
			},
		},
		{
			name:  "replace incompatible value",
			prop:  NewBinary(s.standardAttr),
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

func (s *BinaryPropertyTestSuite) TestDelete() {
	tests := []struct {
		name   string
		prop   Property
		expect func(t *testing.T, err error)
	}{
		{
			name: "delete unassigned",
			prop: NewBinary(s.standardAttr),
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "delete assigned",
			prop: NewBinaryOf(s.standardAttr, s.base64("hello")),
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

func (s *BinaryPropertyTestSuite) TestEqualTo() {
	tests := []struct {
		name   string
		prop   Property
		v      interface{}
		expect bool
	}{
		{
			name:   "equal value",
			prop:   NewBinaryOf(s.standardAttr, s.base64("hello")),
			v:      s.base64("hello"),
			expect: true,
		},
		{
			name:   "unequal value",
			prop:   NewBinaryOf(s.standardAttr, s.base64("hello")),
			v:      s.base64("world"),
			expect: false,
		},
		{
			name:   "unassigned does not equal",
			prop:   NewBinary(s.standardAttr),
			v:      s.base64("hello"),
			expect: false,
		},
		{
			name:   "incompatible does not equal",
			prop:   NewBinaryOf(s.standardAttr, s.base64("hello")),
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

func (s *BinaryPropertyTestSuite) TestPresent() {
	tests := []struct {
		name   string
		prop   Property
		expect bool
	}{
		{
			name:   "assigned is present",
			prop:   NewBinaryOf(s.standardAttr, s.base64("hello")),
			expect: true,
		},
		{
			name:   "unassigned is not present",
			prop:   NewBinary(s.standardAttr),
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			s.testPresent(t, test.prop, test.expect)
		})
	}
}

func (s *BinaryPropertyTestSuite) base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (s *BinaryPropertyTestSuite) Notify(_ Property, _ *Events) error {
	return nil
}

var (
	_ Subscriber = (*BinaryPropertyTestSuite)(nil)
)
