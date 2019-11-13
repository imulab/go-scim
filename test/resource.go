package test

import (
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

// Parse the schema from the relative filePath from this package. If save is true, cache
// the parsed schema into the core.Schemas cache.
func MustParseSchema(t *testing.T, filePath string, save bool) *core.Schema {
	raw, err := ioutil.ReadFile(filePath)
	require.Nil(t, err)

	schema, err := core.ParseSchema(raw)
	require.Nil(t, err)

	if save {
		core.Schemas.Add(schema)
	}

	return schema
}

// Parse the schema companion from the relative filePath from this package. If install is true, will try
// to install the schema companion onto the previously cache schema.
func MustParseSchemaCompanion(t *testing.T, filePath string, install bool) *core.SchemaCompanion {
	raw, err := ioutil.ReadFile(filePath)
	require.Nil(t, err)

	companion, err := core.ParseSchemaCompanion(raw)
	require.Nil(t, err)

	if install {
		companion.MustLoadOntoSchema()
	}

	return companion
}

// Parse the resource type from the relative filePath from this package.
func MustParseResourceType(t *testing.T, filePath string) *core.ResourceType {
	raw, err := ioutil.ReadFile(filePath)
	require.Nil(t, err)

	resourceType, err := core.ParseResourceType(raw)
	require.Nil(t, err)

	return resourceType
}