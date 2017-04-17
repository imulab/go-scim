package shared

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)
	require.NotNil(t, sch)

	for _, test := range []struct {
		getTarget  func() (interface{}, string)
		attributes []string
		excluded   []string
	}{
		{
			func() (interface{}, string) {
				r, _, err := ParseResource("../resources/tests/user_1.json")
				require.Nil(t, err)
				require.NotNil(t, r)

				_, json, err := ParseResource("../resources/tests/user_1_marshal.json")
				require.Nil(t, err)
				require.NotEmpty(t, json)

				return r, json
			},
			nil,
			nil,
		},
		{
			func() (interface{}, string) {
				r, _, err := ParseResource("../resources/tests/user_1.json")
				require.Nil(t, err)
				require.NotNil(t, r)
				return r, `{"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"], "id": "6B69753B-4E38-444E-8AC6-9D0E4D644D80"}`
			},
			[]string{"id"},
			nil,
		},
		{
			func() (interface{}, string) {
				r, _, err := ParseResource("../resources/tests/user_1.json")
				require.Nil(t, err)
				require.NotNil(t, r)
				return r, `{"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"], "id": "6B69753B-4E38-444E-8AC6-9D0E4D644D80", "userName": "david@example.com"}`
			},
			nil,
			[]string{
				"schemas",
				"externalId",
				"name",
				"displayName",
				"nickName",
				"profileUrl",
				"emails",
				"addresses",
				"phoneNumbers",
				"ims",
				"photos",
				"userType",
				"title",
				"preferredLanguage",
				"locale",
				"timezone",
				"active",
				"meta",
			},
		},
		{
			func() (interface{}, string) {
				r, _, err := ParseResource("../resources/tests/user_1.json")
				require.Nil(t, err)
				require.NotNil(t, r)

				listResponse := &ListResponse{
					Schemas:      []string{ListResponseUrn},
					StartIndex:   1,
					ItemsPerPage: 10,
					TotalResults: 100,
					Resources:    []DataProvider{r, r},
				}

				json := `
				{
					"schemas": ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
					"totalResults": 100,
					"itemsPerPage": 10,
					"startIndex": 1,
					"Resources": [
						{
							"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
							"id": "6B69753B-4E38-444E-8AC6-9D0E4D644D80",
							"userName": "david@example.com"
						},
						{
							"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
							"id": "6B69753B-4E38-444E-8AC6-9D0E4D644D80",
							"userName": "david@example.com"
						}
					]
				}
				`
				return listResponse, json
			},
			[]string{"id", "userName"},
			nil,
		},
	} {
		target, json := test.getTarget()
		b, err := MarshalJSON(target, sch, test.attributes, test.excluded)
		assert.Nil(t, err)
		assert.JSONEq(t, json, string(b))
	}
}
