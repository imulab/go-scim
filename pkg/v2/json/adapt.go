package json

import (
	"github.com/imulab/go-scim/pkg/v2/json/internal"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// SchemaToSerializable returns a Serializable wrapper for a schema so it can be used to call json.Serialize
func SchemaToSerializable(sch *spec.Schema) Serializable {
	return &internal.SerializableSchema{Sch: sch}
}

// ResourceTypeToSerializable returns a Serializable wrapper for a resource type so it can be used to call json.Serialize
func ResourceTypeToSerializable(resourceType *spec.ResourceType) Serializable {
	return &internal.SerializableResourceType{ResourceType: resourceType}
}
