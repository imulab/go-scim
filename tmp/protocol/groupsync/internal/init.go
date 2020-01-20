package internal

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/spec"
)

func init() {
	Schema = new(spec.Schema)
	err := json.Unmarshal([]byte(SchemaJSON), Schema)
	if err != nil {
		panic(err)
	}
	spec.SchemaHub.Put(Schema)

	ResourceType = new(spec.ResourceType)
	err = json.Unmarshal([]byte(ResourceTypeJSON), ResourceType)
	if err != nil {
		panic(err)
	}
}
