package facade_test

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/facade"
	scimjson "github.com/imulab/go-scim/pkg/v2/json"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestToResource(t *testing.T) {
	var resourceType *spec.ResourceType
	{
		for _, each := range []struct {
			filepath  string
			structure interface{}
			post      func(parsed interface{})
		}{
			{
				filepath:  "../../../public/schemas/core_schema.json",
				structure: new(spec.Schema),
				post: func(parsed interface{}) {
					spec.Schemas().Register(parsed.(*spec.Schema))
				},
			},
			{
				filepath:  "../../../public/schemas/user_schema.json",
				structure: new(spec.Schema),
				post: func(parsed interface{}) {
					spec.Schemas().Register(parsed.(*spec.Schema))
				},
			},
			{
				filepath:  "../../../public/schemas/user_enterprise_extension_schema.json",
				structure: new(spec.Schema),
				post: func(parsed interface{}) {
					spec.Schemas().Register(parsed.(*spec.Schema))
				},
			},
			{
				filepath:  "../../../public/resource_types/user_resource_type.json",
				structure: new(spec.ResourceType),
				post: func(parsed interface{}) {
					resourceType = parsed.(*spec.ResourceType)
				},
			},
		} {
			f, err := os.Open(each.filepath)
			require.Nil(t, err)

			raw, err := ioutil.ReadAll(f)
			require.Nil(t, err)

			err = json.Unmarshal(raw, each.structure)
			require.Nil(t, err)

			if each.post != nil {
				each.post(each.structure)
			}
		}
	}

	expr.RegisterURN("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")

	u := &User{
		Id:          "testId",
		CreatedAt:   time.Now().Unix(),
		Email:       "foo@bar.com",
		BackupEmail: "foo2@bar.com",
		Name:        "foo",
		NickName:    func() *string { v := "foobar"; return &v }(),
		Manager:     "1003",
		Active:      true,
	}

	res, err := facade.ToResource(u, resourceType)
	assert.NoError(t, err)

	raw, err := scimjson.Serialize(res)
	assert.NoError(t, err)

	println(string(raw))
}

type User struct {
	Id          string  `scim:"id"`
	CreatedAt   int64   `scim:"meta.created"`
	Email       string  `scim:"emails[type eq \"work\" and primary eq true].value"`
	BackupEmail string  `scim:"emails[type eq \"home\"].value"`
	NickName    *string `scim:"nickName"`
	Name        string  `scim:"name.formatted"`
	Manager     string  `scim:"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value"`
	Active      bool    `scim:"active"`
}
