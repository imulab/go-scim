package annotation

const (
	Primary          = "@Primary"
	ExclusivePrimary = "@ExclusivePrimary"

	// Annotation to sync core schema attribute
	SyncSchema          = "@SyncSchema"
	StateSummary        = "@StateSummary"
	SchemaExtensionRoot = "@SchemaExtensionRoot"

	AutoCompact = "@AutoCompact"

	Identity     = "@Identity"
	CopyReadOnly = "@CopyReadOnly"

	// Annotation to contain other annotations intended for the derived element annotations
	ElementAnnotations = "@ElementAnnotations"
)
