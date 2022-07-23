package scim_test

import (
	"encoding/json"
	"github.com/imulab/go-scim"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserSchema(t *testing.T) {
	userSchemaJSON, err := json.Marshal(scim.UserSchema)
	if assert.NoError(t, err) {
		expectedJSON := `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "Defined attributes for the user schema",
  "attributes": [
    {
      "name": "userName",
      "type": "string",
      "multiValued": false,
      "required": true,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "server"
    },
    {
      "name": "name",
      "type": "complex",
      "subAttributes": [
        {
          "name": "formatted",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "familyName",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "givenName",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "middleName",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "honorificPrefix",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "honorificSuffix",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "displayName",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "nickName",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "profileUrl",
      "type": "reference",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "referenceTypes": [
        "external"
      ]
    },
    {
      "name": "title",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "userType",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "preferredLanguage",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "locale",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "timezone",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "active",
      "type": "boolean",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "password",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "writeOnly",
      "returned": "never"
    },
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
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": true,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "phoneNumbers",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "ims",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "photos",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "reference",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default",
          "referenceTypes": [
            "external"
          ]
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "addresses",
      "type": "complex",
      "subAttributes": [
        {
          "name": "formatted",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "streetAddress",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "locality",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "region",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "postalCode",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "country",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "groups",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readOnly",
          "returned": "default"
        },
        {
          "name": "$ref",
          "type": "reference",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readOnly",
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "canonicalValues": [
            "direct",
            "indirect"
          ],
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readOnly",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readOnly",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readOnly",
      "returned": "default"
    },
    {
      "name": "entitlements",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "roles",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "x509Certificates",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "type",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "primary",
          "type": "boolean",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    }
  ]
}
`
		assert.JSONEq(t, expectedJSON, string(userSchemaJSON))
	}
}

func TestUserEnterpriseExtensionSchema(t *testing.T) {
	schemaJSON, err := json.Marshal(scim.UserEnterpriseSchemaExtension)
	if assert.NoError(t, err) {
		expectedJSON := `
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "Enterprise User",
  "description": "Extension attributes for enterprise users",
  "attributes": [
    {
      "name": "employeeNumber",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "costCenter",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "organization",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "division",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "department",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "manager",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "$ref",
          "type": "reference",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        },
        {
          "name": "displayName",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    }
  ]
}
`
		assert.JSONEq(t, expectedJSON, string(schemaJSON))
	}
}

func TestGroupSchema(t *testing.T) {
	groupSchemaJSON, err := json.Marshal(scim.GroupSchema)
	if assert.NoError(t, err) {
		expectedJSON := `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:Group",
  "name": "Group",
  "description": "Defined attributes for the group schema",
  "attributes": [
    {
      "name": "displayName",
      "type": "string",
      "multiValued": false,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    },
    {
      "name": "members",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "immutable",
          "returned": "default"
        },
        {
          "name": "$ref",
          "type": "reference",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "immutable",
          "returned": "default"
        },
        {
          "name": "display",
          "type": "string",
          "multiValued": false,
          "required": false,
          "caseExact": false,
          "mutability": "readWrite",
          "returned": "default"
        }
      ],
      "multiValued": true,
      "required": false,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default"
    }
  ]
}
`
		assert.JSONEq(t, expectedJSON, string(groupSchemaJSON))
	}
}
