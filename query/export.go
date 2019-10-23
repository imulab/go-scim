package query

import "github.com/imulab/go-scim/core"

// Compile the given SCIM path and return the head of the path's step chain. The path is allowed
// to contain intermediate filter.
func CompilePath(path string) (*core.Step, error) {
	return nil, nil
}

// Compile the given SCIM filter and return the root of the filter's tree. The path within the filter
// is not allowed to contain additional filters.
func CompileFilter(query string) (*core.Step, error) {
	return nil, nil
}

// Register recognized schema namespaces with this module, thus the compiler will recognize it as a
// separate path step.
func RegisterPathNamespace(ns string) {
	if len(ns) == 0 {
		return
	}
	namespaces = namespaces.insert(namespaces, ns, 0)
}