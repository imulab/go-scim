package scim

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewResourceType(t *testing.T) {
	type User struct {
		ID    string
		Name  string
		Email string
	}

	rt := NewResourceType[User]("User").
		Name("User").
		Location("https://example.org/scim", "/v2/Users").
		MainSchema(UserSchema).
		ExtendSchema(UserEnterpriseSchemaExtension, false).
		NewFunc(func() *User { return &User{} }).
		AddMapping(func(md *mappingDsl[User]) {
			md.Path("id").
				Getter(func(model *User) (any, error) {
					return model.ID, nil
				}).
				Setter(func(prop Property, model *User) error {
					return nil
				})
		}).
		AddMapping(func(md *mappingDsl[User]) {
			md.Path("name.formatted").
				Getter(func(model *User) (any, error) {
					if len(model.Name) == 0 {
						return nil, nil
					}
					return model.Name, nil
				}).
				Setter(func(prop Property, model *User) error {
					if !prop.Unassigned() {
						model.Name = prop.Value().(string)
					} else {
						model.Name = ""
					}
					return nil
				})
		}).
		AddMapping(func(md *mappingDsl[User]) {
			md.Path(`emails[type eq "work"].value`).
				Getter(func(model *User) (any, error) {
					if len(model.Email) == 0 {
						return nil, nil
					}
					return model.Email, nil
				}).
				Setter(func(prop Property, model *User) error {
					if !prop.Unassigned() {
						model.Email = prop.Value().(string)
					} else {
						model.Email = ""
					}
					return nil
				})
		}).
		Build()

	raw, err := json.Marshal(rt)
	if assert.NoError(t, err) {
		assert.JSONEq(t, `
{
  "id": "User",
  "name": "User",
  "endpoint": "/v2/Users",
  "schema": "urn:ietf:params:scim:schemas:core:2.0:User",
  "schemaExtensions": [
    {
      "schema": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
      "required": false
    }
  ]
}
`, string(raw))
	}
}
