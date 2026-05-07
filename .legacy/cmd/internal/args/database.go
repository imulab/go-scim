package args

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	scimmongo "github.com/imulab/go-scim/mongo/v2"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"path/filepath"
	"strings"
)

// MemoryDB is the configuration options related to a in-memory db.DB implementation.
type MemoryDB struct {
	UseMemoryDB bool
}

func (arg *MemoryDB) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "memory",
			Usage:       "Use the in-memory database implementation. If true, excludes all other database options",
			Value:       false,
			Destination: &arg.UseMemoryDB,
		},
	}
}

// MongoDB is the configuration options related to using the MongoDB implementation of db.DB
type MongoDB struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Database    string
	Options     string
	MetadataDir string
}

// Url returns the MongoDB connection URL created using the set options.
func (arg *MongoDB) Url() string {
	url := "mongodb://"
	if len(arg.Username) > 0 {
		url += fmt.Sprintf("%s:%s@", arg.Username, arg.Password)
	}
	url += arg.Host
	if arg.Port > 0 {
		url += fmt.Sprintf(":%d", arg.Port)
	}
	if len(arg.Database) > 0 {
		url += fmt.Sprintf("/%s", arg.Database)
	}
	if len(arg.Options) > 0 {
		url += fmt.Sprintf("?%s", arg.Options)
	}
	return url
}

// Connect returns a connected MongoDB client using the set options, or an error.
func (arg *MongoDB) Connect(ctx context.Context) (client *mongo.Client, err error) {
	err = backoff.Retry(func() (connectErr error) {
		client, connectErr = mongo.Connect(ctx, options.Client().ApplyURI(arg.Url()))
		return
	}, backoff.NewExponentialBackOff())
	return
}

// RegisterMetadata iterates all JSON files in the MetadataDir and registers its content as SCIM MongoDB metadata.
func (arg *MongoDB) RegisterMetadata() error {
	if len(arg.MetadataDir) == 0 {
		return nil
	}

	return filepath.Walk(arg.MetadataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		return scimmongo.ReadMetadataFromReader(f)
	})
}

func (arg *MongoDB) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "mongo-host",
			Usage:       "Hostname of MongoDB",
			EnvVars:     []string{"MONGO_HOST"},
			Value:       "localhost",
			Destination: &arg.Host,
		},
		&cli.IntFlag{
			Name:        "mongo-port",
			Usage:       "Port of MongoDB",
			EnvVars:     []string{"MONGO_PORT"},
			Value:       27017,
			Destination: &arg.Port,
		},
		&cli.StringFlag{
			Name:        "mongo-username",
			Usage:       "Username for MongoDB",
			EnvVars:     []string{"MONGO_USERNAME"},
			Destination: &arg.Username,
		},
		&cli.StringFlag{
			Name:        "mongo-password",
			Usage:       "Password for MongoDB",
			EnvVars:     []string{"MONGO_PASSWORD"},
			Destination: &arg.Password,
		},
		&cli.StringFlag{
			Name:        "mongo-database",
			Usage:       "Database for MongoDB",
			EnvVars:     []string{"MONGO_DATABASE"},
			Destination: &arg.Database,
		},
		&cli.StringFlag{
			Name:        "mongo-options",
			Usage:       "Options for MongoDB",
			EnvVars:     []string{"MONGO_OPT"},
			Destination: &arg.Options,
		},
		&cli.StringFlag{
			Name:        "mongo-metadata-dir",
			Usage:       "Path to the directory containing MongoDB metadata JSON files",
			EnvVars:     []string{"MONGO_METADATA_DIR"},
			Destination: &arg.MetadataDir,
		},
	}
}
