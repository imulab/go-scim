package json

import (
	"testing"
	"github.com/davidiamyou/go-scim/resource"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	schema, _, err := resource.ParseSchema("user_schema.json")
	assert.Nil(t, err)

	av := resource.GetAttributeVault(schema)

	decoder := ScimDecoder{
		Factory: resource.DefaultFactory{},
		Attributes: av,
	}

	json := []byte(`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"], "id": "abc", "emails" : [{ "value":"foo@bar.com" }], "name" : { "familyName": "Qiu" }, "userName": "david", "active": true}`)
	v, err := decoder.Decode(json)

	t.Logf("%+v", v)
	t.Logf("%+v", err)
}
