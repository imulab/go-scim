package mongo

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/test"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

var (
	dockerEndpoint    = test.EnvOrDefault("TEST_DOCKER_ENDPOINT", "")
	mongoImageName    = test.EnvOrDefault("TEST_DOCKER_MONGO_IMAGE", "bitnami/mongodb")
	mongoImageTag     = test.EnvOrDefault("TEST_DOCKER_MONGO_TAG", "latest")
	mongoUserName     = test.EnvOrDefault("TEST_DOCKER_MONGO_USERNAME", "testUser")
	mongoUserSecret   = test.EnvOrDefault("TEST_DOCKER_MONGO_SECRET", "s3cret")
	mongoDatabaseName = test.EnvOrDefault("TEST_DOCKER_MONGO_DB_NAME", "persistence_test_suite")
	debugTest         = test.EnvExists("TEST_DEBUG")
)

type PersistenceTestSuite struct {
	suite.Suite
	pool      *dockertest.Pool
	mongoDb   *dockertest.Resource
	newClient func() (*mongo.Client, error)
}

func (s *PersistenceTestSuite) SetupSuite() {
	var err error

	// register for a shutdown hook, because TestDownSuite cannot be relied upon in case of termination signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	go func() {
		<-c
		s.TearDownSuite()
	}()

	// connect to docker
	s.pool, err = dockertest.NewPool(dockerEndpoint)
	s.Require().Nil(err)

	// run MongoDB image
	s.mongoDb, err = s.pool.Run(mongoImageName, mongoImageTag, []string{
		fmt.Sprintf("MONGODB_USERNAME=%s", mongoUserName),
		fmt.Sprintf("MONGODB_PASSWORD=%s", mongoUserSecret),
		fmt.Sprintf("MONGODB_DATABASE=%s", mongoDatabaseName),
	})
	s.Require().Nil(err)

	// formulate client constructor
	s.newClient = func() (*mongo.Client, error) {
		mongoUri := fmt.Sprintf("mongodb://%s:%s@localhost:%s/%s",
			mongoUserName,
			mongoUserSecret,
			s.mongoDb.GetPort("27017/tcp"),
			mongoDatabaseName,
		)
		if debugTest {
			log.Printf("try connection to MongoDB at %s\n", mongoUri)
		}

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUri))
		if err != nil {
			if debugTest {
				log.Println(err)
			}
			return nil, err
		}

		if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
			if debugTest {
				log.Println(err)
			}
			return nil, err
		}

		return client, nil
	}

	// ensure connection is alive
	err = s.pool.Retry(func() error {
		_, err := s.newClient()
		return err
	})
	s.Require().Nil(err)
}

func (s *PersistenceTestSuite) TestTotal() {
	client, err := s.newClient()
	s.Require().Nil(err)

	tests := []struct {
		name   string
		setup  func(t *testing.T, client *mongo.Client) *persistenceProvider
		assert func(t *testing.T, n int64, err error)
	}{
		{
			name: "total from a single collection",
			setup: func(t *testing.T, client *mongo.Client) *persistenceProvider {
				var (
					collection   *mongo.Collection
					resourceType *core.ResourceType
				)
				{
					collection = client.
						Database(mongoDatabaseName, options.Database()).
						Collection(fmt.Sprintf("%s/%s", s.T().Name(), "1"), options.Collection())

					_ = core.Schemas.MustLoad("../resource/schema/test_object_schema.json")
					_ = core.Meta.MustLoad("../resource/metadata/test_metadata.json", new(core.DefaultMetadataProvider))
					resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/test_object_resource_type.json")
				}

				for _, path := range []string{
					"../resource/test/test_object_1.json",
					"../resource/test/test_object_2.json",
				} {
					resource := test.MustResource(path, resourceType)
					_, err = collection.InsertOne(context.Background(), newBsonAdapter(resource), options.InsertOne())
					s.Require().Nil(err)
				}

				return &persistenceProvider{
					resourceTypes: []*core.ResourceType{
						resourceType,
					},
					collections: map[string]*mongo.Collection{
						resourceType.Id: collection,
					},
					maxTimePercent: 80,
				}
			},
			assert: func(t *testing.T, n int64, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(2), n)
			},
		},
		{
			name: "total from multiple collection",
			setup: func(t *testing.T, client *mongo.Client) *persistenceProvider {
				var (
					c1 *mongo.Collection
					c2 *mongo.Collection

					rt1 *core.ResourceType
					rt2 *core.ResourceType
				)
				{
					// prepare two collections
					c1 = client.
						Database(mongoDatabaseName, options.Database()).
						Collection(fmt.Sprintf("%s/%s", s.T().Name(), "2"), options.Collection())
					c2 = client.
						Database(mongoDatabaseName, options.Database()).
						Collection(fmt.Sprintf("%s/%s", s.T().Name(), "3"), options.Collection())

					_ = core.Schemas.MustLoad("../resource/schema/test_object_schema.json")
					_ = core.Meta.MustLoad("../resource/metadata/test_metadata.json", new(core.DefaultMetadataProvider))
					rt1 = core.ResourceTypes.MustLoad("../resource/resource_type/test_object_resource_type.json")

					_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
					_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
					rt2 = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
				}

				// insert 2 resources into the first collection
				for _, path := range []string{
					"../resource/test/test_object_1.json",
					"../resource/test/test_object_2.json",
				} {
					resource := test.MustResource(path, rt1)
					_, err = c1.InsertOne(context.Background(), newBsonAdapter(resource), options.InsertOne())
					s.Require().Nil(err)
				}

				// insert 1 resource into the second collection
				for _, path := range []string{
					"../resource/test/test_user_1.json",
				} {
					resource := test.MustResource(path, rt2)
					_, err = c2.InsertOne(context.Background(), newBsonAdapter(resource), options.InsertOne())
					s.Require().Nil(err)
				}

				return &persistenceProvider{
					resourceTypes: []*core.ResourceType{rt1, rt2},
					collections: map[string]*mongo.Collection{
						rt1.Id: c1,
						rt2.Id: c2,
					},
					maxTimePercent: 80,
				}
			},
			assert: func(t *testing.T, n int64, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(3), n)
			},
		},
	}

	for _, each := range tests {
		s.T().Run(each.name, func(t *testing.T) {
			provider := each.setup(t, client)
			n, err := provider.Total(context.Background())
			each.assert(t, n, err)
		})
	}
}

func (s *PersistenceTestSuite) TearDownSuite() {
	if s.pool != nil && s.mongoDb != nil {
		s.Require().Nil(s.pool.Purge(s.mongoDb))
	}
}

func TestPersistenceProvider(t *testing.T) {
	suite.Run(t, new(PersistenceTestSuite))
}
