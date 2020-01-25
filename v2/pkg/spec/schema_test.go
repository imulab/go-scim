package spec

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestSchema(t *testing.T) {
	s := new(SchemaTestSuite)
	suite.Run(t, s)
}

type SchemaTestSuite struct {
	suite.Suite
}

func (s *SchemaTestSuite) TestMarshal() {
	schema := &Schema{
		id:          "urn:ietf:params:scim:schemas:core:2.0:User",
		name:        "User",
		description: "User schema",
		attributes: []*Attribute{
			{name: "foobar", typ: TypeString},
		},
	}

	raw, err := json.Marshal(schema)
	assert.Nil(s.T(), err)

	expect := `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "User schema",
  "attributes": [
    {
      "name": "foobar",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none"
    }
  ]
}
`
	assert.JSONEq(s.T(), expect, string(raw))
}

func (s *SchemaTestSuite) TestUnmarshal() {
	raw := `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "User schema",
  "attributes": [
    {
      "name": "foobar",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "none"
    }
  ]
}
`

	schema := new(Schema)
	err := json.Unmarshal([]byte(raw), schema)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "urn:ietf:params:scim:schemas:core:2.0:User", schema.ID())
	assert.Equal(s.T(), "User", schema.Name())
	assert.Len(s.T(), schema.attributes, 1)
}
