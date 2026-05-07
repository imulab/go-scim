package spec

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestAttribute(t *testing.T) {
	s := new(AttributeTestSuite)
	suite.Run(t, s)
}

type AttributeTestSuite struct {
	suite.Suite
}

func (s *AttributeTestSuite) TestMarshal() {
	attr := &Attribute{
		id:    "urn:ietf:params:scim:schemas:core:2.0:User:emails",
		name:  "emails",
		typ:   TypeComplex,
		index: 110,
		path:  "emails",
		annotations: map[string]map[string]interface{}{
			"@AutoCompact":      {},
			"@ExclusivePrimary": {},
		},
		subAttributes: []*Attribute{
			{
				id:    "urn:ietf:params:scim:schemas:core:2.0:User:emails.value",
				name:  "value",
				typ:   TypeString,
				index: 0,
				path:  "emails.value",
			},
			{
				id:    "urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
				name:  "primary",
				typ:   TypeBoolean,
				index: 1,
				path:  "emails.primary",
			},
		},
	}

	raw, err := json.Marshal(attr)
	assert.Nil(s.T(), err)

	expect := `
{
  "name": "emails",
  "type": "complex",
  "subAttributes": [
    {
      "name": "value",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none"
    },
    {
      "name": "primary",
      "type": "boolean",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none"
    }
  ],
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none"
}
`
	assert.JSONEq(s.T(), expect, string(raw))
}

func (s *AttributeTestSuite) TestUnmarshal() {
	raw := `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:emails",
  "name": "emails",
  "type": "complex",
  "subAttributes": [
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:emails.value",
      "name": "value",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none",
      "_path": "emails.value",
      "_index": 0
    },
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
      "name": "primary",
      "type": "boolean",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none",
      "_path": "emails.primary",
      "_index": 1
    }
  ],
  "multiValued": false,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "emails",
  "_index": 110,
  "_annotations": {
    "@AutoCompact": {},
    "@ExclusivePrimary": {}
  }
}
`

	attr := new(Attribute)
	err := json.Unmarshal([]byte(raw), attr)
	assert.Nil(s.T(), err)
}

func (s *AttributeTestSuite) TestGoesBy() {
	tests := []struct {
		name   string
		attr   *Attribute
		goesBy string
		expect bool
	}{
		{
			name:   "Goes by attribute name",
			attr:   &Attribute{name: "foobar"},
			goesBy: "foobar",
			expect: true,
		},
		{
			name:   "Goes by attribute name (case insensitive)",
			attr:   &Attribute{name: "foobar"},
			goesBy: "FOOBAR",
			expect: true,
		},
		{
			name:   "Goes by attribute path",
			attr:   &Attribute{name: "bar", path: "foo.bar"},
			goesBy: "foo.bar",
			expect: true,
		},
		{
			name:   "Goes by attribute path",
			attr:   &Attribute{name: "bar", path: "foo.bar", id: "urn:test:foo.bar"},
			goesBy: "urn:test:foo.bar",
			expect: true,
		},
		{
			name:   "Not goes by unrelated name",
			attr:   &Attribute{name: "bar", path: "foo.bar", id: "urn:test:foo.bar"},
			goesBy: "unrelated",
			expect: false,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.attr.GoesBy(test.goesBy))
		})
	}
}

func (s *AttributeTestSuite) TestDeriveElementAttribute() {
	raw := `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:emails",
  "name": "emails",
  "type": "complex",
  "subAttributes": [
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:emails.value",
      "name": "value",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none",
      "_path": "emails.value",
      "_index": 0
    },
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
      "name": "primary",
      "type": "boolean",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none",
      "_path": "emails.primary",
      "_index": 1
    }
  ],
  "multiValued": true,
  "required": false,
  "caseExact": false,
  "mutability": "readWrite",
  "returned": "default",
  "uniqueness": "none",
  "_path": "emails",
  "_index": 110,
  "_annotations": {
    "@AutoCompact": {},
    "@ExclusivePrimary": {},
	"@ElementAnnotations": {
      "@AutoCompact": {},
      "@ExclusivePrimary": {}
	}
  }
}
`
	attr := new(Attribute)
	err := json.Unmarshal([]byte(raw), attr)
	require.Nil(s.T(), err)

	elemAttr := attr.DeriveElementAttribute()
	assert.Equal(s.T(), "urn:ietf:params:scim:schemas:core:2.0:User:emails$elem", elemAttr.id)
	assert.False(s.T(), elemAttr.multiValued)
	for _, annotation := range []string{"@AutoCompact", "@ExclusivePrimary"} {
		_, ok := elemAttr.Annotation(annotation)
		assert.True(s.T(), ok)
	}
}
