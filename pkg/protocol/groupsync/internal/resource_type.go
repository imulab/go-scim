package internal

import "github.com/imulab/go-scim/pkg/core/spec"

var (
	ResourceType	*spec.ResourceType
	ResourceTypeJSON = `
{
	"id": "GroupSync",
	"name": "Group Sync",
  	"description": "Group sync internal resource type",
	"endpoint": "",
	"schema": "urn:imulab:scim:schemas:internal:2.0:GroupSync"
}
`
)