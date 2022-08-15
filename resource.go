package scim

import "fmt"

type Resource[T any] struct {
	resourceType *ResourceType[T]
	root         *complexProperty
}

func (r *Resource[T]) ResourceType() *ResourceType[T] {
	return r.resourceType
}

func (r *Resource[T]) Add(path string, value any) error {
	if len(path) == 0 {
		return r.navigator().add(value).err
	}

	head, err := compilePath(path)
	if err != nil {
		return err
	}

	return defaultTraverse(
		r.root,
		r.resourceType.skipMainSchemaNamespace(head),
		func(nav *navigator) error { return nav.add(value).err },
	)
}

func (r *Resource[T]) Replace(path string, value any) error {
	if len(path) == 0 {
		return r.navigator().replace(value).err
	}

	head, err := compilePath(path)
	if err != nil {
		return err
	}

	return defaultTraverse(
		r.root,
		r.resourceType.skipMainSchemaNamespace(head),
		func(nav *navigator) error { return nav.replace(value).err },
	)
}

func (r *Resource[T]) Delete(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("%w: path must be specified for delete operation", ErrInvalidPath)
	}

	head, err := compilePath(path)
	if err != nil {
		return err
	}

	return defaultTraverse(
		r.root,
		r.resourceType.skipMainSchemaNamespace(head),
		func(nav *navigator) error { return nav.delete().err },
	)
}

func (r *Resource[T]) Evaluate(filter string) (bool, error) {
	return evaluate(r.root, filter)
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
