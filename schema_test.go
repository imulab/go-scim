package scim

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserSchema(t *testing.T) {
	j, err := json.Marshal(BuildSchema(UserSchema))
	if assert.NoError(t, err) {
		assert.JSONEq(t, `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "Defined attributes for the user schema",
  "attributes": [
    {
      "name": "userName",
      "type": "string",
      "required": true,
      "uniqueness": "server"
    },
    {
      "name": "name",
      "type": "complex",
      "subAttributes": [
        {
          "name": "formatted",
          "type": "string"
        },
        {
          "name": "familyName",
          "type": "string"
        },
        {
          "name": "givenName",
          "type": "string"
        },
        {
          "name": "middleName",
          "type": "string"
        },
        {
          "name": "honorificPrefix",
          "type": "string"
        },
        {
          "name": "honorificSuffix",
          "type": "string"
        }
      ]
    },
    {
      "name": "displayName",
      "type": "string"
    },
    {
      "name": "nickName",
      "type": "string"
    },
    {
      "name": "profileUrl",
      "type": "reference",
      "referenceTypes": [
        "external"
      ]
    },
    {
      "name": "title",
      "type": "string"
    },
    {
      "name": "userType",
      "type": "string"
    },
    {
      "name": "preferredLanguage",
      "type": "string"
    },
    {
      "name": "locale",
      "type": "string"
    },
    {
      "name": "timezone",
      "type": "string"
    },
    {
      "name": "active",
      "type": "boolean"
    },
    {
      "name": "password",
      "type": "string",
      "mutability": "writeOnly",
      "returned": "never"
    },
    {
      "name": "emails",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string"
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        },
        {
          "name": "display",
          "type": "string"
        }
      ],
      "multiValued": true,
      "required": true
    },
    {
      "name": "phoneNumbers",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string"
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        },
        {
          "name": "display",
          "type": "string"
        }
      ],
      "multiValued": true
    },
    {
      "name": "ims",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string"
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        },
        {
          "name": "display",
          "type": "string"
        }
      ],
      "multiValued": true
    },
    {
      "name": "photos",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "reference",
          "referenceTypes": [
            "external"
          ]
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        }
      ],
      "multiValued": true
    },
    {
      "name": "addresses",
      "type": "complex",
      "subAttributes": [
        {
          "name": "formatted",
          "type": "string"
        },
        {
          "name": "streetAddress",
          "type": "string"
        },
        {
          "name": "locality",
          "type": "string"
        },
        {
          "name": "region",
          "type": "string"
        },
        {
          "name": "postalCode",
          "type": "string"
        },
        {
          "name": "country",
          "type": "string"
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        }
      ],
      "multiValued": true
    },
    {
      "name": "groups",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "mutability": "readOnly"
        },
        {
          "name": "$ref",
          "type": "reference",
          "mutability": "readOnly"
        },
        {
          "name": "type",
          "type": "string",
          "mutability": "readOnly"
        },
        {
          "name": "display",
          "type": "string",
          "mutability": "readOnly"
        }
      ],
      "multiValued": true,
      "mutability": "readOnly"
    },
    {
      "name": "entitlements",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string"
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        },
        {
          "name": "display",
          "type": "string"
        }
      ],
      "multiValued": true
    },
    {
      "name": "roles",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string"
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        },
        {
          "name": "display",
          "type": "string"
        }
      ],
      "multiValued": true
    },
    {
      "name": "x509Certificates",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "binary"
        },
        {
          "name": "type",
          "type": "string"
        },
        {
          "name": "primary",
          "type": "boolean"
        },
        {
          "name": "display",
          "type": "string"
        }
      ],
      "multiValued": true
    }
  ]
}
`, string(j))
	}
}

func TestUserEnterpriseSchemaExtension(t *testing.T) {
	j, err := json.Marshal(BuildSchema(UserEnterpriseSchemaExtension))
	if assert.NoError(t, err) {
		assert.JSONEq(t, `
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "Enterprise User",
  "description": "Extension attributes for enterprise users",
  "attributes": [
    {
      "name": "employeeNumber",
      "type": "string"
    },
    {
      "name": "costCenter",
      "type": "string"
    },
    {
      "name": "organization",
      "type": "string"
    },
    {
      "name": "division",
      "type": "string"
    },
    {
      "name": "department",
      "type": "string"
    },
    {
      "name": "manager",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string"
        },
        {
          "name": "$ref",
          "type": "reference"
        },
        {
          "name": "displayName",
          "type": "string"
        }
      ]
    }
  ]
}
`, string(j))
	}
}

func TestGroupSchema(t *testing.T) {
	j, err := json.Marshal(BuildSchema(GroupSchema))
	if assert.NoError(t, err) {
		assert.JSONEq(t, `
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:Group",
  "name": "Group",
  "description": "Defined attributes for the group schema",
  "attributes": [
    {
      "name": "displayName",
      "type": "string"
    },
    {
      "name": "members",
      "type": "complex",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "mutability": "immutable"
        },
        {
          "name": "$ref",
          "type": "reference",
          "mutability": "immutable"
        },
        {
          "name": "display",
          "type": "string",
          "mutability": "immutable"
        }
      ],
      "multiValued": true
    }
  ]
}
`, string(j))
	}
}
