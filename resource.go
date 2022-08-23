package scim

import (
	"strings"
)

// Resource corresponds to a SCIM resource. It hosts the top level root complex Property. A Resource is bound to the
// type of the user defined model, and uses Import and Export methods to transfer values between them.
type Resource[T any] struct {
	resourceType *ResourceType[T]
	root         *complexProperty
}

// ResourceType returns the ResourceType that this Resource is built on. A Resource always has a non-nil ResourceType.
func (r *Resource[T]) ResourceType() *ResourceType[T] {
	return r.resourceType
}

// Import transfers the mapped values from the model into this Resource. This method would fail at the first mapping
// getter error. The transfer mode is "add", instead of "replace".
func (r *Resource[T]) Import(model *T) error {
	if model == nil {
		panic("expect non-nil model")
	}

	for _, mapping := range r.resourceType.mappings {
		value, err := mapping.getter(model)
		if err != nil {
			return err
		}

		err = r.addCompiled(mapping.compiledPath, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// Export transfers mapped values from this Resource into the model. This method would fail at the first error when
// retrieving the mapping value at given path and invoking the setter.
func (r *Resource[T]) Export(model *T) error {
	if model == nil {
		panic("expect non-nil model")
	}

	for _, mapping := range r.resourceType.mappings {
		prop, err := r.getProperty(mapping.compiledPath)
		if err != nil {
			return err
		}

		err = mapping.setter(prop, model)
		if err != nil {
			return err
		}
	}

	return nil
}

// ExportNew is a shorthand for Export with a new model.
func (r *Resource[T]) ExportNew() (*T, error) {
	model := r.resourceType.newModel()
	return model, r.Export(model)
}

// Patch applies a single PatchOperation on this Resource. If any error is returned, the Resource should be considered
// corrupted and invalid to export to the custom data model.
//
// This method is case-insensitive to PatchOperation.Op. For add operations, the value is always required, the path is
// optional (when absent, the value must be a json object, and the request is split into multiple add operations with
// paths of keys of the json object value); For replace operations, the value is always required, the path is optional
// (when absent, it is essentially replacing the whole of resource); For remove operations, the path is required and
// the value is ignored.
func (r *Resource[T]) Patch(request *PatchOperation) error {
	if err := request.validate(); err != nil {
		return err
	}

	switch strings.ToLower(request.Op) {
	case opAdd:
		if len(request.Path) == 0 {
			for k, v := range request.Value.(map[string]any) {
				if err := r.Patch(&PatchOperation{
					Op:    opAdd,
					Path:  k,
					Value: v,
				}); err != nil {
					return err
				}
			}
			return nil
		}
		return r.add(request.Path, request.Value)
	case opReplace:
		return r.replace(request.Path, request.Value)
	case opRemove:
		return r.delete(request.Path)
	default:
		panic("impossible case after validation")
	}
}

func (r *Resource[T]) getProperty(path *Expr) (Property, error) {
	nav := r.navigator()

	for cur := path; cur != nil; cur = cur.next {
		switch {
		case cur.IsPath():
			nav.dot(cur.value)
			if nav.hasError() {
				return nil, nav.err
			}
		case cur.IsFilterRoot():
			nav.where(func(child Property) bool {
				r, err := evaluateExpr(child, cur)
				return err != nil && r
			})
			if nav.hasError() {
				return nil, nav.err
			}
		default:
			return nil, ErrInvalidPath
		}
	}

	return nav.current(), nil
}

func (r *Resource[T]) add(path string, value any) error {
	if len(path) == 0 {
		return r.navigator().add(value).err
	}

	head, err := compilePath(path)
	if err != nil {
		return err
	}

	return r.addCompiled(head, value)
}

func (r *Resource[T]) addCompiled(path *Expr, value any) error {
	return defaultTraverse(
		r.root,
		r.resourceType.skipMainSchemaNamespace(path),
		func(nav *navigator) error { return nav.add(value).err },
	)
}

func (r *Resource[T]) replace(path string, value any) error {
	if len(path) == 0 {
		return r.navigator().add(value).err
	}

	head, err := compilePath(path)
	if err != nil {
		return err
	}

	return r.replaceCompiled(head, value)
}

func (r *Resource[T]) replaceCompiled(path *Expr, value any) error {
	return defaultTraverse(
		r.root,
		r.resourceType.skipMainSchemaNamespace(path),
		func(nav *navigator) error { return nav.replace(value).err },
	)
}

func (r *Resource[T]) delete(path string) error {
	if len(path) == 0 {
		panic("path is required")
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

func (r *Resource[T]) evaluate(filter string) (bool, error) {
	return evaluate(r.root, filter)
}

func (r *Resource[T]) navigator() *navigator {
	return newNavigator(r.root)
}

func (r *Resource[T]) MarshalJSON() ([]byte, error) {
	return serialize(r, nil, nil)
}

// MarshalJSONWithAttributes serializes this Resource to json format with the specified include-fields. With at least
// one include-fields, attributes whose "returned" is default will not be returned unless mentioned among the include-fields.
// Other "returned" option behaviors are unaffected.
//
// If the include-fields are empty, this method is the same as Resource.MarshalJSON which can be directly called by
// json.Marshal.
func (r *Resource[T]) MarshalJSONWithAttributes(includes ...string) ([]byte, error) {
	return serialize(r, includes, nil)
}

// MarshalJSONWithExcludedAttributes serializes this Resource to json format with the specified exclude-fields. With at
// least one exclude-fields, attributes whose "returned" is default will be returned unless mentioned among the exclude-fields.
// Other "returned" option behaviors are unaffected.
//
// If the exclude-fields are empty, this method is the same as Resource.MarshalJSON which can be directly called by
// json.Marshal.
func (r *Resource[T]) MarshalJSONWithExcludedAttributes(excludes ...string) ([]byte, error) {
	return serialize(r, nil, excludes)
}

func (r *Resource[T]) UnmarshalJSON(bytes []byte) error {
	return deserialize(bytes, r.navigator())
}
