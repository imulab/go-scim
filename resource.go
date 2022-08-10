package scim

type Resource[T any] struct {
	resourceType *ResourceType[T]
	root         *complexProperty
}

func (r *Resource[T]) ResourceType() *ResourceType[T] {
	return r.resourceType
}

func (r *Resource[T]) navigator() *navigator {
	return &navigator{stack: []Property{r.root}}
}

func (r *Resource[T]) UnmarshalJSON(bytes []byte) error {
	return deserialize(bytes, r.navigator())
}
