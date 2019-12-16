package groupsync

import (
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/groupsync/internal"
)

// Get the group sync resource schema
func Schema() *spec.Schema {
	return internal.Schema
}

// Get the group sync resource type
func ResourceType() *spec.ResourceType {
	return internal.ResourceType
}