package core

// This files lists some of the recognized annotations
var (
	// Denotes the field is a 'schemas' field. The 'schemas' shall at least contain the schema id
	// of the base resource type, and may optionally contain schema ids from the resource type extensions.
	// When containing a schema id which is a required schema extension, make sure there is a top level
	// complex property of this name, and it is not unassigned.
	AnnotationSchemas = "@schemas"

	// Denotes the field is a 'id' field. During resource creation, an uuid shall be generated for its value.
	AnnotationId = "@id"

	// Denotes the field is a 'meta' property. It shall contain the sub properties of 'resourceType', 'created',
	// 'lastModified', 'location' and 'version'. At all times, 'resourceType' shall reflect the name of the resource's
	// resource type; 'location' shall reflect the absolute url to this resource. During resource creation, 'created'
	// and 'lastModified' are both assigned the current time, and 'version' is assigned to an initial value; During
	// non-delete resource modifications, 'lastModified' is set to current time and 'version' is bumped.
	AnnotationMeta = "@meta"
)
