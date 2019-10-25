package query

import "github.com/imulab/go-scim/core"

// Compile the given SCIM path and return the head of the path's step chain. The path is allowed
// to contain intermediate filter.
func CompilePath(path string) (*core.Step, error) {
	compiler := &pathCompiler{
		scan: &pathScanner{},
		data: append(copyOf(path), 0, 0),
		off:  0,
		op:   scanPathContinue,
	}
	compiler.scan.init()

	head := &core.Step{}
	cursor := head

	for compiler.hasMore() {
		next, err := compiler.next()
		if err != nil {
			return nil, err
		}
		cursor.Next = next
		cursor = cursor.Next
	}

	return head.Next, nil
}

// Compile the given SCIM filter and return the root of the filter's tree. The path within the filter
// is not allowed to contain additional filters.
func CompileFilter(query string) (*core.Step, error) {
	return nil, nil
}

func copyOf(raw string) []byte {
	data := make([]byte, len(raw))
	copy(data, raw)
	return data
}

// Register recognized schema namespaces with this module, thus the compiler will recognize it as a
// separate path step.
func RegisterPathNamespace(ns string) {
	if len(ns) == 0 {
		return
	}
	namespaces = namespaces.insert(namespaces, ns, 0)
}