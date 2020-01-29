package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetService(t *testing.T) {
	s := new(GetServiceTestSuite)
	suite.Run(t, s)
}

type GetServiceTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *GetServiceTestSuite) TestDo() {
	tests := []struct {
		name       string
		setup      func(t *testing.T) Get
		getRequest func() *GetRequest
		expect     func(t *testing.T, resp *GetResponse, err error)
	}{
		{
			name: "get existing",
			setup: func(t *testing.T) Get {
				database := db.Memory()
				err := database.Insert(context.TODO(), s.resourceOf(t, map[string]interface{}{
					"id": "foobar",
				}))
				require.Nil(t, err)
				return GetService(database)
			},
			getRequest: func() *GetRequest {
				return &GetRequest{
					ResourceID: "foobar",
				}
			},
			expect: func(t *testing.T, resp *GetResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", resp.Resource.Navigator().Dot("id").Current().Raw())
			},
		},
		{
			name: "get non-existing",
			setup: func(t *testing.T) Get {
				return GetService(db.Memory())
			},
			getRequest: func() *GetRequest {
				return &GetRequest{
					ResourceID: "foobar",
				}
			},
			expect: func(t *testing.T, resp *GetResponse, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrNotFound, errors.Unwrap(err))
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

func (s *GetServiceTestSuite) resourceOf(t *testing.T, data interface{}) *prop.Resource {
	r := prop.NewResource(s.resourceType)
	require.Nil(t, r.Navigator().Replace(data).Error())
	return r
}

func (s *GetServiceTestSuite) SetupSuite() {
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
