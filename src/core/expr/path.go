package expr

// Create a new SCIM path expression, returns the head of the path expression linked list, or any error.
// The result may contain a filter root node, depending on the given path expression.
func CompilePath(path string) (*Expression, error) {
	return nil, nil
}

// Create a new path expression list, composed of nodes from the paths arguments. This function simply
// assembles the linked list and does not go through the compiler mechanism.
func NewPath(paths ...string) *Expression {
	if len(paths) == 0 {
		return nil
	}

	var (
		anchor = &Expression{}
		cursor = anchor
	)
	for len(paths) > 0 {
		cursor.next = &Expression{
			token: paths[0],
			typ:   exprPath,
		}
		cursor = cursor.next
		paths = paths[1:]
	}

	cursor = anchor.next
	anchor = nil

	return cursor
}
