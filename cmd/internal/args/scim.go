package args

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

// Scim the configuration options related to the core SCIM specification.
type Scim struct {
	// Path to the service provider config JSON file
	ServiceProviderConfigPath string
	// Path to the user resource type JSON file
	UserResourceTypePath string
	// Path to the group resource type JSON file
	GroupResourceTypePath string
	// Path to the directory containing all schema JSON file
	SchemasDirectory string
}

// ParseServiceProviderConfig returns an instance of spec.ServiceProviderConfig from the JSON definition at
// ServiceProviderConfigPath, or an error.
func (arg *Scim) ParseServiceProviderConfig() (*spec.ServiceProviderConfig, error) {
	f, err := os.Open(arg.ServiceProviderConfigPath)
	if err != nil {
		return nil, err
	}

	config := new(spec.ServiceProviderConfig)
	err = json.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// RegisterSchemas iterates through all JSON files in the SchemasDirectory directory and registers all of them as
// schema files.
func (arg *Scim) RegisterSchemas() error {
	return filepath.Walk(arg.SchemasDirectory, func(path string, info os.FileInfo, err error) error {
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

		schema := new(spec.Schema)
		err = json.NewDecoder(f).Decode(schema)
		if err != nil {
			return err
		}

		spec.Schemas().Register(schema)
		return nil
	})
}

// ParseUserResourceType returns the parsed spec.ResourceType from the JSON schema definition at UserResourceTypePath.
// Caller must make sure RegisterSchemas was invoked first.
func (arg *Scim) ParseUserResourceType() (*spec.ResourceType, error) {
	return arg.parseResourceType(arg.UserResourceTypePath)
}

// ParseGroupResourceType returns the parsed spec.ResourceType from the JSON schema definition at GroupResourceTypePath
// Caller must make sure RegisterSchemas was invoked first.
func (arg *Scim) ParseGroupResourceType() (*spec.ResourceType, error) {
	return arg.parseResourceType(arg.GroupResourceTypePath)
}

func (arg *Scim) parseResourceType(path string) (*spec.ResourceType, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	rt := new(spec.ResourceType)
	err = json.NewDecoder(f).Decode(rt)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

func (arg *Scim) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "user-resource-type",
			Usage:       "Absolute file path to User resource type JSON definition",
			EnvVars:     []string{"USER_RESOURCE_TYPE"},
			Required:    true,
			Destination: &arg.UserResourceTypePath,
		},
		&cli.StringFlag{
			Name:        "group-resource-type",
			Usage:       "Absolute file path to Group resource type JSON definition",
			EnvVars:     []string{"GROUP_RESOURCE_TYPE"},
			Required:    true,
			Destination: &arg.GroupResourceTypePath,
		},
		&cli.StringFlag{
			Name:        "schemas-dir",
			Usage:       "Absolute path to the directory containing all schema JSON definitions",
			EnvVars:     []string{"SCHEMAS_DIR"},
			Required:    true,
			Destination: &arg.SchemasDirectory,
		},
		&cli.StringFlag{
			Name:        "service-provider-config",
			Usage:       "Absolute path to service Provider Config JSON definition",
			EnvVars:     []string{"SERVICE_PROVIDER_CONFIG"},
			Required:    true,
			Destination: &arg.ServiceProviderConfigPath,
		},
	}
}
