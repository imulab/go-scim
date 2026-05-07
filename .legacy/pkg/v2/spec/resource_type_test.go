package spec

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestResourceType(t *testing.T) {
	s := new(ResourceTypeTestSuite)
	suite.Run(t, s)
}

type ResourceTypeTestSuite struct {
	suite.Suite
}

func (s *ResourceTypeTestSuite) TestMarshal() {
	rt := &ResourceType{
		id:       "User",
		name:     "User",
		endpoint: "/v2/Users",
		schema:   &Schema{id: "test"},
	}

	raw, err := json.Marshal(rt)
	assert.Nil(s.T(), err)
	assert.JSONEq(s.T(), `{"id":"User","name":"User","description":"","endpoint":"/v2/Users","schema":"test"}`, string(raw))
}

func (s *ResourceTypeTestSuite) TestUnmarshal() {
	Schemas().Register(&Schema{id: "test"})
	Schemas().Register(&Schema{id: "ext1"})

	raw := `
{
  "id": "User",
  "name": "User",
  "description": "",
  "endpoint": "/v2/Users",
  "schema": "test",
  "schemaExtensions": [
    {
      "schema": "ext1",
      "required": true
    }
  ]
}
`

	rt := new(ResourceType)
	err := json.Unmarshal([]byte(raw), rt)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "User", rt.ID())
	assert.Equal(s.T(), "User", rt.Name())
	assert.Equal(s.T(), "/v2/Users", rt.Endpoint())
	assert.NotNil(s.T(), rt.Schema())
	assert.Len(s.T(), rt.extensions, 1)
}
