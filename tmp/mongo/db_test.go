package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	scimJSON "github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/imulab/go-scim/protocol/crud"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
)

var (
	testDockerEndpoint    = envOrDefault("TEST_DOCKER_ENDPOINT", "")
	testMongoImageName    = envOrDefault("TEST_DOCKER_MONGO_IMAGE", "bitnami/mongodb")
	testMongoImageTag     = envOrDefault("TEST_DOCKER_MONGO_TAG", "latest")
	testMongoUserName     = envOrDefault("TEST_DOCKER_MONGO_USERNAME", "testUser")
	testMongoUserSecret   = envOrDefault("TEST_DOCKER_MONGO_SECRET", "s3cret")
	testMongoDatabaseName = envOrDefault("TEST_DOCKER_MONGO_DB_NAME", "mongo_database_test_suite")
)

func TestMongoDatabase(t *testing.T) {
	s := new(MongoDatabaseTestSuite)
	s.resourceBase = "./internal/mongo_database_test_suite"
	suite.Run(t, s)
}

type MongoDatabaseTestSuite struct {
	suite.Suite
	resourceBase   string
	dockerPool     *dockertest.Pool
	dockerResource *dockertest.Resource
	newClient      func() (*mongo.Client, error)
}

func (s *MongoDatabaseTestSuite) TestQuery() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name        string
		description string
		prepare     func(t *testing.T, database db.DB)
		filter      string
		sort        *crud.Sort
		pagination  *crud.Pagination
		projection  *crud.Projection
		expect      func(t *testing.T, results []*prop.Resource, err error)
	}{
		{
			name: "schemas eq",
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/schemas_eq/user_001.json",
					"/test_query/schemas_eq/user_002.json",
					"/test_query/schemas_eq/user_003.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: fmt.Sprintf("schemas eq %s", strconv.Quote("urn:ietf:params:scim:schemas:core:2.0:User")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, "user001", results[0].NewFluentNavigator().FocusName("userName").Current().Raw())
				assert.Equal(t, "user002", results[1].NewFluentNavigator().FocusName("userName").Current().Raw())
			},
		},
		{
			name: "userName eq",
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/username_eq/user_001.json",
					"/test_query/username_eq/user_002.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: fmt.Sprintf("userName eq %s", strconv.Quote("user001")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "user001", results[0].NewFluentNavigator().FocusName("userName").Current().Raw())
			},
		},
		{
			name: "nested eq",
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/nested_eq/user_001.json",
					"/test_query/nested_eq/user_002.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: fmt.Sprintf("name.familyName eq %s", strconv.Quote("Q")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "userName", "name.familyName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "Q",
					results[0].NewFluentNavigator().FocusName("name").FocusName("familyName").Current().Raw())
			},
		},
		{
			name: "multiValue nested eq",
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/multi_nested_eq/user_001.json",
					"/test_query/multi_nested_eq/user_002.json",
					"/test_query/multi_nested_eq/user_003.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: fmt.Sprintf("emails.value eq %s", strconv.Quote("foo@bar.com")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "userName", "emails"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, "user001", results[0].NewFluentNavigator().FocusName("userName").Current().Raw())
				assert.Equal(t, "user003", results[1].NewFluentNavigator().FocusName("userName").Current().Raw())
			},
		},
		{
			name: "nickName pr",
			description: `
Tests the presence of a top level field: "nickName".
----------------------------------------------------
Two resources are inserted into database first:
- user_001.json has nickName "foo"
- user_002.json has no nickName
----------------------------------------------------
The filter "nickName pr" expects to return only user_001.json in the result.
`,
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/nickname_pr/user_001.json",
					"/test_query/nickname_pr/user_002.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: "nickName pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "nickName", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "user001", results[0].NewFluentNavigator().FocusName("userName").Current().Raw())
			},
		},
		{
			name: "nested pr",
			description: `
Tests the presence of a nested field: name.familyName.
------------------------------------------------------
Two resources were inserted into database:
- user_001.json has no name.familyName
- user_002.json has name.familyName "Qiu"
------------------------------------------------------
The filter "name.familyName pr" expects only user_002.json to be returned.
`,
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/nested_pr/user_001.json",
					"/test_query/nested_pr/user_002.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: "name.familyName pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "name.familyName", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "user002", results[0].NewFluentNavigator().FocusName("userName").Current().Raw())
			},
		},
		{
			name: "multiValued nested pr",
			description: `
Tests the presence of a nested field within a multiValued property: emails.display
----------------------------------------------------------------------------------
Three resources are inserted into database:
- user_001.json has emails.display set on one of the two emails
- user_002.json has no emails.display set on any emails
- user_003.json has emails.display set of both of the two emails
----------------------------------------------------------------------------------
The filter "emails.display pr" expects user_001.json and user_003.json be returned in the result.
`,
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/multi_nested_pr/user_001.json",
					"/test_query/multi_nested_pr/user_002.json",
					"/test_query/multi_nested_pr/user_003.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: "emails.display pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "name.familyName", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, "user001", results[0].NewFluentNavigator().FocusName("userName").Current().Raw())
				assert.Equal(t, "user003", results[1].NewFluentNavigator().FocusName("userName").Current().Raw())
			},
		},
		{
			name: "logical and",
			description: `
Tests the effect of the logical and operator, together with the less tested sw and gt operator.
-----------------------------------------------------------------------------------------------
Five resources are inserted into to the database:
- user_001.json has userName "user001" and meta.created "2019-12-21T09:00:00"
- user_002.json has userName "user002" and meta.created "2019-12-21T10:30:00"
- user_003.json has userName "user003" and meta.created "2019-12-21T10:00:00"
- user_004.json has userName "foobar" and meta.created "2019-12-21T10:30:00"
- user_005.json has userName "user005" and meta.created "2019-12-21T11:00:00"
-----------------------------------------------------------------------------------------------
The filter (userName sw "user") and (meta.created gt "2019-12-21T10:00:00") expects to return
user_002.json and user_005.json in results
`,
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					"/test_query/logical_and/user_001.json",
					"/test_query/logical_and/user_002.json",
					"/test_query/logical_and/user_003.json",
					"/test_query/logical_and/user_004.json",
					"/test_query/logical_and/user_005.json",
				} {
					assert.Nil(t, database.Insert(context.Background(), s.mustResource(f, resourceType)))
				}
			},
			filter: fmt.Sprintf("(userName sw %s) and (meta.created gt %s)", strconv.Quote("user"), strconv.Quote("2019-12-21T10:00:00")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "meta", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, "user002", results[0].NewFluentNavigator().FocusName("userName").Current().Raw())
				assert.Equal(t, "user005", results[1].NewFluentNavigator().FocusName("userName").Current().Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			client, err := s.newClient()
			require.Nil(t, err)
			coll := client.Database(testMongoDatabaseName).Collection(t.Name())
			database := DB(resourceType, log.None(), coll, Options())
			test.prepare(t, database)
			r, err := database.Query(context.Background(), test.filter, test.sort, test.pagination, test.projection)
			test.expect(t, r, err)
		})
	}
}

