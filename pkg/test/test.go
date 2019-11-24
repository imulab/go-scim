package test

import (
	"github.com/imulab/go-scim/pkg/core"
	"github.com/imulab/go-scim/pkg/json"
	"io/ioutil"
	"os"
)

// Return environment variable value with a default value in case environment variable does not exist.
func EnvOrDefault(envVar string, defaultVal string) string {
	env := os.Getenv(envVar)
	if len(env) == 0 {
		return defaultVal
	}
	return env
}

// Return true if the environment variable is not empty
func EnvExists(envVar string) bool {
	return len(os.Getenv(envVar)) > 0
}

func MustResource(filePath string, resourceType *core.ResourceType) *core.Resource {
	resource := core.Resources.New(resourceType)

	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	err = json.Deserialize(raw, resource)
	if err != nil {
		panic(err)
	}

	return resource
}
