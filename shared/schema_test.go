package shared

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSchema(t *testing.T) {
	sch, json, err := ParseSchema("../resources/tests/user_schema.json")
	assert.NotNil(t, sch)
	assert.NotEmpty(t, json)
	assert.Nil(t, err)
}

func TestSchema_GetAttribute(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.NotNil(t, sch)
	require.Nil(t, err)

	for _, test := range []struct {
		pathText  string
		assertion func(attr *Attribute)
	}{
		{
			"schemas",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "schemas", attr.Assist.FullPath)
			},
		},
		{
			"ID",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "id", attr.Assist.FullPath)
			},
		},
		{
			"meta.Created",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "meta.created", attr.Assist.FullPath)
			},
		},
		{
			"Name.familyName",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName", attr.Assist.FullPath)
			},
		},
		{
			"urn:ietf:params:scim:schemas:core:2.0:User:emails",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:emails", attr.Assist.FullPath)
			},
		},
		{
			"urn:ietf:params:scim:schemas:core:2.0:User:groups.value",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:groups.value", attr.Assist.FullPath)
			},
		},
		{
			"groups[type eq \"direct\"].value",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:groups.value", attr.Assist.FullPath)
			},
		},
	} {
		p, err := NewPath(test.pathText)
		require.Nil(t, err)
		test.assertion(sch.GetAttribute(p, true))
	}
}
