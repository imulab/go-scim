package scratch_test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildResourceType(t *testing.T) {
	type User struct {
		Id        string
		FirstName string
		LastName  string
		Email     string
	}

	userRType := scim.BuildResourceType[User]("User").
		Name("User").
		Description("User resource type").
		Endpoint("/v2/User").
		SelfLocation("https://test.org/v2/ResourceType/User").
		MainSchema(scim.UserSchema).
		AddExtensionSchema(scim.UserEnterpriseSchemaExtension).
		NewFunc(func() *User { return new(User) }).
		AddMapping(
			scim.BuildMapping[User]("id").
				Getter(func(model *User) any {
					return model.Id
				}).
				Setter(func(prop scim.Property, model *User) {
					if !prop.Unassigned() {
						model.Id = prop.Value().(string)
					}
				}).
				EnableFilter().
				Build(),
			scim.BuildMapping[User]("name.givenName").
				Getter(func(model *User) any {
					return model.FirstName
				}).
				Setter(func(prop scim.Property, model *User) {
					if !prop.Unassigned() {
						model.FirstName = prop.Value().(string)
					}
				}).Build(),
			scim.BuildMapping[User]("name.familyName").
				Getter(func(model *User) any {
					return model.LastName
				}).
				Setter(func(prop scim.Property, model *User) {
					if !prop.Unassigned() {
						model.LastName = prop.Value().(string)
					}
				}).Build(),
			scim.BuildMapping[User](`emails[type eq "work"].value`).
				Getter(func(model *User) any {
					return model.Email
				}).
				Setter(func(prop scim.Property, model *User) {
					if !prop.Unassigned() {
						model.Email = prop.Value().(string)
					}
				}).Build(),
		).
		Build()

	userRTypeJSON, err := json.Marshal(userRType)
	if assert.NoError(t, err) {
		expectedJSON := `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:ResourceType"
  ],
  "meta": {
    "location": "https://test.org/v2/ResourceType/User",
    "resourceType": "ResourceType"
  },
  "id": "User",
  "name": "User",
  "endpoint": "/v2/User",
  "description": "User resource type",
  "schema": "urn:ietf:params:scim:schemas:core:2.0:User",
  "extensions": [
    {
      "schema": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
      "required": false
    }
  ]
}
`
		assert.JSONEq(t, expectedJSON, string(userRTypeJSON))
	}
}
