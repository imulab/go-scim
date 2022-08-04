package scim

// Mapping defines the relation between a Resource structure and a backend model. It is parameterized to bind to the
// type of the backend model.
//
// Mapping is defined on a specific SCIM path that a model field is mapped to. Data binding is achieved by a getter
// and setter which is to be implemented when setting up the mapping.
//
// The mapping path can be filtered to allow it to appear in a SCIM filter expression. Non-filterable or unmapped
// paths may be rejected when appearing in a SCIM filter, as they do not map to a meaningful concrete data property.
type Mapping[T any] struct {
	path          string
	compiledPath  any
	getter        func(model *T) (any, error)
	setter        func(prop Property, model *T) error
	filterEnabled bool
}
