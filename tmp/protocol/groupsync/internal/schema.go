package internal

import (
	"github.com/imulab/go-scim/core/spec"
)

var (
	Schema *spec.Schema
	SchemaJSON = `
{
	"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync",
	"name": "GroupSync",
	"attributes": [
		{
			"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:group",
			"name": "group",
			"type": "complex",
			"subAttributes": [
				{
					"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:group.id",
					"name": "id",
					"type": "string",
					"_path": "group.id",
					"_index": 0
				},
				{
					"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:group.location",
					"name": "location",
					"type": "string",
					"_path": "group.location",
					"_index": 1
				},
				{
					"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:group.display",
					"name": "display",
					"type": "string",
					"_path": "group.display",
					"_index": 2
				}
			],
			"_index": 100,
			"_path": "group"
		},
		{
			"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:diff",
			"name": "diff",
			"type": "complex",
			"multiValued": true,
			"subAttributes": [
				{
					"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:diff.id",
					"name": "id",
					"type": "string",
					"_path": "diff.id",
					"_index": 0
				},
				{
					"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:diff.type",
					"name": "type",
					"type": "string",
					"canonicalValues": ["unknown", "direct", "indirect"],
					"_path": "diff.type",
					"_index": 1
				},
				{
					"id": "urn:imulab:scim:schemas:internal:2.0:GroupSync:diff.status",
					"name": "status",
					"type": "string",
					"canonicalValues": ["joined", "left"],
					"_path": "diff.status",
					"_index": 2
				}
			],
			"_index": 100,
			"_path": "group",
			"_annotations": ["@AutoCompact"]
		}
	]
}
`
)
