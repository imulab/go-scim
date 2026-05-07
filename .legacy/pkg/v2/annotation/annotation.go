// This package describes the internally used annotations
package annotation

const (
	// @Primary annotates a boolean property which is considered a primary property.
	// Among the sub properties of a multiValued complex property, only one primary property
	// can hold the value true.
	Primary = "@Primary"
	// @ExclusivePrimary annotates a multiValued complex property, who wishes to have its
	// @Primary sub property regulated
	ExclusivePrimary = "@ExclusivePrimary"
	// @Root annotates the derived super attribute from a resource type. It is where all propagated events end
	Root = "@Root"
	// @SyncSchema load the SyncSchemaSubscriber to keep the schema property up to date in the assigned and unassigned
	// events of schema extension attributes. It must be annotated on the @Root attribute.
	SyncSchema = "@SyncSchema"
	// @StateSummary load the ComplexStateSummarySubscriber to summarize complex property state changes. It must be
	// annotated on a singular complex property.
	StateSummary = "@StateSummary"
	// @SchemaExtensionRoot annotates the derived complex attribute of a schema extension. It denotes the root of
	// the schema extension attribute. Together with @StateSummary, it provides information whether a schema extension
	// sub property is assigned or unassigned. @SyncSchema can use this information to update the schemas property.
	SchemaExtensionRoot = "@SchemaExtensionRoot"
	// @AutoCompact loads a AutoCompactSubscriber to automatically compact the multiValued property in case its element
	// becomes unassigned. It must be annotated on a multiValued property.
	AutoCompact = "@AutoCompact"
	// @Identity annotates sub attributes of a complex attribute which wishes to participate in equality comparison.
	// When one or more sub attributes are annotated with @Identity, they form the identity of the complex attribute.
	// If none of the sub attributes are annotated with @Identity, all sub attribute form the identity of the complex attribute.
	Identity = "@Identity"
	// @ElementAnnotations annotates additional annotations in its parameters which will be assigned as annotations for
	// the derived element attribute. This gives user explicit control as to what annotations will be loaded as a
	// multiValued property element.
	ElementAnnotations = "@ElementAnnotations"
	// @UUID annotates a string property. A uuid will be generated to replace the value of the annotated property
	// if the property is in the unassigned state when processed.
	UUID = "@UUID"
	// @BCrypt annotates a string property or a binary property`. The value of the property will be bCrypt hashed and
	// replace the original value. If the property is unassigned, no operation will be carried out. The annotation takes
	// a integer parameter named "cost". This will determine the strength of the bCrypt hashing. If omitted, default
	// cost is 10. The value replacement does not trigger event propagation, it is strictly local.
	BCrypt = "@BCrypt"
	// @ReadOnly annotates a readOnly property and indicates how filters should handle its value. Two options are
	// available. The first a boolean named "reset": if true, filters shall delete the property value; The second
	// is a boolean named "copy": if true, filters shall copy value from the reference property, if available.
	ReadOnly = "@ReadOnly"
	// @Enum annotates attributes that has canonicalValues and enforces its values to the range of the defined
	// canonicalValues. The defined values will be treated as strings and compared with respect to the caseExact
	// setting.
	Enum = "@Enum"
)
