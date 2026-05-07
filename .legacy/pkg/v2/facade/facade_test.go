package facade_test

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/facade"
	scimjson "github.com/imulab/go-scim/pkg/v2/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestFacade(t *testing.T) {
	suite.Run(t, new(facadeTestSuite))
}

type facadeTestSuite struct {
	suite.Suite
	rt *spec.ResourceType
}

func (s *facadeTestSuite) TestExport() {
	res, err := facade.Export(testUser, s.rt)
	assert.NoError(s.T(), err)

	raw, err := scimjson.Serialize(res)
	assert.NoError(s.T(), err)

	assert.JSONEq(s.T(), testJSON, string(raw))
}

func (s *facadeTestSuite) TestImport() {
	r := prop.NewResource(s.rt)
	err := scimjson.Deserialize([]byte(testJSON), r)
	require.NoError(s.T(), err)

	u := new(User)
	err = facade.Import(r, u)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), testUser.Id, u.Id)
	assert.Equal(s.T(), testUser.Email, u.Email)
	assert.Equal(s.T(), testUser.BackupEmail, u.BackupEmail)
	assert.Equal(s.T(), testUser.Name, u.Name)
	assert.Nil(s.T(), u.NickName)
	assert.Equal(s.T(), testUser.CreatedAt, u.CreatedAt)
	assert.Equal(s.T(), testUser.UpdatedAt, u.UpdatedAt)
	assert.Equal(s.T(), testUser.Active, u.Active)
	assert.Equal(s.T(), testUser.Manager, u.Manager)
}

func (s *facadeTestSuite) SetupSuite() {
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
				s.rt = parsed.(*spec.ResourceType)
			},
		},
	} {
		f, err := os.Open(each.filepath)
		require.NoError(s.T(), err)

		raw, err := ioutil.ReadAll(f)
		require.NoError(s.T(), err)

		err = json.Unmarshal(raw, each.structure)
		require.NoError(s.T(), err)

		if each.post != nil {
			each.post(each.structure)
		}
	}

	expr.RegisterURN("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
}

func ref(v string) *string {
	return &v
}

type User struct {
	Id          string  `scim:"id"`
	Email       string  `scim:"userName,emails[type eq \"work\" and primary eq true].value"`
	BackupEmail *string `scim:"emails[type eq \"work\" and primary eq false].value"`
	Name        string  `scim:"name.formatted"`
	NickName    *string `scim:"nickName"`
	CreatedAt   int64   `scim:"meta.created"`
	UpdatedAt   int64   `scim:"meta.lastModified"`
	Active      bool    `scim:"active"`
	Manager     *string `scim:"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value"`
}

var testUser = &User{
	Id:          "test",
	Email:       "john@gmail.com",
	BackupEmail: ref("john@outlook.com"),
	Name:        "John Doe",
	NickName:    nil,
	//CreatedAt:   1608795238,
	CreatedAt: 1608824038,
	UpdatedAt: 1608824038,
	//UpdatedAt: 1608795238,
	Active:  false,
	Manager: ref("tom"),
}

const testJSON = `
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User",
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
  ],
  "id": "test",
  "meta": {
    "resourceType": "User",
    "created": "2020-12-24T15:33:58",
    "lastModified": "2020-12-24T15:33:58"
  },
  "userName": "john@gmail.com",
  "name": {
    "formatted": "John Doe"
  },
  "active": false,
  "emails": [
    {
      "value": "john@gmail.com",
      "type": "work",
      "primary": true
    },
    {
      "value": "john@outlook.com",
      "type": "work",
      "primary": false
    }
  ],
  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
    "manager": {
      "value": "tom"
    }
  }
}
`
