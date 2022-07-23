package scim

// visitor can be implemented by parties that wish to react to walkProperty events.
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

// walkProperty does a DFS traversal on the given Property and invokes related methods on the visitor.
func walkProperty(prop Property, visitor visitor) error {
	if !visitor.shouldVisit(prop) {
		return nil
	}

	if err := visitor.visit(prop); err != nil {
		return err
	}

	if prop.Attribute().multiValued || prop.Attribute().typ == TypeComplex {
		visitor.beginChildren(prop)

		if err := prop.Iterate(func(_ int, child Property) error {
			return walkProperty(child, visitor)
		}); err != nil {
			return err
		}

		visitor.endChildren(prop)
	}

	return nil
}
