package args

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"path/filepath"
	"strings"
)

type Mongo struct {
	Host           string
	Port           int
	Username       string
	Password       string
	Database       string
	Options        string
	MetadataFolder string
}

func (arg *Mongo) Url() string {
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

func (arg *Mongo) Connect() (*mongo.Client, error) {
	mc, err := mongo.Connect(context.Background(), options.Client().ApplyURI(arg.Url()))
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func (arg *Mongo) ReadMetadataBytes() ([][]byte, error) {
	if len(arg.MetadataFolder) == 0 {
		return [][]byte{}, nil
	}
	results := make([][]byte, 0)
	err := filepath.Walk(arg.MetadataFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".json") {
			raw, err := readFile(path)
			if err != nil {
				return err
			}
			results = append(results, raw)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (arg *Mongo) Flags() []cli.Flag {
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
			Name:        "mongo-metadata-folder",
			Usage:       "Path to the folder containing MongoDB metadata JSON files",
			EnvVars:     []string{"MONGO_METADATA_FOLDER"},
			Destination: &arg.MetadataFolder,
		},
	}
}
