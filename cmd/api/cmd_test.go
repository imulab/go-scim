package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/ory/dockertest"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
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
	testRabbitImageName   = envOrDefault("TEST_DOCKER_RABBIT_IMAGE", "bitnami/rabbitmq")
	testRabbitImageTag    = envOrDefault("TEST_DOCKER_RABBIT_TAG", "latest")
	testRabbitUserName    = envOrDefault("TEST_DOCKER_RABBIT_USERNAME", "user")
	testRabbitUserSecret  = envOrDefault("TEST_DOCKER_RABBIT_SECRET", "bitnami")
)

func TestCommand(t *testing.T) {
	s := new(CommandTestSuite)
	suite.Run(t, s)
}

type CommandTestSuite struct {
	sync.Mutex
	suite.Suite
	dockerPool           *dockertest.Pool
	dockerMongoResource  *dockertest.Resource
	dockerRabbitResource *dockertest.Resource
}

func (s *CommandTestSuite) TestCommand() {
	app := &cli.App{
		Name: "scim",
		Commands: []*cli.Command{
			Command(),
		},
	}

	go func() {
		err := app.Run([]string{
			"scim", "api",
			"--log-level", "DEBUG",
			"--mongo-host", "localhost",
			"--mongo-port", s.dockerMongoResource.GetPort("27017/tcp"),
			"--mongo-username", testMongoUserName,
			"--mongo-password", testMongoUserSecret,
			"--mongo-database", testMongoDatabaseName,
			"--mongo-metadata-dir", s.mustAbs("../../public/mongo_metadata"),
			"--rabbit-host", "localhost",
			"--rabbit-port", s.dockerRabbitResource.GetPort("5672/tcp"),
			"--rabbit-username", testRabbitUserName,
			"--rabbit-password", testRabbitUserSecret,
			"--rabbit-vhost", "/",
			"--schemas-dir", s.mustAbs("../../public/schemas"),
			"--user-resource-type", s.mustAbs("../../public/resource_types/user_resource_type.json"),
			"--group-resource-type", s.mustAbs("../../public/resource_types/group_resource_type.json"),
			"--service-provider-config", s.mustAbs("../../public/service_provider_config.json"),
		})
		assert.Nil(s.T(), err)
	}()

	err := backoff.Retry(func() error {
		resp, err := http.Get("http://localhost:8080/health")
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New("non-200 status")
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 10))
	assert.Nil(s.T(), err)
}

func (s *CommandTestSuite) SetupSuite() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	go func() {
		<-c
		s.TearDownSuite()
	}()

	var g errgroup.Group
	g.Go(func() error {
		return s.setupDockerMongoDB()
	})
	g.Go(func() error {
		return s.setupDockerRabbitMQ()
	})
	s.Require().Nil(g.Wait())
}

func (s *CommandTestSuite) getDockerPool() (*dockertest.Pool, error) {
	var err error
	if s.dockerPool == nil {
		s.Lock()
		if s.dockerPool == nil {
			s.dockerPool, err = dockertest.NewPool(testDockerEndpoint)
		}
		s.Unlock()
	}
	return s.dockerPool, err
}

func (s *CommandTestSuite) setupDockerRabbitMQ() (err error) {
	s.dockerPool, err = s.getDockerPool()
	if err != nil {
		return
	}

	s.dockerRabbitResource, err = s.dockerPool.Run(testRabbitImageName, testRabbitImageTag, []string{
		fmt.Sprintf("RABBITMQ_USERNAME=%s", testRabbitUserName),
		fmt.Sprintf("RABBITMQ_PASSWORD=%s", testRabbitUserSecret),
	})
	if err != nil {
		return
	}

	err = s.dockerPool.Retry(func() error {
		rabbitUri := fmt.Sprintf("amqp://%s:%s@localhost:%s/",
			testRabbitUserName,
			testRabbitUserSecret,
			s.dockerRabbitResource.GetPort("5672/tcp"),
		)
		if _, err := amqp.Dial(rabbitUri); err != nil {
			return err
		} else {
			return nil
		}
	})
	if err != nil {
		return
	}

	return
}

func (s *CommandTestSuite) setupDockerMongoDB() (err error) {
	s.dockerPool, err = s.getDockerPool()
	if err != nil {
		return
	}

	s.dockerMongoResource, err = s.dockerPool.Run(testMongoImageName, testMongoImageTag, []string{
		fmt.Sprintf("MONGODB_USERNAME=%s", testMongoUserName),
		fmt.Sprintf("MONGODB_PASSWORD=%s", testMongoUserSecret),
		fmt.Sprintf("MONGODB_DATABASE=%s", testMongoDatabaseName),
	})
	if err != nil {
		return
	}

	err = s.dockerPool.Retry(func() error {
		mongoUri := fmt.Sprintf("mongodb://%s:%s@localhost:%s/%s",
			testMongoUserName,
			testMongoUserSecret,
			s.dockerMongoResource.GetPort("27017/tcp"),
			testMongoDatabaseName,
		)
		if client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUri)); err != nil {
			return err
		} else if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
			return err
		} else {
			return nil
		}
	})
	if err != nil {
		return
	}

	return
}

func (s *CommandTestSuite) TearDownSuite() {
	if s.dockerMongoResource != nil {
		_ = s.dockerPool.Purge(s.dockerMongoResource)
	}
	if s.dockerRabbitResource != nil {
		_ = s.dockerPool.Purge(s.dockerRabbitResource)
	}
}

func (s *CommandTestSuite) mustAbs(path string) string {
	p, err := filepath.Abs(path)
	s.Require().Nil(err)
	return p
}

func envOrDefault(name string, defaultValue string) string {
	if v := os.Getenv(name); len(v) > 0 {
		return v
	}
	return defaultValue
}
