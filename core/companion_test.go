package core

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestParseSchemaCompanion(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		assertion func(t *testing.T, sc *SchemaCompanion, err error)
	}{
		{
			name:     "parse user schema companion",
			filePath: "../resource/companion/user_schema_companion.json",
			assertion: func(t *testing.T, sc *SchemaCompanion, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", sc.Schema)
				assert.Equal(t, 66, len(sc.Metadata))
			},
		},
		{
			name:     "parse user enterprise extension schema companion",
			filePath: "../resource/companion/user_enterprise_schema_companion.json",
			assertion: func(t *testing.T, sc *SchemaCompanion, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User", sc.Schema)
				assert.Equal(t, 9, len(sc.Metadata))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := ioutil.ReadFile(test.filePath)
			assert.Nil(t, err)

			sc, err := ParseSchemaCompanion(raw)
			test.assertion(t, sc, err)
		})
	}
}
