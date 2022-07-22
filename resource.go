package scim

type Resource[T any] struct {
	rtype *ResourceType[T]
	root  *complexProperty
}

func (r *Resource[T]) ResourceType() *ResourceType[T] {
	return r.rtype
}

func (r *Resource[T]) Clone() *Resource[T] {
	return &Resource[T]{
		rtype: r.rtype,
		root:  r.root.Clone().(*complexProperty),
	}
}

func (r *Resource[T]) Import(model *T) error {
	panic("not implemented")
}

func (r *Resource[T]) Export() (*T, error) {
	_ = r.rtype.newModel()
	panic("not implemented")
}