func (s *MongoDatabaseTestSuite) TestSaveGetDeleteCount() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	resource := s.mustResource("/user_001.json", resourceType)

	client, err := s.newClient()
	s.Require().Nil(err)
	coll := client.Database(testMongoDatabaseName).Collection(s.T().Name())
	database := DB(resourceType, log.None(), coll, Options())

	err = database.Insert(context.Background(), resource)
	assert.Nil(s.T(), err)

	n, err := database.Count(context.Background(), fmt.Sprintf("id eq %s", strconv.Quote(resource.ID())))
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, n)

	got, err := database.Get(context.Background(), resource.ID(), nil)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), got)

	err = database.Delete(context.Background(), got)
	assert.Nil(s.T(), err)

	n, err = database.Count(context.Background(), fmt.Sprintf("id eq %s", strconv.Quote(resource.ID())))
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, n)
}

// connect to MongoDB docker container before the suite
func (s *MongoDatabaseTestSuite) SetupSuite() {
	var err error

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	go func() {
		<-c
		s.TearDownSuite()
	}()

	s.dockerPool, err = dockertest.NewPool(testDockerEndpoint)
	s.Require().Nil(err)

	s.dockerResource, err = s.dockerPool.Run(testMongoImageName, testMongoImageTag, []string{
		fmt.Sprintf("MONGODB_USERNAME=%s", testMongoUserName),
		fmt.Sprintf("MONGODB_PASSWORD=%s", testMongoUserSecret),
		fmt.Sprintf("MONGODB_DATABASE=%s", testMongoDatabaseName),
	})
	s.Require().Nil(err)

	s.newClient = func() (client *mongo.Client, err error) {
		mongoUri := fmt.Sprintf("mongodb://%s:%s@localhost:%s/%s",
			testMongoUserName,
			testMongoUserSecret,
			s.dockerResource.GetPort("27017/tcp"),
			testMongoDatabaseName,
		)
		client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUri))
		if err != nil {
			return
		}
		err = client.Ping(context.Background(), readpref.Primary())
		if err != nil {
			return
		}
		return
	}

	err = s.dockerPool.Retry(func() error {
		_, err := s.newClient()
		return err
	})
	s.Require().Nil(err)
}

// Clean and disconnect from MongoDB docker container after the suite
func (s *MongoDatabaseTestSuite) TearDownSuite() {
	_ = s.dockerPool.Purge(s.dockerResource)
}

func (s *MongoDatabaseTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *MongoDatabaseTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *MongoDatabaseTestSuite) mustSchema(filePath string) *spec.Schema {
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

func envOrDefault(name string, defaultValue string) string {
	if v := os.Getenv(name); len(v) > 0 {
		return v
	}
	return defaultValue
}
