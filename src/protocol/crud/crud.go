package crud

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/expr"
	"github.com/imulab/go-scim/src/core/prop"
)

// Add value to SCIM resource at the given SCIM path. If SCIM path is empty, value will be added
// to the root of the resource. The supplied value must be compatible with the target property attribute,
// otherwise error will be returned.
func Add(resource *prop.Resource, path string, value interface{}) error {
	if len(path) == 0 {
		return resource.Add(value)
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	return traverse(resource.NewNavigator(), skipResourceNamespace(resource, head), func(target core.Property) error {
		return target.Add(value)
	})
}

// Replace value in SCIM resource at the given SCIM path. If SCIM path is empty, the root of the resource
// will be replaced. The supplied value must be compatible with the target property attribute, otherwise
// error will be returned.
func Replace(resource *prop.Resource, path string, value interface{}) error {
	if len(path) == 0 {
		return resource.Replace(value)
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	return traverse(resource.NewNavigator(), skipResourceNamespace(resource, head), func(target core.Property) error {
		return target.Replace(value)
	})
}

// Delete value from the SCIM resource at the specified SCIM path. The path cannot be empty.
func Delete(resource *prop.Resource, path string) error {
	if len(path) == 0 {
		return errors.InvalidPath("path must not be empty when deleting from resource")
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	return traverse(resource.NewNavigator(), skipResourceNamespace(resource, head), func(target core.Property) error {
		return target.Delete()
	})
}
