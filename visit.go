package scim

// visitor can be implemented by parties that wish to traverse the Property tree.
type visitor interface {
	// shouldVisit returns boolean value to control whether the Property will be visited.
	shouldVisit(prop Property) bool
	// visit is called when shouldVisit returns true. The walkProperty process can be aborted
	// by returning an error here.
	visit(prop Property) error
	// beginChildren is called when walkProperty traversal is about to visit the sub properties
	// of a container property (i.e. multiValued, complex).
	beginChildren(container Property)
	// endChildren is called when walkProperty has ended traversal of sub properties of a container property.
	endChildren(container Property)
}

// visit does a DFS traversal on the given Property and invokes related methods on the visitor.
func visit(prop Property, visitor visitor) error {
	if !visitor.shouldVisit(prop) {
		return nil
	}

	if err := visitor.visit(prop); err != nil {
		return err
	}

	if prop.Attr().multiValued || prop.Attr().typ == TypeComplex {
		visitor.beginChildren(prop)

		if err := prop.ForEach(func(_ int, child Property) error {
			return visit(child, visitor)
		}); err != nil {
			return err
		}

		visitor.endChildren(prop)
	}

	return nil
}
