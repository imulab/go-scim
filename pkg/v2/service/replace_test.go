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

func TestReplaceService(t *testing.T) {
	s := new(ReplaceServiceTestSuite)
	suite.Run(t, s)
}

type ReplaceServiceTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *ReplaceServiceTestSuite) TestDo() {
	tests := []struct {
		name       string
		setup      func(t *testing.T) Replace
		getRequest func() *ReplaceRequest
		expect     func(t *testing.T, resp *ReplaceResponse, err error)
	}{
		{
			name: "replace with a valid resource",
			setup: func(t *testing.T) Replace {
				database := db.Memory()
				err := database.Insert(context.TODO(), s.resourceOf(t, map[string]interface{}{
					"schemas":  []interface{}{"urn:ietf:params:scim:schemas:core:2.0:User"},
					"id":       "foo",
					"userName": "foo",
				}))
				require.Nil(t, err)
				return ReplaceService(&spec.ServiceProviderConfig{}, s.resourceType, database, []filter.ByResource{
					filter.ByPropertyToByResource(
						filter.ReadOnlyFilter(),
						filter.BCryptFilter(),
					),
					filter.ByPropertyToByResource(filter.ValidationFilter(database)),
					filter.MetaFilter(),
				})
			},
			getRequest: func() *ReplaceRequest {
				return &ReplaceRequest{
					ResourceID: "foo",
					PayloadSource: strings.NewReader(`
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "foo",
  "userName": "bar",
  "emails": [
    {
      "value": "foo@bar.com"
    }
  ]
}
`),
				}
			},
			expect: func(t *testing.T, resp *ReplaceResponse, err error) {
				assert.Nil(t, err)
				assert.True(t, resp.Replaced)
				assert.Equal(t, "bar", resp.Resource.Navigator().Dot("userName").Current().Raw())
			},
		},
		{
			name: "replace with an identical resource",
			setup: func(t *testing.T) Replace {
				database := db.Memory()
				err := database.Insert(context.TODO(), s.resourceOf(t, map[string]interface{}{
					"schemas":  []interface{}{"urn:ietf:params:scim:schemas:core:2.0:User"},
					"id":       "foo",
					"userName": "foo",
					"emails": []interface{}{
						map[string]interface{}{
							"value": "foo@bar.com",
						},
					},
				}))
				require.Nil(t, err)
				return ReplaceService(&spec.ServiceProviderConfig{}, s.resourceType, database, []filter.ByResource{
					filter.ByPropertyToByResource(
						filter.ReadOnlyFilter(),
						filter.BCryptFilter(),
					),
					filter.ByPropertyToByResource(filter.ValidationFilter(database)),
					filter.MetaFilter(),
				})
			},
			getRequest: func() *ReplaceRequest {
				return &ReplaceRequest{
					ResourceID: "foo",
					PayloadSource: strings.NewReader(`
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "foo",
  "userName": "foo",
  "emails": [
    {
      "value": "foo@bar.com"
    }
  ]
}
`),
				}
			},
			expect: func(t *testing.T, resp *ReplaceResponse, err error) {
				assert.Nil(t, err)
				assert.False(t, resp.Replaced)
			},
		},
		{
			name: "replace with an invalid resource",
			setup: func(t *testing.T) Replace {
				database := db.Memory()
				err := database.Insert(context.TODO(), s.resourceOf(t, map[string]interface{}{
					"schemas":  []interface{}{"urn:ietf:params:scim:schemas:core:2.0:User"},
					"id":       "foo",
					"userName": "foo",
					"emails": []interface{}{
						map[string]interface{}{
							"value": "foo@bar.com",
						},
					},
				}))
				require.Nil(t, err)
				return ReplaceService(&spec.ServiceProviderConfig{}, s.resourceType, database, []filter.ByResource{
					filter.ByPropertyToByResource(
						filter.ReadOnlyFilter(),
						filter.BCryptFilter(),
					),
					filter.ByPropertyToByResource(filter.ValidationFilter(database)),
					filter.MetaFilter(),
				})
			},
			getRequest: func() *ReplaceRequest {
				return &ReplaceRequest{
					ResourceID: "foo",
					PayloadSource: strings.NewReader(`
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "emails": [
    {
      "value": "foo@bar.com"
    }
  ]
}
`),
				}
			},
			expect: func(t *testing.T, resp *ReplaceResponse, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidValue, errors.Unwrap(err))
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.setup(t)
			resp, err := service.Do(context.TODO(), test.getRequest())
			test.expect(t, resp, err)
		})
	}
}

func (s *ReplaceServiceTestSuite) resourceOf(t *testing.T, data interface{}) *prop.Resource {
	r := prop.NewResource(s.resourceType)
	require.Nil(t, r.Navigator().Replace(data).Error())
	return r
}

func (s *ReplaceServiceTestSuite) SetupSuite() {
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
