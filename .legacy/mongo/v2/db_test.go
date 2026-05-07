package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/db"
	scimjson "github.com/imulab/go-scim/pkg/v2/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
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
	suite.Run(t, s)
}

type MongoDatabaseTestSuite struct {
	suite.Suite
	resourceType   *spec.ResourceType
	dockerPool     *dockertest.Pool
	dockerResource *dockertest.Resource
	newClient      func() (*mongo.Client, error)
}

func (s *MongoDatabaseTestSuite) TestQuery() {
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
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "userName": "user001"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User", "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"],
  "id": "user002",
  "userName": "user002"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:FooBarUser"],
  "id": "user003",
  "userName": "user003"
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
				}
			},
			filter: fmt.Sprintf("schemas eq %s", strconv.Quote("urn:ietf:params:scim:schemas:core:2.0:User")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, "user001", results[0].Navigator().Dot("userName").Current().Raw())
				assert.Equal(t, "user002", results[1].Navigator().Dot("userName").Current().Raw())
			},
		},
		{
			name: "userName eq",
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "userName": "user001"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user002",
  "userName": "user002"
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
				}
			},
			filter: fmt.Sprintf("userName eq %s", strconv.Quote("user001")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "user001", results[0].Navigator().Dot("userName").Current().Raw())
			},
		},
		{
			name: "nested eq",
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "userName": "user001",
  "name": { "familyName": "Qiu" }
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user002",
  "userName": "user002",
  "name": { "familyName": "Q" }
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
				}
			},
			filter: fmt.Sprintf("name.familyName eq %s", strconv.Quote("Q")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "userName", "name"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "Q",
					results[0].Navigator().Dot("name").Dot("familyName").Current().Raw())
			},
		},
		{
			name: "multiValue nested eq",
			prepare: func(t *testing.T, database db.DB) {
				for _, f := range []string{
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "userName": "user001",
  "emails": [
    { "value": "foo@bar.com", "type": "work", "primary": true, "display": "foo@bar.com" },
    { "value": "imulab@bar.com", "type": "home", "display": "imulab@bar.com" }
  ]
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user002",
  "userName": "user002",
  "emails": [
    { "value": "imulab@foo.com", "type": "work", "primary": true, "display": "imulab@foo.com" },
    { "value": "imulab@bar.com", "type": "home", "display": "imulab@bar.com" }
  ]
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:InvalidUser"],
  "id": "user003",
  "userName": "user003",
  "emails": [
    { "value": "foo@bar.com", "type": "work", "primary": true, "display": "foo@bar.com" }
  ]
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
				}
			},
			filter: fmt.Sprintf("emails.value eq %s", strconv.Quote("foo@bar.com")),
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "userName", "emails"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, "user001", results[0].Navigator().Dot("userName").Current().Raw())
				assert.Equal(t, "user003", results[1].Navigator().Dot("userName").Current().Raw())
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
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "userName": "user001",
  "nickName": "foo"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user002",
  "userName": "user002"
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
				}
			},
			filter: "nickName pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "nickName", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "user001", results[0].Navigator().Dot("userName").Current().Raw())
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
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "userName": "user001",
  "name": { "givenName": "Weinan", "honorificPrefix": "Mr." }
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "75FEBF2C-6443-4CF1-9554-ADE2A37CD8DE",
  "userName": "user002",
  "name": { "familyName": "Qiu" }
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
				}
			},
			filter: "name.familyName pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "name", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "user002", results[0].Navigator().Dot("userName").Current().Raw())
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
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "userName": "user001",
  "emails": [
    {"value": "foo@bar.com","type": "work","primary": true},
    {"value": "imulab@bar.com","type": "home","display": "imulab@bar.com"}
  ]
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user002",
  "userName": "user002",
  "emails": [
    {"value": "imulab@foo.com","type": "work","primary": true},
    {"value": "imulab@bar.com","type": "home"}
  ]
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user003",
  "userName": "user003",
  "emails": [
    {"value": "foo@bar.com","type": "work","primary": true,"display": "foo@bar.com"}
  ]
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
				}
			},
			filter: "emails.display pr",
			sort: &crud.Sort{
				By:    "userName",
				Order: crud.SortAsc,
			},
			pagination: nil,
			projection: &crud.Projection{
				Attributes: []string{"schemas", "id", "emails", "userName"},
			},
			expect: func(t *testing.T, results []*prop.Resource, err error) {
				assert.Nil(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, "user001", results[0].Navigator().Dot("userName").Current().Raw())
				assert.Equal(t, "user003", results[1].Navigator().Dot("userName").Current().Raw())
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
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user001",
  "meta": {"created": "2019-12-21T09:00:00"},
  "userName": "user001"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user002",
  "meta": {"created": "2019-12-21T10:30:00"},
  "userName": "user002"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user003",
  "meta": {"created": "2019-12-21T10:00:00"},
  "userName": "user003"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "foobar",
  "meta": {"created": "2019-12-21T10:30:00"},
  "userName": "foobar"
}
`,
					`
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "user005",
  "meta": {"created": "2019-12-21T11:00:00"},
  "userName": "user005"
}
`,
				} {
					r := prop.NewResource(s.resourceType)
					assert.Nil(t, scimjson.Deserialize([]byte(f), r))
					assert.Nil(t, database.Insert(context.Background(), r))
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
				assert.Equal(t, "user002", results[0].Navigator().Dot("userName").Current().Raw())
				assert.Equal(t, "user005", results[1].Navigator().Dot("userName").Current().Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			client, err := s.newClient()
			require.Nil(t, err)
			coll := client.Database(testMongoDatabaseName).Collection(t.Name())
			database := DB(s.resourceType, coll, Options())
			test.prepare(t, database)
			r, err := database.Query(context.Background(), test.filter, test.sort, test.pagination, test.projection)
			test.expect(t, r, err)
		})
	}
}

func (s *MongoDatabaseTestSuite) TestSaveGetDeleteCount() {
	resource := prop.NewResource(s.resourceType)
	assert.Nil(s.T(), scimjson.Deserialize([]byte(`
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User"
  ],
  "id": "a5866759-32ca-4e2a-9808-a0fe74f94b18",
  "meta": {
    "resourceType": "User",
    "created": "2019-11-20T13:09:00",
    "lastModified": "2019-11-20T13:09:00",
    "location": "https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
    "version": "W/\"1\""
  },
  "userName": "user001",
  "emails": [
    {
      "value": "imulab@foo.com",
      "type": "work",
      "primary": true,
      "display": "imulab@foo.com"
    }
  ]
}
`), resource))

	client, err := s.newClient()
	s.Require().Nil(err)
	coll := client.Database(testMongoDatabaseName).Collection(s.T().Name())
	database := DB(s.resourceType, coll, Options())

	err = database.Insert(context.Background(), resource)
	assert.Nil(s.T(), err)

	n, err := database.Count(context.Background(), fmt.Sprintf("id eq %s", strconv.Quote(resource.IdOrEmpty())))
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, n)

	got, err := database.Get(context.Background(), resource.IdOrEmpty(), nil)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), got)

	err = database.Delete(context.Background(), got)
	assert.Nil(s.T(), err)

	n, err = database.Count(context.Background(), fmt.Sprintf("id eq %s", strconv.Quote(resource.IdOrEmpty())))
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, n)
}

// connect to MongoDB docker container before the suite
func (s *MongoDatabaseTestSuite) SetupSuite() {
	s.parseResourceType()
	s.setupDockerMongoDB()
}

func (s *MongoDatabaseTestSuite) setupDockerMongoDB() {
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

func (s *MongoDatabaseTestSuite) parseResourceType() {
	for _, each := range []struct {
		filepath  string
		structure interface{}
		post      func(parsed interface{})
	}{
		{
			filepath:  "../../public/schemas/core_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../public/schemas/user_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../public/resource_types/user_resource_type.json",
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

// Clean and disconnect from MongoDB docker container after the suite
func (s *MongoDatabaseTestSuite) TearDownSuite() {
	_ = s.dockerPool.Purge(s.dockerResource)
}

func envOrDefault(name string, defaultValue string) string {
	if v := os.Getenv(name); len(v) > 0 {
		return v
	}
	return defaultValue
}
