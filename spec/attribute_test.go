package spec

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
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
