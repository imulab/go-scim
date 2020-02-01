package service

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/crud"
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

func TestQueryService(t *testing.T) {
	s := new(QueryServiceTestSuite)
	suite.Run(t, s)
}

type QueryServiceTestSuite struct {
	suite.Suite
	config       *spec.ServiceProviderConfig
	resourceType *spec.ResourceType
}

func (s *QueryServiceTestSuite) TestDo() {
	tests := []struct {
		name       string
		setup      func(t *testing.T) Query
		getRequest func() *QueryRequest
		expect     func(t *testing.T, resp *QueryResponse, err error)
	}{
		{
			name: "count",
			setup: func(t *testing.T) Query {
				database := db.Memory()
				for _, userData := range []interface{}{
					map[string]interface{}{"id": "user001"},
					map[string]interface{}{"id": "user002"},
					map[string]interface{}{"id": "user003"},
					map[string]interface{}{"id": "user004"},
					map[string]interface{}{"id": "user005"},
				} {
					require.Nil(t, database.Insert(context.TODO(), s.resourceOf(t, userData)))
				}
				return QueryService(s.config, database)
			},
			getRequest: func() *QueryRequest {
				return &QueryRequest{
					Filter: "id pr",
					Pagination: &crud.Pagination{
						StartIndex: 1,
						Count:      0,
					},
				}
			},
			expect: func(t *testing.T, resp *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 5, resp.TotalResults)
				assert.Empty(t, resp.Resources)
			},
		},
		{
			name: "filter",
			setup: func(t *testing.T) Query {
				database := db.Memory()
				for _, userData := range []interface{}{
					map[string]interface{}{"id": "user001", "userName": "user001"},
					map[string]interface{}{"id": "user002", "userName": "user002"},
					map[string]interface{}{"id": "user003", "userName": "user003"},
					map[string]interface{}{"id": "user004", "userName": "user004"},
					map[string]interface{}{"id": "user005", "userName": "user005"},
				} {
					require.Nil(t, database.Insert(context.TODO(), s.resourceOf(t, userData)))
				}
				return QueryService(s.config, database)
			},
			getRequest: func() *QueryRequest {
				return &QueryRequest{
					Filter: "userName ew \"003\"",
				}
			},
			expect: func(t *testing.T, resp *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 1, resp.TotalResults)
				assert.Len(t, resp.Resources, 1)
				assert.Equal(t, "user003", resp.Resources[0].(*prop.Resource).Navigator().Dot("id").Current().Raw())
			},
		},
		{
			name: "sort",
			setup: func(t *testing.T) Query {
				database := db.Memory()
				for _, userData := range []interface{}{
					map[string]interface{}{"id": "user003", "userName": "user003"},
					map[string]interface{}{"id": "user001", "userName": "user001"},
					map[string]interface{}{"id": "user005", "userName": "user005"},
					map[string]interface{}{"id": "user002", "userName": "user002"},
					map[string]interface{}{"id": "user004", "userName": "user004"},
				} {
					require.Nil(t, database.Insert(context.TODO(), s.resourceOf(t, userData)))
				}
				return QueryService(s.config, database)
			},
			getRequest: func() *QueryRequest {
				return &QueryRequest{
					Filter: "userName pr",
					Sort: &crud.Sort{
						By:    "userName",
						Order: crud.SortDesc,
					},
				}
			},
			expect: func(t *testing.T, resp *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 5, resp.TotalResults)
				assert.Len(t, resp.Resources, 5)
				for i, expected := range []string{"user005", "user004", "user003", "user002", "user001"} {
					assert.Equal(t, expected, resp.Resources[i].(*prop.Resource).Navigator().Dot("id").Current().Raw())
				}
			},
		},
		{
			name: "paginate",
			setup: func(t *testing.T) Query {
				database := db.Memory()
				for _, userData := range []interface{}{
					map[string]interface{}{"id": "user003", "userName": "user003"},
					map[string]interface{}{"id": "user001", "userName": "user001"},
					map[string]interface{}{"id": "user005", "userName": "user005"},
					map[string]interface{}{"id": "user002", "userName": "user002"},
					map[string]interface{}{"id": "user004", "userName": "user004"},
				} {
					require.Nil(t, database.Insert(context.TODO(), s.resourceOf(t, userData)))
				}
				return QueryService(s.config, database)
			},
			getRequest: func() *QueryRequest {
				return &QueryRequest{
					Filter: "userName pr",
					Sort: &crud.Sort{
						By:    "userName",
						Order: crud.SortAsc,
					},
					Pagination: &crud.Pagination{
						StartIndex: 2,
						Count:      2,
					},
				}
			},
			expect: func(t *testing.T, resp *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 5, resp.TotalResults)
				assert.Len(t, resp.Resources, 2)
				for i, expected := range []string{"user002", "user003"} {
					assert.Equal(t, expected, resp.Resources[i].(*prop.Resource).Navigator().Dot("id").Current().Raw())
				}
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

func (s *QueryServiceTestSuite) resourceOf(t *testing.T, data interface{}) *prop.Resource {
	r := prop.NewResource(s.resourceType)
	require.Nil(t, r.Navigator().Replace(data).Error())
	return r
}

func (s *QueryServiceTestSuite) SetupSuite() {
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

	s.config = new(spec.ServiceProviderConfig)
	require.Nil(s.T(), json.Unmarshal([]byte(`
{
  "filter": {
    "supported": true
  },
  "sort": {
    "supported": true
  }
}
`), s.config))
}
