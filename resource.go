package scim

type Resource[T any] struct {
	resourceType *ResourceType[T]
	root         *complexProperty
}

func (r *Resource[T]) ResourceType() *ResourceType[T] {
	return r.resourceType
}

func (r *Resource[T]) navigator() *navigator {
	return newNavigator(r.root)
}

func (r *Resource[T]) MarshalJSON() ([]byte, error) {
	return serialize(r, nil, nil)
}

func (r *Resource[T]) MarshalJSONWithAttributes(includes ...string) ([]byte, error) {
	return serialize(r, includes, nil)
}

func (r *Resource[T]) MarshalJSONWithExcludedAttributes(excludes ...string) ([]byte, error) {
	return serialize(r, nil, excludes)
}

func (r *Resource[T]) UnmarshalJSON(bytes []byte) error {
	return deserialize(bytes, r.navigator())
}
