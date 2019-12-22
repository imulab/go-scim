package args

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/spec"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

type Scim struct {
	// Path to the service provider config JSON file
	ServiceProviderConfigPath string
	// Path to the user resource type JSON file
	UserResourceTypePath string
	// Path to the group resource type JSON file
	GroupResourceTypePath string
	// Path to the folder containing all schema JSON file
	SchemasFolderPath string
}

func (arg *Scim) ParseServiceProviderConfig() (*spec.ServiceProviderConfig, error) {
	raw, err := readFile(arg.ServiceProviderConfigPath)
	if err != nil {
		return nil, err
	}
	spc := new(spec.ServiceProviderConfig)
	if err = json.Unmarshal(raw, spc); err != nil {
		return nil, err
	}
	return spc, nil
}

func (arg *Scim) ParseSchemas() ([]*spec.Schema, error) {
	results := make([]*spec.Schema, 0)
	err := filepath.Walk(arg.SchemasFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".json") {
			raw, err := readFile(path)
			if err != nil {
				return err
			}
			schema := new(spec.Schema)
			err = json.Unmarshal(raw, schema)
			if err != nil {
				return err
			}
			results = append(results, schema)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (arg *Scim) ParseUserResourceType() (*spec.ResourceType, error) {
	return arg.parseResourceType(arg.UserResourceTypePath)
}

func (arg *Scim) ParseGroupResourceType() (*spec.ResourceType, error) {
	return arg.parseResourceType(arg.GroupResourceTypePath)
}

func (arg *Scim) parseResourceType(path string) (*spec.ResourceType, error) {
	raw, err := readFile(path)
	if err != nil {
		return nil, err
	}
	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
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
			Name:        "schemas-folder",
			Usage:       "Absolute path to the folder containing all schema JSON definitions",
			EnvVars:     []string{"SCHEMAS_FOLDER"},
			Required:    true,
			Destination: &arg.SchemasFolderPath,
		},
		&cli.StringFlag{
			Name:        "service-provider-config",
			Usage:       "Absolute path to Service Provider Config JSON definition",
			EnvVars:     []string{"SERVICE_PROVIDER_CONFIG"},
			Required:    true,
			Destination: &arg.ServiceProviderConfigPath,
		},
	}
}