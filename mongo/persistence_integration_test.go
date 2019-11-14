package mongo

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/test"
	"github.com/ory/dockertest"
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

type PersistenceTestSuite struct {
	suite.Suite
	pool 	*dockertest.Pool
	mongoDb	*dockertest.Resource
	newClient func() (*mongo.Client, error)
}

func (s *PersistenceTestSuite) SetupSuite() {
	var (
		dockerEndpoint    = test.EnvOrDefault("TEST_DOCKER_ENDPOINT", "")
		mongoImageName    = test.EnvOrDefault("TEST_DOCKER_MONGO_IMAGE", "bitnami/mongodb")
		mongoImageTag     = test.EnvOrDefault("TEST_DOCKER_MONGO_TAG", "latest")
		mongoUserName     = test.EnvOrDefault("TEST_DOCKER_MONGO_USERNAME", "testUser")
		mongoUserSecret   = test.EnvOrDefault("TEST_DOCKER_MONGO_SECRET", "s3cret")
		mongoDatabaseName = test.EnvOrDefault("TEST_DOCKER_MONGO_DB_NAME", "persistence_test_suite")
		debugTest         = test.EnvExists("TEST_DEBUG")

		err     error
	)

	// register for a shutdown hook, because TestDownSuite cannot be relied upon in case of termination signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
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
	println("test total")
}

func (s *PersistenceTestSuite) TearDownSuite() {
	if s.pool != nil && s.mongoDb != nil {
		s.Require().Nil(s.pool.Purge(s.mongoDb))
	}
}

func TestPersistenceProvider(t *testing.T) {
	suite.Run(t, new(PersistenceTestSuite))
}
