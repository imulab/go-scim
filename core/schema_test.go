package core

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestParseSchema(t *testing.T) {
	tests := []struct{
		name 		string
		filePath	string
		assertion	func(t *testing.T, schema *Schema, err error)
	}{
		{
			name: "parse user schema",
			filePath: "../resource/schema/user_schema.json",
			assertion: func(t *testing.T, schema *Schema, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", schema.Id)
				assert.Equal(t, 21, len(schema.Attributes))
			},
		},
		{
			name: "parse user enterprise extension schema",
			filePath: "../resource/schema/user_enterprise_schema.json",
			assertion: func(t *testing.T, schema *Schema, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User", schema.Id)
				assert.Equal(t, 6, len(schema.Attributes))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := ioutil.ReadFile(test.filePath)
			assert.Nil(t, err)

			schema, err := ParseSchema(raw)
			test.assertion(t, schema, err)
		})
	}
}