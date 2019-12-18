package services

import (
	"context"
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/imulab/go-scim/protocol/crud"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestQueryService(t *testing.T) {
	s := new(QueryServiceTestSuite)
	s.resourceBase = "./internal/query_service_test_suite"
	suite.Run(t, s)
}

type QueryServiceTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *QueryServiceTestSuite) TestQuery() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	spc := s.mustServiceProviderConfig("/service_provider_config.json")

	tests := []struct {
		name       string
		getService func(t *testing.T) *QueryService
		request    *QueryRequest
		expect     func(t *testing.T, response *QueryResponse, err error)
	}{
		{
			name: "simple count",
			getService: func(t *testing.T) *QueryService {
				database := db.Memory()
				for _, f := range []string{
					"/user_003.json",
					"/user_002.json",
					"/user_001.json",
					"/user_004.json",
					"/user_005.json",
					"/user_009.json",
					"/user_010.json",
					"/user_008.json",
					"/user_007.json",
					"/user_006.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return &QueryService{
					Logger:                log.None(),
					Database:              database,
					ServiceProviderConfig: spc,
				}
			},
			request: &QueryRequest{
				Filter: "userName pr",
				Pagination: &crud.Pagination{
					StartIndex: 1,
					Count:      0,
				},
			},
			expect: func(t *testing.T, response *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 10, response.TotalResults)
				assert.Empty(t, response.Resources)
			},
		},
		{
			name: "filter",
			getService: func(t *testing.T) *QueryService {
				database := db.Memory()
				for _, f := range []string{
					"/user_003.json",
					"/user_002.json",
					"/user_001.json",
					"/user_004.json",
					"/user_005.json",
					"/user_009.json",
					"/user_010.json",
					"/user_008.json",
					"/user_007.json",
					"/user_006.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return &QueryService{
					Logger:                log.None(),
					Database:              database,
					ServiceProviderConfig: spc,
				}
			},
			request: &QueryRequest{
				Filter: "userName eq \"user001\"",
			},
			expect: func(t *testing.T, response *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 1, response.TotalResults)
				assert.Len(t, response.Resources, 1)
			},
		},
		{
			name: "sort",
			getService: func(t *testing.T) *QueryService {
				database := db.Memory()
				for _, f := range []string{
					"/user_003.json",
					"/user_002.json",
					"/user_001.json",
					"/user_004.json",
					"/user_005.json",
					"/user_009.json",
					"/user_010.json",
					"/user_008.json",
					"/user_007.json",
					"/user_006.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return &QueryService{
					Logger:           log.None(),
					Database:         database,
					ServiceProviderConfig: spc,
				}
			},
			request: &QueryRequest{
				Filter: "userName pr",
				Sort: &crud.Sort{
					By:    "userName",
					Order: crud.SortDesc,
				},
			},
			expect: func(t *testing.T, response *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 10, response.TotalResults)
				assert.Len(t, response.Resources, 10)
				userNames := make([]interface{}, 0)
				for _, r := range response.Resources {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "user010", userNames[0])
				assert.Equal(t, "user009", userNames[1])
				assert.Equal(t, "user008", userNames[2])
				assert.Equal(t, "user007", userNames[3])
				assert.Equal(t, "user006", userNames[4])
				assert.Equal(t, "user005", userNames[5])
				assert.Equal(t, "user004", userNames[6])
				assert.Equal(t, "user003", userNames[7])
				assert.Equal(t, "user002", userNames[8])
				assert.Equal(t, "user001", userNames[9])
			},
		},
		{
			name: "paginate",
			getService: func(t *testing.T) *QueryService {
				database := db.Memory()
				for _, f := range []string{
					"/user_003.json",
					"/user_002.json",
					"/user_001.json",
					"/user_004.json",
					"/user_005.json",
					"/user_009.json",
					"/user_010.json",
					"/user_008.json",
					"/user_007.json",
					"/user_006.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return &QueryService{
					Logger:           log.None(),
					Database:         database,
					ServiceProviderConfig: spc,
				}
			},
			request: &QueryRequest{
				Filter: "userName pr",
				Sort: &crud.Sort{
					By:    "userName",
					Order: crud.SortAsc,
				},
				Pagination: &crud.Pagination{
					StartIndex: 5,
					Count:      2,
				},
			},
			expect: func(t *testing.T, response *QueryResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 10, response.TotalResults)
				assert.Len(t, response.Resources, 2)
				userNames := make([]interface{}, 0)
				for _, r := range response.Resources {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "user005", userNames[0])
				assert.Equal(t, "user006", userNames[1])
			},
		},
		{
			name: "too many",
			getService: func(t *testing.T) *QueryService {
				database := db.Memory()
				for _, f := range []string{
					"/user_003.json",
					"/user_002.json",
					"/user_001.json",
					"/user_004.json",
					"/user_005.json",
					"/user_009.json",
					"/user_010.json",
					"/user_008.json",
					"/user_007.json",
					"/user_006.json",
				} {
					err := database.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				limitedSPC := s.mustServiceProviderConfig("/service_provider_config.json")
				limitedSPC.Filter.MaxResults = 5
				return &QueryService{
					Logger:           log.None(),
					Database:         database,
					ServiceProviderConfig: limitedSPC,
				}
			},
			request: &QueryRequest{
				Filter: "userName pr",
				Sort: &crud.Sort{
					By:    "userName",
					Order: crud.SortAsc,
				},
			},
			expect: func(t *testing.T, response *QueryResponse, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			service := test.getService(t)
			response, err := service.QueryResource(context.Background(), test.request)
			test.expect(t, response, err)
		})
	}
}

func (s *QueryServiceTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *QueryServiceTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *QueryServiceTestSuite) mustSchema(filePath string) *spec.Schema {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	sch := new(spec.Schema)
	err = json.Unmarshal(raw, sch)
	s.Require().Nil(err)

	spec.SchemaHub.Put(sch)

	return sch
}

func (s *QueryServiceTestSuite) mustServiceProviderConfig(filePath string) *spec.ServiceProviderConfig {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	spc := new(spec.ServiceProviderConfig)
	err = json.Unmarshal(raw, spc)
	s.Require().Nil(err)

	return spc
}
