package db

import (
	"context"
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/crud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestMemoryDB(t *testing.T) {
	s := new(MemoryDBTestSuite)
	s.resourceBase = "../../tests/memory_db_test_suite"
	suite.Run(t, s)
}

type MemoryDBTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *MemoryDBTestSuite) TestInsert() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name         string
		getResources func() []*prop.Resource
		expectCount  int
	}{
		{
			name: "insert 10 users",
			getResources: func() []*prop.Resource {
				return []*prop.Resource{
					s.mustResource("/user_001.json", resourceType),
					s.mustResource("/user_002.json", resourceType),
					s.mustResource("/user_003.json", resourceType),
					s.mustResource("/user_004.json", resourceType),
					s.mustResource("/user_005.json", resourceType),
					s.mustResource("/user_006.json", resourceType),
					s.mustResource("/user_007.json", resourceType),
					s.mustResource("/user_008.json", resourceType),
					s.mustResource("/user_009.json", resourceType),
					s.mustResource("/user_010.json", resourceType),
				}
			},
			expectCount: 10,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			db := Memory()
			for _, r := range test.getResources() {
				err := db.Insert(context.Background(), r)
				assert.Nil(t, err)
			}
			count, err := db.Count(context.Background(), "")
			assert.Nil(t, err)
			assert.Equal(t, test.expectCount, count)
		})
	}
}

func (s *MemoryDBTestSuite) TestGet() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name   string
		getDB  func(t *testing.T) DB
		id     string
		expect func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "find existing",
			getDB: func(t *testing.T) DB {
				db := Memory()
				err := db.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return db
			},
			id: "a5866759-32ca-4e2a-9808-a0fe74f94b18",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, r)
				assert.Equal(t, "a5866759-32ca-4e2a-9808-a0fe74f94b18", r.ID())
			},
		},
		{
			name: "find non-existing",
			getDB: func(t *testing.T) DB {
				return Memory()
			},
			id: "a5866759-32ca-4e2a-9808-a0fe74f94b18",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			db := test.getDB(t)
			r, err := db.Get(context.Background(), test.id, nil)
			test.expect(t, r, err)
		})
	}
}

func (s *MemoryDBTestSuite) TestCount() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name   string
		getDB  func(t *testing.T) DB
		filter string
		expect func(t *testing.T, count int, err error)
	}{
		{
			name: "find all",
			getDB: func(t *testing.T) DB {
				db := Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := db.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return db
			},
			filter: "",
			expect: func(t *testing.T, count int, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 10, count)
			},
		},
		{
			name: "find by username",
			getDB: func(t *testing.T) DB {
				db := Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := db.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return db
			},
			filter: "userName eq \"user003\"",
			expect: func(t *testing.T, count int, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "find by username (multiple)",
			getDB: func(t *testing.T) DB {
				db := Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := db.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return db
			},
			filter: "userName eq \"user003\" or userName eq \"user004\"",
			expect: func(t *testing.T, count int, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 2, count)
			},
		},
		{
			name: "find by predicate within multiValued",
			getDB: func(t *testing.T) DB {
				db := Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := db.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return db
			},
			filter: "emails.value eq \"imulab@foo.com\"",
			expect: func(t *testing.T, count int, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 10, count)
			},
		},
		{
			name: "find by predicate within multiValued 2",
			getDB: func(t *testing.T) DB {
				db := Memory()
				for _, f := range []string{
					"/user_001.json",
					"/user_002.json",
					"/user_003.json",
					"/user_004.json",
					"/user_005.json",
					"/user_006.json",
					"/user_007.json",
					"/user_008.json",
					"/user_009.json",
					"/user_010.json",
				} {
					err := db.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return db
			},
			filter: "emails.value eq \"foobar\"",
			expect: func(t *testing.T, count int, err error) {
				assert.Nil(t, err)
				assert.Equal(t, 0, count)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			db := test.getDB(t)
			c, err := db.Count(context.Background(), test.filter)
			test.expect(t, c, err)
		})
	}
}

func (s *MemoryDBTestSuite) TestReplace() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name           string
		getDB          func(t *testing.T) DB
		getReplacement func(t *testing.T) *prop.Resource
		expect         func(t *testing.T, db DB, err error)
	}{
		{
			name: "replacement existing",
			getDB: func(t *testing.T) DB {
				db := Memory()
				err := db.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return db
			},
			getReplacement: func(t *testing.T) *prop.Resource {
				r := s.mustResource("/user_001.json", resourceType)
				p, err := r.NewNavigator().FocusName("userName")
				require.Nil(t, err)
				err = p.Replace("foobar")
				require.Nil(t, err)
				return r
			},
			expect: func(t *testing.T, db DB, err error) {
				assert.Nil(t, err)
				r, err := db.Get(context.Background(), "a5866759-32ca-4e2a-9808-a0fe74f94b18", nil)
				assert.Nil(t, err)
				p, err := r.NewNavigator().FocusName("userName")
				assert.Nil(t, err)
				assert.Equal(t, "foobar", p.Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			db := test.getDB(t)
			replacement := test.getReplacement(t)
			err := db.Replace(context.Background(), replacement)
			test.expect(t, db, err)
		})
	}
}

func (s *MemoryDBTestSuite) TestDelete() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name   string
		getDB  func(t *testing.T) DB
		id     string
		expect func(t *testing.T, db DB, err error)
	}{
		{
			name: "delete existing",
			getDB: func(t *testing.T) DB {
				db := Memory()
				err := db.Insert(context.Background(), s.mustResource("/user_001.json", resourceType))
				require.Nil(t, err)
				return db
			},
			id: "a5866759-32ca-4e2a-9808-a0fe74f94b18",
			expect: func(t *testing.T, db DB, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			db := test.getDB(t)
			err := db.Delete(context.Background(), test.id)
			test.expect(t, db, err)
		})
	}
}

func (s *MemoryDBTestSuite) TestQuery() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name       string
		getDB      func(t *testing.T) DB
		filter     string
		sort       *crud.Sort
		pagination *crud.Pagination
		expect     func(t *testing.T, results []*prop.Resource, err error)
	}{
		{
			name: "filter and sort",
			getDB: func(t *testing.T) DB {
				db := Memory()
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
					err := db.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return db
			},
			filter: "userName pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 10)

				userNames := make([]interface{}, 0)
				for _, r := range results {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "user001", userNames[0])
				assert.Equal(t, "user002", userNames[1])
				assert.Equal(t, "user003", userNames[2])
				assert.Equal(t, "user004", userNames[3])
				assert.Equal(t, "user005", userNames[4])
				assert.Equal(t, "user006", userNames[5])
				assert.Equal(t, "user007", userNames[6])
				assert.Equal(t, "user008", userNames[7])
				assert.Equal(t, "user009", userNames[8])
				assert.Equal(t, "user010", userNames[9])
			},
		},
		{
			name: "filter, sort and paginate",
			getDB: func(t *testing.T) DB {
				db := Memory()
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
					err := db.Insert(context.Background(), s.mustResource(f, resourceType))
					require.Nil(t, err)
				}
				return db
			},
			filter: "userName pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: &crud.Pagination{
				StartIndex: 5,
				Count:      2,
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)

				userNames := make([]interface{}, 0)
				for _, r := range results {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "user005", userNames[0])
				assert.Equal(t, "user006", userNames[1])
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			db := test.getDB(t)
			results, err := db.Query(context.Background(), test.filter, test.sort, test.pagination, nil)
			test.expect(t, results, err)
		})
	}
}

func (s *MemoryDBTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *MemoryDBTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *MemoryDBTestSuite) mustSchema(filePath string) *spec.Schema {
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
