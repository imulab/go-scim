package scim

type Mapping[T any] struct {
	path       string
	getter     func(model *T) any
	setter     func(prop Property, model *T)
	filterable bool
}

func BuildMapping[T any](path string) *mappingDsl[T] {
	return &mappingDsl[T]{
		Mapping: &Mapping[T]{path: path},
	}
}

type mappingDsl[T any] struct {
	*Mapping[T]
}

func (d *mappingDsl[T]) EnableFilter() *mappingDsl[T] {
	d.filterable = true
	return d
}

func (d *mappingDsl[T]) Getter(getter func(model *T) any) *mappingDsl[T] {
	d.getter = getter
	return d
}

func (d *mappingDsl[T]) Setter(setter func(prop Property, model *T)) *mappingDsl[T] {
	d.setter = setter
	return d
}

func (d *mappingDsl[T]) Build() *Mapping[T] {
	switch {
	case d.getter == nil:
		panic("getter is required")
	case d.setter == nil:
		panic("setter is required")
	default:
		return d.Mapping
	}
}
