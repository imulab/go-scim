package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/service/filter"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestCreateService(t *testing.T) {
	s := new(CreateServiceTestSuite)
	suite.Run(t, s)
}

type CreateServiceTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *CreateServiceTestSuite) TestDo() {
	defaultSetup := func(t *testing.T) Create {
		memoryDB := db.Memory()
		return CreateService(s.resourceType, memoryDB, []filter.ByResource{
			filter.ByPropertyToByResource(
				filter.ReadOnlyFilter(),
				filter.UUIDFilter(),
				filter.BCryptFilter(),
			),
			filter.MetaFilter(),
			filter.ByPropertyToByResource(filter.ValidationFilter(memoryDB)),
		})
	}

	tests := []struct {
		name       string
		setup      func(t *testing.T) Create
		getRequest func() *CreateRequest
		expect     func(t *testing.T, resp *CreateResponse, err error)
	}{
		{
			name:  "create a new user",
			setup: defaultSetup,
			getRequest: func() *CreateRequest {
				return &CreateRequest{
					PayloadSource: strings.NewReader(`
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "userName": "foo",
  "emails": [
    {
      "value": "foo@bar.com",
      "primary": true
    },
    {
      "value": "bar@foo.com"
    }
  ]
}
`),
				}
			},
			expect: func(t *testing.T, resp *CreateResponse, err error) {
				n := func() prop.Navigator { return resp.Resource.Navigator() }
				assert.Nil(t, err)
				assert.Equal(t, []interface{}{"urn:ietf:params:scim:schemas:core:2.0:User"}, n().Dot("schemas").Current().Raw())
				assert.Equal(t, "foo", n().Dot("userName").Current().Raw())
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo@bar.com",
						"primary": true,
						"display": nil,
						"type":    nil,
					},
					map[string]interface{}{
						"value":   "bar@foo.com",
						"primary": nil,
						"display": nil,
						"type":    nil,
					},
				}, n().Dot("emails").Current().Raw())
				assert.NotEmpty(t, n().Dot("id").Current().Raw())
				assert.Equal(t, "User", n().Dot("meta").Dot("resourceType").Current().Raw())
				assert.NotEmpty(t, n().Dot("meta").Dot("created").Current().Raw())
				assert.NotEmpty(t, n().Dot("meta").Dot("lastModified").Current().Raw())
				assert.NotEmpty(t, n().Dot("meta").Dot("location").Current().Raw())
				assert.NotEmpty(t, n().Dot("meta").Dot("version").Current().Raw())
			},
		},
		{
			name:  "create a new user with missing userName",
			setup: defaultSetup,
			getRequest: func() *CreateRequest {
				return &CreateRequest{
					PayloadSource: strings.NewReader(`
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "emails": [
    {
      "value": "foo@bar.com",
      "primary": true
    },
    {
      "value": "bar@foo.com"
    }
  ]
}
`),
				}
			},
			expect: func(t *testing.T, resp *CreateResponse, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidValue, errors.Unwrap(err))
			},
		},
		{
			name:  "self administered readOnly fields are ignored",
			setup: defaultSetup,
			getRequest: func() *CreateRequest {
				return &CreateRequest{
					PayloadSource: strings.NewReader(`
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "foobar",
  "userName": "foo",
  "emails": [
    {
      "value": "foo@bar.com",
      "primary": true
    },
    {
      "value": "bar@foo.com"
    }
  ]
}
`),
				}
			},
			expect: func(t *testing.T, resp *CreateResponse, err error) {
				assert.Nil(t, err)
				assert.NotEqual(t, "foobar", resp.Resource.Navigator().Dot("id").Current().Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			resp, err := service.Do(context.Background(), test.getRequest())
			test.expect(t, resp, err)
		})
	}
}

func (s *CreateServiceTestSuite) SetupSuite() {
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
			filepath:  "../../../public/resource_types/user_resource_type.json",
			structure: new(spec.ResourceType),
			post: func(parsed interface{}) {
				s.resourceType = parsed.(*spec.ResourceType)
			},
		},
	} {
		f, err := os.Open(each.filepath)
		require.Nil(s.T(), err)

		raw, err := ioutil.ReadAll(f)
		require.Nil(s.T(), err)

		err = json.Unmarshal(raw, each.structure)
		require.Nil(s.T(), err)

		if each.post != nil {
			each.post(each.structure)
		}
	}
}
